package cmd

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	"github.com/dhowden/tag"
	"github.com/spf13/cobra"

	"music-utils/common"
)

type organizedReport struct {
	Moved    []common.FileMovedResult `json:"moved"`
	Untagged []untaggedResult         `json:"untagged"`
}

var organizedReportResult organizedReport

var organizeCmd = &cobra.Command{
	Use:   "organize",
	Short: "Organizes music by tags",
	Long: `Organizes music by tags. 
This lacks the flexibility of Picard, but you can do
dry-runs and it outputs a report that can be used to
repair any playlists that are broken using the playlist repair
cmd. The organization of files aims to keep music files from 
the same album together (our target is Jellyfin). 

The files are organized IN PLACE, the outputDir is only
used for saving the export report.

To validate repair music recursively found in the input and outputs 
a report to the ~/Music/reports folder:

music-utils organize -i $HOME/Music -o $HOME/Music/reports
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
		return organizeMusicFiles()
	},
}

func init() {
	rootCmd.AddCommand(organizeCmd)
}

func organizeMusicFiles() error {
	organizedReportResult = organizedReport{
		Moved:    make([]common.FileMovedResult, 0),
		Untagged: make([]untaggedResult, 0),
	}
	_, err := common.WalkAllMusicFiles(inputDir, createDestinationFromTags)
	if err != nil {
		return err
	}
	// go through the files and copy each one to the target destination
	for i, m := range slices.Backward(organizedReportResult.Moved) {
		common.Sanitize(&m.Dest)
		// We get the file as a relative path from the isFileTagged function
		// Make the path absolute before attempting to copy the file
		// It looks confusing that we're using inputDir for Dest,
		// but we're organizing the files in place and the outputDir
		// is used only for generating the report.
		m.Dest = filepath.Join(inputDir, m.Dest)
		organizedReportResult.Moved[i].Dest = m.Dest
		if !isDryRun {
			// TODO any overwrite warnings?
			// copy Source -> Dest (function handles creation of directories etc)
			err := common.CopyFile(m.Source, m.Dest)
			fmt.Printf("+ %s\n", m.Dest)
			if err != nil {
				// update the report by removing this file from Moved
				organizedReportResult.Moved = append(organizedReportResult.Moved[:i], organizedReportResult.Moved[i+1:]...)
				fmt.Printf("error copying file. err: %v\n", err)
				// remove the file if it was created
				os.Remove(m.Dest)
				// TODO add anything to the report that we skipped this file?
				continue
			}
			// delete the source file
			os.Remove(m.Source)
			fmt.Printf("- %s\n", m.Source)
		}
	}
	j, _ := json.Marshal(&organizedReportResult)
	jsonString := string(j)
	if isDryRun {
		fmt.Println(jsonString)
	} else {
		// create a destination for the report
		reportPath := filepath.Join(outputDir, "organized.json")
		// We don't want to overwrite reports, so make sure the path is unique
		reportPath = common.FindFileNameNoOverWrite(reportPath)
		// create the file at the path
		f, err := os.Create(reportPath)
		if err != nil {
			return err
		}

		// close the file when done
		defer f.Close()

		// write to the file
		_, err = f.WriteString(jsonString)
		if err != nil {
			return err
		}

		// clean up any empty directories that may have been left behind
		err = common.RemoveEmptyDirectories(inputDir)
		if err != nil {
			return err
		}
	}
	return nil
}

func createDestinationFromTags(path string, info fs.FileInfo, results *common.WalkResults) error {
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
		organizedReportResult.Untagged = append(organizedReportResult.Untagged, r)
		results.Files = append(results.Files, path)
	} else {
		defer file.Close()

		var title string
		var album string
		var artist string
		var trackNumber string
		// Use dhowden/tag to read metadata
		m, err := tag.ReadFrom(file)
		if err != nil {
			r.Reasons = append(r.Reasons, "Could not read tags")
			organizedReportResult.Untagged = append(organizedReportResult.Untagged, r)
			results.Files = append(results.Files, path)
			results.Count++
		} else {
			isTagGood := true
			// Must have Album, Title, Track, Artist
			if m.Album() == "" {
				r.Reasons = append(r.Reasons, "Missing Album tag")
				isTagGood = false
			} else {
				album = m.Album()
			}

			if m.Title() == "" {
				r.Reasons = append(r.Reasons, "Missing Title tag")
				isTagGood = false
			} else {
				title = m.Title()
			}

			// This is Track artist NOT album artist
			if m.Artist() == "" {
				r.Reasons = append(r.Reasons, "Missing Artist tag")
				isTagGood = false
			} else {
				artist = m.Artist()
			}

			trackNum, _ := m.Track()
			if trackNum == 0 {
				r.Reasons = append(r.Reasons, "Missing Track Number tag")
				isTagGood = false
			} else {
				trackNumber = fmt.Sprintf("%02d", trackNum)
			}

			if !isTagGood {
				results.Files = append(results.Files, path)
				organizedReportResult.Untagged = append(organizedReportResult.Untagged, r)
				results.Count++
			} else {
				// Desired destination is:
				// ./[Artist]/[Album]/[Track Number] - [Title].ext
				// BUT we want to optimize that the music tracks of the
				// same album are in the same folder for Jellyfin (for serving)
				// or Picard (for tag editing).
				// So we are going to start with:
				// ./[Album]/[Track Number] - [Artist] - [Title].ext
				dest := fmt.Sprintf("./%s/%s - %s - %s%s", album, trackNumber, artist, title, filepath.Ext(path))
				m := common.FileMovedResult{
					Source: path,
					Dest:   dest,
				}
				organizedReportResult.Moved = append(organizedReportResult.Moved, m)
			}
		}
	}

	return nil
}
