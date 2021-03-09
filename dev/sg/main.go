package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

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

		// Build it
		c := exec.CommandContext(ctx, "bash", "-c", cmd.Install)
		c.Dir = root
		c.Env = os.Environ()
		for k, v := range conf.Env {
			c.Env = append(c.Env, fmt.Sprintf("%s=%s", k, v))
		}
		out, err := c.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to install %q: %s (output: %s)", args[0], err, out)
		}

		// Run it
		c = exec.CommandContext(ctx, "bash", "-c", cmd.Cmd)
		c.Dir = root
		c.Env = os.Environ()
		for k, v := range conf.Env {
			c.Env = append(c.Env, fmt.Sprintf("%s=%s", k, v))
		}
		out, err = c.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to install %q: %s. output:\n%s", args[0], err, out)
		}

		return nil
	},
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
