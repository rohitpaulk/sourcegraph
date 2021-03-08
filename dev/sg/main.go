package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/peterbourgon/ff/v3/ffcli"
)

var (
	rootFlagSet   = flag.NewFlagSet("textctl", flag.ExitOnError)
	verbose       = rootFlagSet.Bool("v", false, "increase log verbosity")
	repeatFlagSet = flag.NewFlagSet("textctl repeat", flag.ExitOnError)
	n             = repeatFlagSet.Int("n", 3, "how many times to repeat")
)

var repeat = &ffcli.Command{
	Name:       "repeat",
	ShortUsage: "textctl repeat [-n times] <arg>",
	ShortHelp:  "Repeatedly print the argument to stdout.",
	FlagSet:    repeatFlagSet,
	Exec: func(_ context.Context, args []string) error {
		// foo
		return nil
	},
}

var count = &ffcli.Command{
	Name:       "count",
	ShortUsage: "textctl count [<arg> ...]",
	ShortHelp:  "Count the number of bytes in the arguments.",
	Exec: func(_ context.Context, args []string) error {
		// foo
		return nil
	},
}

var root = &ffcli.Command{
	ShortUsage:  "textctl [flags] <subcommand>",
	FlagSet:     rootFlagSet,
	Subcommands: []*ffcli.Command{repeat, count},
}

func main() {
	if err := root.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}
