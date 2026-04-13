package cmd

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"

	"github.com/dhowden/tag"

	"github.com/spf13/cobra"

	"music-utils/common"
)

type untaggedResult struct {
	Path    string   `json:"path"`
	Reasons []string `json:"reasons"`
}

type untaggedReport struct {
	CopiesCreated []string         `json:"copiesCreated"`
	Untagged      []untaggedResult `json:"untagged"`
}

var untaggedReportResult untaggedReport

var untaggedCmd = &cobra.Command{
	Use:   "untagged",
	Short: "Copies files where required metadata does NOT exist",
	Long: `Copies files where the required metadata does NOT exists in the given 
input directory to the given output directory preserving subfolders. Dry run 
shows a JSON report of where files would have been copied to.
Music files must have a mp3, mp4, m4a, or m4p extension.
eg:
Check all the files in the root folder $HOME/Music and export any without tags 
to $HOME/no-tags

music-utils untagged -i "$HOME/Music" -o "$HOME/no-tags"`,
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
	rootCmd.AddCommand(untaggedCmd)
}

func findUntaggedFiles(rootPath string) error {
	untaggedReportResult = untaggedReport{
		CopiesCreated: make([]string, 0),
		Untagged:      make([]untaggedResult, 0),
	}
	res, err := common.WalkAllMusicFiles(rootPath, isFileUntagged)
	if err != nil {
		return err
	}
	// go through the files and copy each one to the target destination
	// create the report for each
	for _, path := range res.Files {
		common.Sanitize(&path)
		newPath := common.SwapRoot(path, rootPath, outputDir)
		if !isDryRun {
			// copy to new location
			err := common.CopyFile(path, newPath)
			if err != nil {
				fmt.Printf("error copying file, err: %v\n", err)
				return err
			}
			// Print showing the file being created in the new location
			fmt.Printf("+ %s\n", newPath)
		}
		// save for the report
		untaggedReportResult.CopiesCreated = append(untaggedReportResult.CopiesCreated, newPath)
	}
	j, _ := json.Marshal(&untaggedReportResult)
	jsonString := string(j)
	if isDryRun {
		fmt.Println(jsonString)
	}
	return nil
}

func isFileUntagged(path string, info fs.FileInfo, results *common.WalkResults) error {
	// skip if the file is encrypted or a playlist
	if common.IsEncryptedFile(path) || common.IsPlaylistFile(path) {
		return nil
	}
	r := untaggedResult{
		Path:    path,
		Reasons: make([]string, 0),
	}
	// Open the file to get more details
	file, err := os.Open(path)
	if err != nil {
		r.Reasons = append(r.Reasons, "Could not open file")
		untaggedReportResult.Untagged = append(untaggedReportResult.Untagged, r)
		results.Files = append(results.Files, path)
	} else {
		defer file.Close()

		// Use dhowden/tag to read metadata
		m, err := tag.ReadFrom(file)
		if err != nil {
			r.Reasons = append(r.Reasons, "Could not read tags")
			untaggedReportResult.Untagged = append(untaggedReportResult.Untagged, r)
			results.Files = append(results.Files, path)
			results.Count++
		} else {
			isTagGood := true
			// Must have Album, Title, Track, Artist
			if m.Album() == "" {
				r.Reasons = append(r.Reasons, "Missing Album tag")
				isTagGood = false
			}

			if m.Title() == "" {
				r.Reasons = append(r.Reasons, "Missing Title tag")
				isTagGood = false
			}

			if m.Artist() == "" {
				r.Reasons = append(r.Reasons, "Missing Artist tag")
				isTagGood = false
			}

			trackNum, _ := m.Track()
			if trackNum == 0 {
				r.Reasons = append(r.Reasons, "Missing TrackNumber tag")
				isTagGood = false
			}

			if !isTagGood {
				results.Files = append(results.Files, path)
				untaggedReportResult.Untagged = append(untaggedReportResult.Untagged, r)
				results.Count++
			}
		}
	}

	return nil
}
