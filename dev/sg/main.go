package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/peterbourgon/ff/v3/ffcli"
)

var (
	runFlagSet = flag.NewFlagSet("sg run", flag.ExitOnError)

	runCommand = &ffcli.Command{
		Name:       "run",
		ShortUsage: "sg run <command>",
		ShortHelp:  "Run the given command.",
		FlagSet:    runFlagSet,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) != 1 {
				return errors.New("whoops, only one command")
			}

			cmd, ok := conf.Commands[args[0]]
			if !ok {
				return fmt.Errorf("command %q not found", args[0])
			}

			return run(ctx, cmd)
		},
	}
)

var (
	runSetFlagSet = flag.NewFlagSet("sg run", flag.ExitOnError)

	runSetCommand = &ffcli.Command{
		Name:       "run-set",
		ShortUsage: "sg run-set <commandset>",
		ShortHelp:  "Run the given command set.",
		FlagSet:    runSetFlagSet,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) != 1 {
				return errors.New("whoops, only one commandset")
			}

			names, ok := conf.Commandsets[args[0]]
			if !ok {
				return fmt.Errorf("commandset %q not found", args[0])
			}

			cmds := make([]Command, 0, len(names))
			for _, name := range names {
				cmd, ok := conf.Commands[name]
				if !ok {
					return fmt.Errorf("command %q not found", name)
				}

				cmds = append(cmds, cmd)
			}

			return run(ctx, cmds...)
		},
	}
)

var (
	rootFlagSet = flag.NewFlagSet("sg", flag.ExitOnError)
	configFlag  = rootFlagSet.String("config", "sg.config.yaml", "configuration file")
	conf        *Config

	rootCommand = &ffcli.Command{
		ShortUsage:  "sg [flags] <subcommand>",
		FlagSet:     rootFlagSet,
		Subcommands: []*ffcli.Command{runCommand, runSetCommand},
	}
)

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
