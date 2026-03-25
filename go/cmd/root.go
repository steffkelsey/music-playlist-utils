package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"music-utils/common"
)

var inputDir string
var outputDir string

var rootCmd = &cobra.Command{
	Use:   "music-utils",
	Short: "Utilities for managing music libraries",
	Long: `music-utils helps you deal with common problems when maintaining
music libraries. Remove duplicates, find untagged files, rename and move files
while preserving m3u playlists.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&inputDir, common.ParamInputDir, "i", "$HOME/Music", "input directory")
	rootCmd.PersistentFlags().StringVarP(&outputDir, common.ParamOutputDir, "o", "$HOME/music-utils-out", "output directory")
}

