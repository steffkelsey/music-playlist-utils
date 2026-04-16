package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"music-utils/common"
)

var configFileOrDir string
var sourceToDestMap map[string]string

type repairPlaylistsReport struct {
	Config   []string                 `json:"config"`
	Skipped  []string                 `json:"skipped"`
	Repaired []string                 `json:"repaired"`
	Failed   []validatePlaylistResult `json:"failed"`
}

var repairCmd = &cobra.Command{
	Use:   "repair",
	Short: "Repairs broken links in one or more playlists",
	Long: `
Repairs broken links in one or more playlists using
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
		results := repairPlaylistsReport{
			Config:   make([]string, 0),
			Skipped:  make([]string, 0),
			Repaired: make([]string, 0),
			Failed:   make([]validatePlaylistResult, 0),
		}
		maybeValidReports, err := getReportsToValidate(configFileOrDir)
		if err != nil {
			return err
		}

		sourceToDestMap = make(map[string]string)
		for _, r := range maybeValidReports {
			t, o := isValidOrganizedReport(r)
			if t {
				results.Config = append(results.Config, r)
				// parse the report into source to dest map
				// where source and dest are absolute paths of where music files moved
				for _, m := range o.Moved {
					sourceToDestMap[m.Source] = m.Dest
				}
			}
		}

		// get playlists to validate
		v, err := getPlaylistsToValidate()
		if err != nil {
			return err
		}
		validateResults, err := validateAllPlaylists(v)
		if err != nil {
			return err
		}

		// any valid have Skipped repair
		results.Skipped = append(results.Skipped, validateResults.Valid...)

		// repair each playlists that fail validation
		for i, invalid := range slices.Backward(validateResults.Invalid) {
			// if it is the dry-run, just check against the sourceToDest map for each
			// bad path
			if isDryRun {
				// Simulate repair by going through the bad paths in the validation result
				for a, b := range slices.Backward(invalid.BadPaths) {
					_, ok := sourceToDestMap[b]
					if ok {
						// remove from bad paths slice
						invalid.BadPaths = append(invalid.BadPaths[:a], invalid.BadPaths[a+1:]...)
					}
				}
				if len(invalid.BadPaths) == 0 {
					// remove the invalid report
					validateResults.Invalid = append(validateResults.Invalid[:i], validateResults.Invalid[i+1:]...)
					// add to the path to the repaired slice
					results.Repaired = append(results.Repaired, invalid.Path)
				}
				continue
			}

			// Use getPlaylistPaths with a repair process func to get new music file paths
			newPaths, err := getPlaylistPaths(invalid.Path, getPathWhenRepairingPlaylist, filepath.Dir(invalid.Path))
			if err != nil {
				fmt.Printf("error repairing path. %v\n", err)
				continue
			}
			newPathsStr := strings.Join(newPaths, "\n")
			// Write into the file
			f, err := os.Create(invalid.Path)
			if err != nil {
				fmt.Printf("error creating file. %s\n", err)
				continue
			}

			// close the file when done
			defer f.Close()

			// write to the file
			_, err = f.WriteString(newPathsStr)
			if err != nil {
				fmt.Printf("error writing file. %s\n", err)
				continue
			}
		}

		if isDryRun {
			// make sure the ones that didn't get fixed are reported as Failed
			results.Failed = append(results.Failed, validateResults.Invalid...)
			// print the json repair report
			j, _ := json.Marshal(&results)
			fmt.Println(string(j))
			return nil
		}

		// validate all repaired playlists
		v = append(make([]validatePlaylistResult, 0), validateResults.Invalid...)
		validateResults, err = validateAllPlaylists(v)
		if err != nil {
			return err
		}

		results.Repaired = append(results.Repaired, validateResults.Valid...)
		results.Failed = append(results.Failed, validateResults.Invalid...)

		return nil
	},
}

func init() {
	playlistCmd.AddCommand(repairCmd)

	repairCmd.Flags().StringVarP(&configFileOrDir, "config-path", "c", "", "Config file or directory containing multiple")
}

func getReportsToValidate(reportPath string) ([]string, error) {
	paths := make([]string, 0)

	// Is the inputDir a directory or a file?
	info, err := os.Stat(reportPath)
	if err != nil {
		return paths, err
	}
	if info.IsDir() {
		r, err := getAllJsonReportsInFolder(reportPath)
		if err != nil {
			return paths, err
		}
		paths = append(paths, r...)
	} else {
		if common.IsJsonFile(reportPath) {
			paths = append(paths, reportPath)
		}
	}

	return paths, nil
}

func getAllJsonReportsInFolder(rootPath string) ([]string, error) {
	results := make([]string, 0)
	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return results, err
	}

	for _, entry := range entries {
		// Skip subdirectories
		if entry.IsDir() {
			continue
		}

		path := filepath.Join(rootPath, entry.Name())
		// only act on json files by extension
		if common.IsJsonFile(path) {
			results = append(results, path)
		}
	}
	return results, nil
}

func isValidOrganizedReport(path string) (bool, organizedReport) {
	var o organizedReport
	// open and read the whole file (json is usually tiny)
	data, err := os.ReadFile(path)
	if err != nil {
		return false, o
	}

	// attempt to unmarshal into a organizedReport struct
	err = json.Unmarshal(data, &o)
	if err != nil {
		return false, o
	}
	return true, o
}

// getPathWhenRepairingPlaylist is a process function for getting the path of a music file
// when the playlist is being repaired while the music files remain in place
func getPathWhenRepairingPlaylist(path string, sourcePlDir string, destPlDir string) string {
	if !filepath.IsAbs(path) {
		path = common.CreateAbsPath(path, sourcePlDir)
	}
	// check if the path is good
	if !common.FileExists(path) {
		// see if we have a destination for this one
		goodPath, ok := sourceToDestMap[path]
		if ok {
			// Use it!
			path = goodPath
		}
	}

	// Convert to relative path
	return common.MoveRelativePath(path, sourcePlDir, destPlDir)
}
