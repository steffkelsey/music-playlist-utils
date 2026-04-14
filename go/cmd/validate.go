package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"music-utils/common"
)

type validatePlaylistResult struct {
	IsValid  bool     `json:"-"`
	Path     string   `json:"path"`
	Reason   string   `json:"reason"`
	BadPaths []string `json:"badPaths,omitempty"`
}

type validatePlaylistsReport struct {
	Valid   []string                 `json:"valid"`
	Invalid []validatePlaylistResult `json:"invalid"`
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "A command for validating one or more playlists",
	Long: `This command outputs JSON and validates one or more playlists. 
A valid playlist is an m3u file that only contains links to music files.
The validator ensures that there is a file existing for each link.

To validate one playlist:

music-utils playlist validate --input-file "$HOME/Music/folder/list.m3u"

To validate all playlists found in the inputDir (but no deeper):

music-utils playlist validate -i "$HOME/Music"

To validate all playlists recursively found in the inputDir and all sub-directories:

music-utils playlist validate -i -r "$HOME/Music"`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		// Verify that input dir exists
		inputDir, err = common.FlagDirectoryExists(inputDir)
		if err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := getPlaylistsToValidate()
		if err != nil {
			return err
		}
		r, err := validateAllPlaylists(v)
		if err != nil {
			return err
		}
		// print the json report
		j, _ := json.Marshal(&r)
		fmt.Println(string(j))
		return nil
	},
}

func init() {
	playlistCmd.AddCommand(validateCmd)
}

func getPlaylistsToValidate() ([]validatePlaylistResult, error) {
	toValidate := make([]validatePlaylistResult, 0)
	if len(inputFile) > 0 {
		// expand to deal with possible env variables like $HOME
		inputFile = os.ExpandEnv(inputFile)
		// verify input file exists
		if !common.FileExists(inputFile) {
			return toValidate, fmt.Errorf("playlist does not exist at %s", inputFile)
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
				return toValidate, err
			}
		} else {
			// walk just the inputDir finding all playlist files
			res, err = common.WalkAllMusicFilesNotRecursive(inputDir, findPlaylistFile)
			if err != nil {
				return toValidate, err
			}
		}
		// add any paths found to the slice as validatePlaylistResult
		for _, p := range res.Files {
			toValidate = append(toValidate, validatePlaylistResult{
				IsValid:  true,
				Path:     p,
				Reason:   "",
				BadPaths: make([]string, 0),
			})
		}
	}
	return toValidate, nil
}

func validateAllPlaylists(toValidate []validatePlaylistResult) (validatePlaylistsReport, error) {
	r := validatePlaylistsReport{
		Valid:   make([]string, 0),
		Invalid: make([]validatePlaylistResult, 0),
	}
	for _, v := range toValidate {
		err := isPlaylistValid(&v, filepath.Dir(v.Path))
		if err != nil {
			return r, err
		}
		if v.IsValid {
			r.Valid = append(r.Valid, v.Path)
		} else {
			r.Invalid = append(r.Invalid, v)
		}
	}
	return r, nil
}

func isPlaylistValid(r *validatePlaylistResult, listDir string) error {
	file, err := os.Open(r.Path)
	if err != nil {
		r.Reason = "Could not open file"
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		maybeValidMusicFilePath := scanner.Text()
		if !filepath.IsAbs(maybeValidMusicFilePath) {
			maybeValidMusicFilePath = common.CreateAbsPath(maybeValidMusicFilePath, listDir)
		}
		// for all path type, check that it exists
		if !common.FileExists(maybeValidMusicFilePath) {
			r.BadPaths = append(r.BadPaths, maybeValidMusicFilePath)
		}
	}

	if err := scanner.Err(); err != nil {
		r.Reason = "Could not scan the lines"
		return err
	}

	if len(r.BadPaths) > 0 {
		r.Reason = "One or more bad paths"
	}
	r.IsValid = len(r.BadPaths) == 0
	return nil
}
