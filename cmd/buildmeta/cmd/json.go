package cmd

import (
	"encoding/json"
	"log"
	"os"

	"github.com/andrewstuart/buildmeta"
	"github.com/spf13/cobra"
)

// jsonCmd represents the json command
var jsonCmd = &cobra.Command{
	Use:   "json [path]",
	Short: "Exports JSON representation of buildmeta for the working directory",
	Run: func(cmd *cobra.Command, args []string) {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		i, err := buildmeta.GenerateInfo(path)
		if err != nil {
			log.Fatal(err)
		}
		bs, err := json.MarshalIndent(i, "", "  ")
		if err != nil {
			log.Fatal("JSON marshal error ", err)
		}
		_, err = os.Stdout.Write(bs)
		if err != nil {
			log.Fatal("Stdout write error ", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(jsonCmd)
}
