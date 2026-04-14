package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"music-utils/common"
)

var inputFile string
var isExport bool
var isRecursive bool
var validateOnly bool

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

type exportedPlaylist struct {
	Path       string   `json:"path"`
	MusicFiles []string `json:"musicFiles"`
}

type exportPlaylistsResult struct {
	Paths    []string                 `json:"-"`
	Exported []exportedPlaylist       `json:"exported"`
	Skipped  []validatePlaylistResult `json:"skipped"`
}

type movePlaylistsResult struct {
	Paths   []string                 `json:"-"`
	Moved   []common.FileMovedResult `json:"moved"`
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

music-utils playlist -i "$HOME/Music" -r -v

To export one playlist all linked music files:

music-utils playlist --input-file "$HOME/Music/playlist 1.m3u" -e -o $HOME/Music/playlist 1"`,
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
			err := isPlaylistValid(&v, filepath.Dir(v.Path))
			if err != nil {
				return err
			}
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

		if isExport {
			exportedResult, err := exportPlaylists(validateResult)
			if err != nil {
				return err
			}
			if isDryRun {
				j, _ := json.Marshal(&exportedResult)
				fmt.Println(string(j))
				return nil
			}
			// validate the exported playlists
			for _, e := range exportedResult.Exported {
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
				if v.IsValid {
					// TODO
				} else {
					// TODO
				}
			}
		} else {
			movedResult, err := movePlaylists(validateResult)
			if err != nil {
				return err
			}
			// If a dry run, we can't do anything else but print the report
			if isDryRun {
				j, _ := json.Marshal(&movedResult)
				fmt.Println(string(j))
				return nil
			}
			// validate the moved playlists
			for _, m := range movedResult.Moved {
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
					movedResult.Skipped = append(movedResult.Skipped, v)
				}
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(playlistCmd)

	playlistCmd.Flags().StringVar(&inputFile, "input-file", "", "Single playlist file to move")

	playlistCmd.Flags().BoolVarP(&isRecursive, common.ParamRecursive, "r", false, "Searches all sub-directories")
	playlistCmd.Flags().BoolVarP(&validateOnly, common.ParamValidate, "v", false, "Validate playlist file(s) only. No changes are made")
	playlistCmd.Flags().BoolVarP(&isExport, common.ParamExport, "e", false, "Export playlist and linked music files (creates copies)")
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

// Gets playlists paths from the source Playlist path and alters them to work
// for the playlist to be moved to the destination directory (music files are static)
func getPlaylistPaths(sourcePlPath string, func processFunc(path string, plSourcePath string, dest string) string) ([]string, error) {
	newPaths := make([]string, 0)
	file, err := os.Open(sourcePlPath)
	if err != nil {
		return newPaths, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		p := scanner.Text()
		p = processFunc(p, filepath.Dir(sourcePlPath), destDir)
		newPaths = append(newPaths, p)
	}

	if err := scanner.Err(); err != nil {
		return newPaths, err
	}

	return newPaths, nil
}

// getPathWhenMovingPlaylist is a process function for getting the path of a music file
// when the playlist is being moved to a new location while the music files remain in
// place
func getPathWhenMovingPlaylist(path string, plSourcePath string, plDestpath string) string {
	// create the new path
	if !filepath.IsAbs(path) {
		path = common.MoveRelativePath(p, filepath.Dir(sourcePlPath), destDir)
	}
	return path
}

// getPathWhenExportingPlaylist wants to create a playlist where all the music files
// have been moved into the playlist destination folder
func getPathWhenExportingPlaylist(path string, plSourcePath string, destPath string) string {
	newPath = fmt.Sprintf("./%s", filepath.Base(path))
	if !isDryRun {
		// TODO actually copy the music files?
	}
	return newPath
}

func movePlaylists(validateReport validatePlaylistsReport) (movePlaylistsResult, error) {
	toMove := movePlaylistsResult{
		Paths:   make([]string, 0),
		Moved:   make([]common.FileMovedResult, 0),
		Skipped: make([]validatePlaylistResult, 0),
	}
	// Get what is needed from the validate playlists report
	toMove.Paths = append(toMove.Paths, validateReport.Valid...)
	toMove.Skipped = append(toMove.Skipped, validateReport.Invalid...)

	for _, p := range toMove.Paths {
		r := common.FileMovedResult{
			Source: p,
			Dest:   filepath.Join(outputDir, filepath.Base(p)),
		}
		newPaths, err := getPlaylistPaths(p, getPathWhenMovingPlaylist)
		if err != nil {
			toMove.Skipped = append(toMove.Skipped, validatePlaylistResult{
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
					toMove.Skipped = append(toMove.Skipped, validatePlaylistResult{
						Path:   r.Source,
						Reason: fmt.Sprintf("Error occurred when saving: %s", err),
					})
					continue
				}
				if !didSave {
					toMove.Skipped = append(toMove.Skipped, validatePlaylistResult{
						Path:   r.Source,
						Reason: "User canceled. File already exists",
					})
					continue
				}
			} else {
				f, err := os.Create(r.Dest)
				if err != nil {
					toMove.Skipped = append(toMove.Skipped, validatePlaylistResult{
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
					toMove.Skipped = append(toMove.Skipped, validatePlaylistResult{
						Path:   r.Source,
						Reason: err.Error(),
					})
					continue
				}
			}
			fmt.Printf("+ %s\n", r.Dest)
		}
		toMove.Moved = append(toMove.Moved, r)
	}

	return toMove, nil
}

func exportPlaylists(validateReport validatePlaylistsReport) (exportPlaylistsResult, error) {
	toExport := exportPlaylistsResult{
		Paths:    make([]string, 0),
		Exported: make([]exportedPlaylist, 0),
		Skipped:  make([]validatePlaylistResult, 0),
	}
	// Get what is needed from the validate playlists report
	toExport.Paths = append(toExport.Paths, validateReport.Valid...)
	toExport.Skipped = append(toExport.Skipped, validateReport.Invalid...)
	// TODO
	return toExport, nil
}
