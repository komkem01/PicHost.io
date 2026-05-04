package main

import (
	"log/slog"
	"os"

	"mcop/app/console"
	"mcop/internal/cmd"

	"github.com/spf13/cobra"
)

func main() {
	cobra.EnableCommandSorting = false
	if err := exec(); err != nil {
		slog.Error("Error running")
		os.Exit(1)
	}
}

func command() error {
	cmda := &cobra.Command{
		Use:  "app",
		Args: cmd.NotReqArgs,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}

	cmds := &cobra.Command{
		Use:   "cmd",
		Short: "List all commands",
	}
	cmds.AddCommand(console.Commands()...)

	cmda.AddCommand(cmd.HTTP(false), cmd.HTTP(true))
	cmda.AddCommand(cmd.Migrate())
	cmda.AddCommand(cmds)

	return cmda.Execute()
}
