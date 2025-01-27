package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/peterbourgon/ff/v3/ffcli"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
	}))

	//

	var (
		rootCmd = &ffcli.Command{
			Name:       "website-generator",
			ShortUsage: "website-generator [flags] <subcommand> [flags] [<args>...]",
			Exec: func(ctx context.Context, args []string) error {
				return flag.ErrHelp
			},
		}
		generateCmd = newGenerateCmd(logger)
	)

	rootCmd.Subcommands = []*ffcli.Command{
		generateCmd,
	}

	err := rootCmd.ParseAndRun(context.Background(), os.Args[1:])
	switch {
	case errors.Is(err, flag.ErrHelp):
		os.Exit(1)
	case err != nil:
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
