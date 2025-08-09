package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
	}))

	rootCmd := &cobra.Command{
		Use:   "website-generator",
		Short: "website-generator [flags] <subcommand> [flags] [<args>...]",
	}

	rootCmd.AddCommand(newGenerateCmd(logger))

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
