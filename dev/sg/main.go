package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/peterbourgon/ff/v3/ffcli"
)

var (
	rootFlagSet  = flag.NewFlagSet("sg", flag.ExitOnError)
	watchFlagSet = flag.NewFlagSet("sg watch", flag.ExitOnError)
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

var rootCommand = &ffcli.Command{
	ShortUsage:  "sg [flags] <subcommand>",
	FlagSet:     rootFlagSet,
	Subcommands: []*ffcli.Command{watchCommand},
}

func main() {
	if err := rootCommand.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}
