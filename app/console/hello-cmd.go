// Package console implements CLI commands for the application
package console

import "github.com/spf13/cobra"

func helloCMD() *cobra.Command {
	return &cobra.Command{
		Use:   "hello",
		Short: "Prints hello world",
		Run: func(cmd *cobra.Command, _ []string) {
			cmd.Println("Hello, world!")
		},
	}
}
