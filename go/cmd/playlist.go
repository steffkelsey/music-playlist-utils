package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"music-utils/common"
)

var inputFile string

var playlistCmd = &cobra.Command{
	Use:   "playlist",
	Short: "Command for moving playlists and/or validating them",
	Long: `Command for moving one or more playlists. Also for 
validating them. To move one playlist:

music-utils playlist --input-file "$HOME/Music/folder/list.m3u" -o "$HOME/Music/playlists"

To move all playlists recursively found in the inputDir:

music-utils playlist -i "$HOME/Music" -o "$HOME/Music/playlists"`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		// Verify that input dir exists
		inputDir, err = common.FlagDirectoryExists(inputDir)
		if err != nil {
			return err
		}
		// Verify the output directory exists
		outputDir, err = common.FlagDirectoryExists(outputDir)
		if err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(inputFile) > 0 {
			fmt.Printf("playlist called, input-file: %s\n", inputFile)
		} else {
			fmt.Printf("playlist called. Walking %s for playlist files...\n", inputDir)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(playlistCmd)

	playlistCmd.Flags().StringVar(&inputFile, "input-file", "", "Single playlist file to move")
}
