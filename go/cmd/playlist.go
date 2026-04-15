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
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("playlist called")
	},
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
