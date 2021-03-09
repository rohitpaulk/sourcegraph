package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/dev/sg/root"
)

var (
	rootFlagSet = flag.NewFlagSet("sg", flag.ExitOnError)
	configFlag  = rootFlagSet.String("config", "sg.config.yaml", "configuration file")

	watchFlagSet = flag.NewFlagSet("sg watch", flag.ExitOnError)

	runFlagSet = flag.NewFlagSet("sg run", flag.ExitOnError)
)

var watchCommand = &ffcli.Command{
	Name:       "watch",
	ShortUsage: "sg watch <arg>",
	ShortHelp:  "Watch changes to the repository.",
	FlagSet:    watchFlagSet,
	Exec: func(ctx context.Context, args []string) error {
		paths, err := watch()
		if err != nil {
			return err
		}

		for path := range paths {
			fmt.Printf("SOMETHING CHANGED: %v\n", path)
		}

		return nil
	},
}

var runCommand = &ffcli.Command{
	Name:       "run",
	ShortUsage: "sg run <command>",
	ShortHelp:  "Watch changes to the repository.",
	FlagSet:    watchFlagSet,
	Exec: func(ctx context.Context, args []string) error {
		if len(args) != 1 {
			return errors.New("whoops, only one command")
		}

		cmd, ok := conf.Commands[args[0]]
		if !ok {
			return fmt.Errorf("command %q not found", args[0])
		}

		root, err := root.RepositoryRoot()
		if err != nil {
			return err
		}

		pathChanges, err := watch()
		if err != nil {
			return err
		}

		for {
			// Build it
			fmt.Printf("Installing...\n")

			c := exec.CommandContext(ctx, "bash", "-c", cmd.Install)
			c.Dir = root
			c.Env = makeEnv(conf.Env, cmd.Env)
			out, err := c.CombinedOutput()
			if err != nil {
				return fmt.Errorf("failed to install %q: %s (output: %s)", args[0], err, out)
			}

			//
			// Run it
			fmt.Printf("Running...\n")

			commandCtx, cancel := context.WithCancel(ctx)
			c = exec.CommandContext(commandCtx, "bash", "-c", cmd.Cmd)
			c.Dir = root
			c.Env = makeEnv(conf.Env, cmd.Env)
			stdout, err := c.StdoutPipe()
			if err != nil {
				return err
			}

			stderr, err := c.StderrPipe()
			if err != nil {
				return err
			}

			wg := &sync.WaitGroup{}

			readIntoBuf := func(prefix string, r io.Reader) {
				defer wg.Done()

				scanner := bufio.NewScanner(r)
				for scanner.Scan() {
					fmt.Fprintf(os.Stdout, "%s: %s\n", prefix, scanner.Text())
				}
			}

			wg.Add(2)
			go readIntoBuf("stdout", stdout)
			go readIntoBuf("stderr", stderr)

			if err := c.Start(); err != nil {
				return err
			}

			errs := make(chan error, 1)
			go func() {
				defer close(errs)

				errs <- (func() error {
					wg.Wait()

					if err := c.Wait(); err != nil {
						if exitErr, ok := err.(*exec.ExitError); ok {
							return fmt.Errorf("exited with %d", exitErr.ExitCode())
						}

						return err
					}

					return nil
				})()
			}()

		outer:
			for {
				select {
				case path := <-pathChanges:
					found := false
					for _, prefix := range cmd.Watch {
						if strings.HasPrefix(path, prefix) {
							found = true
						}
					}
					if !found {
						// Not a path we care about
						continue outer
					}

					fmt.Printf("Path changed, restarting: %v\n", path)

					cancel()    // Stop command
					<-errs      // Wait for exit
					break outer // Reinstall

				case err := <-errs:
					// Exited on its own or errored
					return err
				}
			}
		}

		return nil
	},
}

func makeEnv(envs ...map[string]string) []string {
	combined := os.Environ()
	for _, env := range envs {
		for k, v := range env {
			// TODO - should expand env variables that reference others?
			// SRC_REPOS_DIR: $HOME/.sourcegraph/repos is a literal value today
			combined = append(combined, fmt.Sprintf("%s=%s", k, v))
		}
	}

	return combined
}

var rootCommand = &ffcli.Command{
	ShortUsage:  "sg [flags] <subcommand>",
	FlagSet:     rootFlagSet,
	Subcommands: []*ffcli.Command{watchCommand, runCommand},
}

var conf *Config

func main() {
	if err := rootCommand.Parse(os.Args[1:]); err != nil {
		log.Fatal(err)
	}

	var err error
	conf, err = ParseConfigFile(*configFlag)
	if err != nil {
		log.Fatal(err)
	}

	if err := rootCommand.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
