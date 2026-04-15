package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"music-utils/common"
)

var configFileOrDir string

var repairCmd = &cobra.Command{
	Use:   "repair",
	Short: "Repairs broken links in one or more playlists",
	Long: `
Repairs broklen links in one or more playlists using
input from one or more reports that have moved file results.

To repair one playlist using one report:

music-utils playlist repair --input-file $HOME/Music/playlist.m3u -c $HOME/Music/report.json

To repair one playlist using all reports found in one directory (non-recursive):

music-utils playlist repair --input-file $HOME/Music/playlist.m3u -c $HOME/Music/reports

To repair all playlist in the input directory using one report:

music-utils playlist repair -i $HOME/Music -c $HOME/Music/report.json

To repair all playlist recursively in the input directory and all subs using one report:

music-utils playlist repair -r -i $HOME/Music -c $HOME/Music/report.json
`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		// Verify that input dir exists (also expands the path)
		inputDir, err = common.FlagDirectoryExists(inputDir)
		if err != nil {
			return err
		}
		// Verify the config file or directory exists (also expands the path)
		configFileOrDir, err = common.FlagDirectoryExists(configFileOrDir)
		if err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("repair called")
		return nil
	},
}

func init() {
	playlistCmd.AddCommand(repairCmd)

	repairCmd.Flags().StringVarP(&configFileOrDir, "config-path", "c", "", "Config file or directory containing multiple")
}
