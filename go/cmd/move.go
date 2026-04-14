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

type movePlaylistsResult struct {
	Paths   []string                 `json:"-"`
	Moved   []common.FileMovedResult `json:"moved"`
	Skipped []validatePlaylistResult `json:"skipped"`
}

var moveCmd = &cobra.Command{
	Use:   "move",
	Short: "Moves one or many playlists to the destination directory",
	Long: `Moves one or many playlists to the destination directory.
The music files linked inside the playlist(s) remain static. The 
playlist is moved and the links within are updated.

To move one playlist:

music-utils playlist move --input-file "$HOME/Music/folder/list.m3u" -o "$HOME/Music/playlists"

To move all playlists found in the inputDir:

music-utils playlist move -i "$HOME/Music" -o "$HOME/Music/playlists

To move all playlists recursively found in the inputDir and all sub-directories:

music-utils playlist move -i "$HOME/Music" -r -o "$HOME/Music/playlists`,
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
		movedReport, err := movePlaylists(r)
		if err != nil {
			return err
		}

		// If a dry run, we can't do anything else but print the report
		if isDryRun {
			j, _ := json.Marshal(&movedReport)
			fmt.Println(string(j))
			return nil
		}

		// validate the playlists that we moved
		for _, m := range movedReport.Moved {
			v := validatePlaylistResult{
				IsValid:  true,
				Path:     m.Dest,
				Reason:   "",
				BadPaths: make([]string, 0),
			}
			err := isPlaylistValid(&v, filepath.Dir(v.Path))
			if err != nil {
				return err
			}
			if v.IsValid {
				// delete the old file
				err := os.Remove(m.Source)
				if err != nil {
					return err
				}
				// print to user that we deleted the file
				fmt.Printf("- %s\n", m.Source)
			} else {
				fmt.Printf("VALIDATION FAILED: %s\n", m.Dest)
				// delete the new file
				err := os.Remove(m.Dest)
				if err != nil {
					return err
				}
				fmt.Printf("- %s\n", m.Dest)
				movedReport.Skipped = append(movedReport.Skipped, v)
			}
		}

		return nil
	},
}

func init() {
	playlistCmd.AddCommand(moveCmd)
}

// getPathWhenMovingPlaylist is a process function for getting the path of a music file
// when the playlist is being moved to a new location while the music files remain in
// place
func getPathWhenMovingPlaylist(path string, sourcePlPath string, destPlDir string) string {
	// create the new path
	if !filepath.IsAbs(path) {
		path = common.MoveRelativePath(path, filepath.Dir(sourcePlPath), destPlDir)
	}
	return path
}

func movePlaylists(validateReport validatePlaylistsReport) (movePlaylistsResult, error) {
	m := movePlaylistsResult{
		Paths:   make([]string, 0),
		Moved:   make([]common.FileMovedResult, 0),
		Skipped: make([]validatePlaylistResult, 0),
	}
	// Get what is needed from the validate playlists report
	m.Paths = append(m.Paths, validateReport.Valid...)
	m.Skipped = append(m.Skipped, validateReport.Invalid...)

	for _, p := range m.Paths {
		r := common.FileMovedResult{
			Source: p,
			Dest:   filepath.Join(outputDir, filepath.Base(p)),
		}
		newPaths, err := getPlaylistPaths(p, getPathWhenMovingPlaylist, outputDir)
		if err != nil {
			m.Skipped = append(m.Skipped, validatePlaylistResult{
				Path:   r.Source,
				Reason: err.Error(),
			})
			continue
		}
		if !isDryRun {
			// create the data to write into the file
			newPathsStr := strings.Join(newPaths, "\n")
			// check if the file already exists
			if common.FileExists(r.Dest) {
				// create a message for the overwrite prompt
				msg := fmt.Sprintf(`File already exists!
Overwrite playlist at:
%s
`, r.Dest)
				// Prompt to overwrite
				didSave, err := common.PromptAndMaybeSaveFile(r.Dest, []byte(newPathsStr), msg)
				if err != nil {
					m.Skipped = append(m.Skipped, validatePlaylistResult{
						Path:   r.Source,
						Reason: fmt.Sprintf("Error occurred when saving: %s", err),
					})
					continue
				}
				if !didSave {
					m.Skipped = append(m.Skipped, validatePlaylistResult{
						Path:   r.Source,
						Reason: "User canceled. File already exists",
					})
					continue
				}
			} else {
				f, err := os.Create(r.Dest)
				if err != nil {
					m.Skipped = append(m.Skipped, validatePlaylistResult{
						Path:   r.Source,
						Reason: err.Error(),
					})
					continue
				}

				// close the file when done
				defer f.Close()

				// write to the file
				_, err = f.WriteString(newPathsStr)
				if err != nil {
					m.Skipped = append(m.Skipped, validatePlaylistResult{
						Path:   r.Source,
						Reason: err.Error(),
					})
					continue
				}
			}
			fmt.Printf("+ %s\n", r.Dest)
		}
		m.Moved = append(m.Moved, r)
	}

	return m, nil
}
