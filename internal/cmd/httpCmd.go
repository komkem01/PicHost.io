package cmd

import (
	"fmt"

	"mcop/internal/http"

	"github.com/spf13/cobra"
)

// HTTP is serve http ot https
func HTTP(isHTTPS bool) *cobra.Command {
	name := "http"
	if isHTTPS {
		name = "https"
	}
	cmd := &cobra.Command{
		Use:   name,
		Short: fmt.Sprintf("Run server on %s protocal", name),
		Run:   http.D(isHTTPS),
	}
	return cmd
}
