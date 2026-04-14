package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"music-utils/common"
)

type exportedPlaylist struct {
	Path       string   `json:"path"`
	MusicFiles []string `json:"musicFiles"`
}

type exportPlaylistsResult struct {
	Paths    []string                 `json:"-"`
	Exported []exportedPlaylist       `json:"exported"`
	Skipped  []validatePlaylistResult `json:"skipped"`
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export one or more playlists and all linked music files to a directory",
	Long: `Export one or more playlists and all linked music files to a directory.
Copies are made, the original playlist and music files stay in place.
The new playlist abnd music files are all placed in the output directory.
Subfolders are not preserved.
To export one playlist and all linked music files:

music-utils playlist export --input-file "$HOME/Music/playlist 1.m3u" -o $HOME/Music/playlist 1"

To export all playlists in the input directory and all linked music files:

music-utils playlist export -i "$HOME/Music" -o $HOME/Music/all-playlists"

To export all playlists in the input directory and recursively in all sub-folders:

music-utils playlist export -r -i "$HOME/Music" -o $HOME/Music/totally-all-playlists"
`,
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
		// get and validate playlists
		v, err := getPlaylistsToValidate()
		if err != nil {
			return err
		}
		r, err := validateAllPlaylists(v)
		if err != nil {
			return err
		}
		exportedReport, err := exportPlaylists(r)
		if err != nil {
			return err
		}

		// if dryrun, print the report
		if isDryRun {
			j, _ := json.Marshal(&exportedReport)
			fmt.Println(string(j))
			return nil
		}

		// validate the playlists that we exported
		for _, e := range exportedReport.Exported {
			v := validatePlaylistResult{
				IsValid:  true,
				Path:     e.Path,
				Reason:   "",
				BadPaths: make([]string, 0),
			}
			err := isPlaylistValid(&v, filepath.Dir(v.Path))
			if err != nil {
				return err
			}
			if !v.IsValid {
				fmt.Printf("VALIDATION FAILED: %s\n", e.Path)
				// delete the new file
				err := os.Remove(e.Path)
				if err != nil {
					return err
				}
				fmt.Printf("- %s\n", e.Path)
				// TODO remove the moved music files?
				exportedReport.Skipped = append(exportedReport.Skipped, v)
			}
		}

		return nil
	},
}

func init() {
	playlistCmd.AddCommand(exportCmd)
}

// getPathWhenExportingPlaylist wants to create a playlist where all the music files
// have been moved into the playlist destination folder
func getPathWhenExportingPlaylist(path string, sourcePlDir string, destDir string) string {
	newPath := common.CreateAbsPath(filepath.Base(path), destDir)
	if !isDryRun {
		// copy the music file to the new location
		from := common.CreateAbsPath(path, sourcePlDir)
		err := common.CopyFile(from, newPath)
		if err != nil {
			fmt.Printf("COPY FAILED: %s\n", newPath)
		}
		fmt.Printf("+ %s\n", newPath)
	}
	return newPath
}

func exportPlaylists(validateReport validatePlaylistsReport) (exportPlaylistsResult, error) {
	exportReport := exportPlaylistsResult{
		Paths:    make([]string, 0),
		Exported: make([]exportedPlaylist, 0),
		Skipped:  make([]validatePlaylistResult, 0),
	}
	// Get what is needed from the validate playlists report
	exportReport.Paths = append(exportReport.Paths, validateReport.Valid...)
	exportReport.Skipped = append(exportReport.Skipped, validateReport.Invalid...)

	for _, p := range exportReport.Paths {
		e := exportedPlaylist{
			Path:       filepath.Join(outputDir, filepath.Base(p)),
			MusicFiles: make([]string, 0),
		}
		newPaths, err := getPlaylistPaths(p, getPathWhenExportingPlaylist, outputDir)
		if err != nil {
			exportReport.Skipped = append(exportReport.Skipped, validatePlaylistResult{
				Path:   p,
				Reason: err.Error(),
			})
			continue
		}

		// Keep the absolute file paths for the report
		e.MusicFiles = append(e.MusicFiles, newPaths...)

		if !isDryRun {
			// convert each newPath to be relative
			for i, n := range newPaths {
				newPaths[i] = fmt.Sprintf("./%s", filepath.Base(n))
			}
			// create the data to write to the playlist file
			newPathsStr := strings.Join(newPaths, "\n")
			// write out the exported playlist
			f, err := os.Create(e.Path)
			if err != nil {
				exportReport.Skipped = append(exportReport.Skipped, validatePlaylistResult{
					Path:   p,
					Reason: err.Error(),
				})
				continue
			}

			// close the file when done
			defer f.Close()

			// write to the file
			_, err = f.WriteString(newPathsStr)
			if err != nil {
				exportReport.Skipped = append(exportReport.Skipped, validatePlaylistResult{
					Path:   p,
					Reason: err.Error(),
				})
				continue
			}
		}

		exportReport.Exported = append(exportReport.Exported, e)
	}

	return exportReport, nil
}
