package cmd

import (
	"fmt"
	"log"

	"github.com/andrewstuart/buildmeta"
	"github.com/spf13/cobra"
)

// ldflagsCmd represents the ldflags command
var ldflagsCmd = &cobra.Command{
	Use:   "ldflags [repoPath]",
	Short: "Prints to stdout the values that should be used for go build -ldflags",
	Run: func(cmd *cobra.Command, args []string) {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		i, err := buildmeta.GenerateInfo(path)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print(i.LDFlags())
	},
}

func init() {
	rootCmd.AddCommand(ldflagsCmd)
}
