package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/peterbourgon/ff/v3/ffcli"
	"go.uber.org/zap"
)

func main() {
	zapConfig := zap.NewDevelopmentConfig()
	logger, err := zapConfig.Build()
	if err != nil {
		log.Fatal(err)
	}
	zap.ReplaceGlobals(logger)

	//

	var (
		rootCmd = &ffcli.Command{
			Name:       "website-generator",
			ShortUsage: "website-generator [flags] <subcommand> [flags] [<args>...]",
			Exec: func(ctx context.Context, args []string) error {
				return flag.ErrHelp
			},
		}
		generateCmd = newGenerateCmd()
	)

	rootCmd.Subcommands = []*ffcli.Command{
		generateCmd,
	}

	err = rootCmd.ParseAndRun(context.Background(), os.Args[1:])
	switch {
	case errors.Is(err, flag.ErrHelp):
		os.Exit(1)
	case err != nil:
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
