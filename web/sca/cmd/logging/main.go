package main

import "github.com/spf13/cobra"

var (
	RootCmd = &cobra.Command{}
)

func main() {
	if err := RootCmd.Execute(); err != nil {
		RootCmd.Println(err)
	}
}
