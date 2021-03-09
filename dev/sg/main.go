package main

import (
	"context"
	"flag"
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
	Exec:       watch,
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
