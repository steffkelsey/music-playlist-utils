package cmd

import (
	"github.com/spf13/cobra"
)

var data string
var decodedData []byte

var compareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Commands for comparing music data objects (Tracks, Albums)",
	Long:  `Command for comparing music data objects (Tracks, Albums, etc)`,
}

func init() {
	rootCmd.AddCommand(compareCmd)
	compareCmd.PersistentFlags().StringVar(&data, "data", "", "Base64 encoded data string")
}
