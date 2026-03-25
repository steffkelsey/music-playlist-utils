package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/dhowden/tag"

	"github.com/spf13/cobra"

	"music-utils/common"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Indicates if the required metadata exist in each file",
	Long: `Indicates if the required metadata exists in each music file 
in the given input directory.	
Music files must have a mp3, mp4, m4a, or m4p extension.
eg:
Check all the files in the root folder $HOME/Music and export any without tags to $HOME/no-tags
music-utils check -i "$HOME/Music" -o "$HOME/no-tags"`,
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
		err := findUntaggedFiles(inputDir)
		return err
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}

func findUntaggedFiles(rootPath string) error {
	res, err := common.WalkAllMusicFiles(rootPath, isFileUntagged)
	if err != nil {
		return err
	}
	var sb strings.Builder
	sb.WriteString(`{"untagged": `)
	sb.WriteString(fmt.Sprintf("[%s]", strings.Join(res.Files, ",")))
	// dump the string
	jsonString := sb.String()
	// close the object
	jsonString += "}"
	fmt.Println(jsonString)
	return nil
}

func isFileUntagged(path string, info fs.FileInfo, results *common.WalkResults) error {
	// Open the file to get more details
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("  Error opening file: %v\n", err)
	} else {
		defer file.Close()

		// Use dhowden/tag to read metadata
		m, err := tag.ReadFrom(file)
		if err != nil {
			results.Files = append(results.Files, fmt.Sprintf("\"%s\"", path))
			results.Count++
		} else {
			isTagGood := true
			// Must have Album, Title, Track, Artist
			if m.Album() == "" {
				isTagGood = false
			}

			if m.Title() == "" {
				isTagGood = false
			}

			if m.Artist() == "" {
				isTagGood = false
			}

			trackNum, _ := m.Track()
			if trackNum == 0 {
				isTagGood = false
			}

			if !isTagGood {
				results.Files = append(results.Files, fmt.Sprintf("\"%s\"", path))
				results.Count++
			}
		}
	}

	return nil
}
