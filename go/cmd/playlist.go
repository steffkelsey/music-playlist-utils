package cmd

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"

	"github.com/spf13/cobra"

	"music-utils/common"
)

var inputFile string
var isRecursive bool
var validateOnly bool

type validatePlaylistResult struct {
	IsValid  bool     `json:"-"`
	Path     string   `json:"path"`
	Reason   string   `json:"reason"`
	BadPaths []string `json:"badPaths"`
}

type validatePlaylistsReport struct {
	Valid   []string                 `json:"valid"`
	Invalid []validatePlaylistResult `json:"invalid"`
}

type movePlaylistResult struct {
	Paths   []string                 `json:"-"`
	Moved   []string                 `json:"moved"`
	Skipped []validatePlaylistResult `json:"skipped"`
}

var playlistCmd = &cobra.Command{
	Use:   "playlist",
	Short: "Command for moving playlists and/or validating them",
	Long: `Command for moving one or more playlists. Also for 
validating them. To move one playlist:

music-utils playlist --input-file "$HOME/Music/folder/list.m3u" -o "$HOME/Music/playlists"

To validate one playlist:

music-utils playlist --input-file "$HOME/Music/folder/list.m3u" -v"

To move all playlists found in the inputDir:

music-utils playlist -i "$HOME/Music" -o "$HOME/Music/playlists

To move all playlists recursively found in the inputDir and all sub-directories:

music-utils playlist -i "$HOME/Music" -r -o "$HOME/Music/playlists

To validate all playlists recursively found in the inputDir and all sub-directories:

music-utils playlist -i "$HOME/Music" -r -v"`,
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
		toValidate := make([]validatePlaylistResult, 0)
		toMove := movePlaylistResult{
			Paths:   make([]string, 0),
			Moved:   make([]string, 0),
			Skipped: make([]validatePlaylistResult, 0),
		}

		if len(inputFile) > 0 {
			// expand to deal with possible env variables like $HOME
			inputFile = os.ExpandEnv(inputFile)
			// verify input file exists
			if !common.FileExists(inputFile) {
				return fmt.Errorf("playlist does not exist at %s", inputFile)
			}
			// add the path to toValidate slice
			toValidate = append(toValidate, validatePlaylistResult{
				IsValid:  true,
				Path:     inputFile,
				Reason:   "",
				BadPaths: make([]string, 0),
			})

		} else {
			var res common.WalkResults
			var err error
			if isRecursive {
				// walk recursively finding all playlist files
				res, err = common.WalkAllMusicFiles(inputDir, findPlaylistFile)
				if err != nil {
					return err
				}
			} else {
				// walk just the inputDir finding all playlist files
				res, err = common.WalkAllMusicFilesNotRecursive(inputDir, findPlaylistFile)
				if err != nil {
					return err
				}
			}
			for _, p := range res.Files {
				toValidate = append(toValidate, validatePlaylistResult{
					IsValid:  true,
					Path:     p,
					Reason:   "",
					BadPaths: make([]string, 0),
				})
			}
		}

		validateResult := validatePlaylistsReport{
			Valid:   make([]string, 0),
			Invalid: make([]validatePlaylistResult, 0),
		}
		for _, v := range toValidate {
			// TODO open the playlist file
			// TODO read each line in the file
			// TODO for each line, check if relative path or abs
			// TODO if relative path, make it abs
			// TODO for all path type, check that it exists
			if v.IsValid {
				validateResult.Valid = append(validateResult.Valid, v.Path)
			} else {
				validateResult.Invalid = append(validateResult.Invalid, v)
			}
		}

		if validateOnly {
			// print the json report
			j, _ := json.Marshal(&validateResult)
			fmt.Println(string(j))
			return nil
		}

		for _, p := range toMove.Paths {
			fmt.Printf("Moving %s...\n", p)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(playlistCmd)

	playlistCmd.Flags().StringVar(&inputFile, "input-file", "", "Single playlist file to move")

	playlistCmd.PersistentFlags().BoolVarP(&isRecursive, common.ParamRecursive, "r", false, "Searches all sub-directories")
	playlistCmd.PersistentFlags().BoolVarP(&validateOnly, common.ParamValidate, "v", false, "Validate playlist file(s) only. No changes are made")
}

func findPlaylistFile(path string, info fs.FileInfo, results *common.WalkResults) error {
	// Ignore everything but playlist files
	if !common.IsPlaylistFile(path) {
		return nil
	}
	results.Files = append(results.Files, path)
	results.Count++
	return nil
}
