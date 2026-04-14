package cmd

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"music-utils/common"
)

var inputFile string
var isRecursive bool

var playlistCmd = &cobra.Command{
	Use:   "playlist",
	Short: "Commands for validating, exporting, and moving playlists",
	Long: `Command for validating, exporting, and moving playlists. 
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
		fmt.Println("playlist called")
		return nil
	},
	//RunE: func(cmd *cobra.Command, args []string) error {
	//	toValidate := make([]validatePlaylistResult, 0)

	//	if len(inputFile) > 0 {
	//		// expand to deal with possible env variables like $HOME
	//		inputFile = os.ExpandEnv(inputFile)
	//		// verify input file exists
	//		if !common.FileExists(inputFile) {
	//			return fmt.Errorf("playlist does not exist at %s", inputFile)
	//		}
	//		// add the path to toValidate slice
	//		toValidate = append(toValidate, validatePlaylistResult{
	//			IsValid:  true,
	//			Path:     inputFile,
	//			Reason:   "",
	//			BadPaths: make([]string, 0),
	//		})

	//	} else {
	//		var res common.WalkResults
	//		var err error
	//		if isRecursive {
	//			// walk recursively finding all playlist files
	//			res, err = common.WalkAllMusicFiles(inputDir, findPlaylistFile)
	//			if err != nil {
	//				return err
	//			}
	//		} else {
	//			// walk just the inputDir finding all playlist files
	//			res, err = common.WalkAllMusicFilesNotRecursive(inputDir, findPlaylistFile)
	//			if err != nil {
	//				return err
	//			}
	//		}
	//		for _, p := range res.Files {
	//			toValidate = append(toValidate, validatePlaylistResult{
	//				IsValid:  true,
	//				Path:     p,
	//				Reason:   "",
	//				BadPaths: make([]string, 0),
	//			})
	//		}
	//	}

	//	validateResult := validatePlaylistsReport{
	//		Valid:   make([]string, 0),
	//		Invalid: make([]validatePlaylistResult, 0),
	//	}
	//	for _, v := range toValidate {
	//		err := isPlaylistValid(&v, filepath.Dir(v.Path))
	//		if err != nil {
	//			return err
	//		}
	//		if v.IsValid {
	//			validateResult.Valid = append(validateResult.Valid, v.Path)
	//		} else {
	//			validateResult.Invalid = append(validateResult.Invalid, v)
	//		}
	//	}

	//	if validateOnly {
	//		// print the json report
	//		j, _ := json.Marshal(&validateResult)
	//		fmt.Println(string(j))
	//		return nil
	//	}

	//	if isExport {
	//		exportedResult, err := exportPlaylists(validateResult)
	//		if err != nil {
	//			return err
	//		}
	//		if isDryRun {
	//			j, _ := json.Marshal(&exportedResult)
	//			fmt.Println(string(j))
	//			return nil
	//		}
	//		// validate the exported playlists
	//		for _, e := range exportedResult.Exported {
	//			v := validatePlaylistResult{
	//				IsValid:  true,
	//				Path:     e.Path,
	//				Reason:   "",
	//				BadPaths: make([]string, 0),
	//			}
	//			err := isPlaylistValid(&v, filepath.Dir(v.Path))
	//			if err != nil {
	//				return err
	//			}
	//			if v.IsValid {
	//				// TODO
	//			} else {
	//				// TODO
	//			}
	//		}
	//	} else {
	//		movedResult, err := movePlaylists(validateResult)
	//		if err != nil {
	//			return err
	//		}
	//		// If a dry run, we can't do anything else but print the report
	//		if isDryRun {
	//			j, _ := json.Marshal(&movedResult)
	//			fmt.Println(string(j))
	//			return nil
	//		}
	//		// validate the moved playlists
	//		for _, m := range movedResult.Moved {
	//			v := validatePlaylistResult{
	//				IsValid:  true,
	//				Path:     m.Dest,
	//				Reason:   "",
	//				BadPaths: make([]string, 0),
	//			}
	//			err := isPlaylistValid(&v, filepath.Dir(v.Path))
	//			if err != nil {
	//				return err
	//			}
	//			if v.IsValid {
	//				// delete the old file
	//				err := os.Remove(m.Source)
	//				if err != nil {
	//					return err
	//				}
	//				// print to user that we deleted the file
	//				fmt.Printf("- %s\n", m.Source)
	//			} else {
	//				fmt.Printf("VALIDATION FAILED: %s\n", m.Dest)
	//				// delete the new file
	//				err := os.Remove(m.Dest)
	//				if err != nil {
	//					return err
	//				}
	//				fmt.Printf("- %s\n", m.Dest)
	//				movedResult.Skipped = append(movedResult.Skipped, v)
	//			}
	//		}
	//	}

	//	return nil
	//},
}

func init() {
	rootCmd.AddCommand(playlistCmd)

	playlistCmd.PersistentFlags().StringVar(&inputFile, "input-file", "", "Single playlist file to move")

	playlistCmd.PersistentFlags().BoolVarP(&isRecursive, common.ParamRecursive, "r", false, "Searches all sub-directories")
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

// Gets playlists paths from the source Playlist path and alters them to work
// for the playlist to be moved to the destination directory (music files are static)
func getPlaylistPaths(sourcePlPath string, processFunc func(path string, sourcePlDir string, destDir string) string, destDir string) ([]string, error) {
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
