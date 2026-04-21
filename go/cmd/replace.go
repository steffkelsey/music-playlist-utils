package cmd

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"maps"
	"os"

	"github.com/dhowden/tag"
	"github.com/spf13/cobra"

	"music-utils/common"
)

type replacedReport struct {
	Moved []common.FileMovedResult `json:"moved"`
}

var replaceCmd = &cobra.Command{
	Use:   "replace",
	Short: "Replaces encrypted music files with matching DRM-free versions.",
	Long: `Replaces encrypted music files with matching DRM-free versions.
Inputs an exif report detailing the files to replace plus an 
input directory containing the possible replacements.
Exports a json report of files moved that can be used to 
repair any playlists damaged in the process.

To replace encrypted files enumerated in the json with downloaded files from the ~/Music/dl folder and save the report in ~/Music: 
./music-utils encrypted replace -i $HOME/Music/dl -c $HOME/Music/encrypted-exif.json -o $HOME/Music

The location of the files being replaced is in the json.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		// Verify that input dir exists
		inputDir, err = common.FlagDirectoryExists(inputDir)
		if err != nil {
			return err
		}
		// Verify the config file or directory exists (also expands the path)
		configFileOrDir, err = common.FlagDirectoryExists(configFileOrDir)
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
		return replaceEncryptedFiles()
	},
}

func init() {
	encryptedCmd.AddCommand(replaceCmd)
	replaceCmd.Flags().StringVarP(&configFileOrDir, "config-path", "c", "", "Config file or directory containing multiple")
}

func replaceEncryptedFiles() error {
	replacedReportResult := replacedReport{
		Moved: make([]common.FileMovedResult, 0),
	}

	maybeValidReports, err := getReportsToValidate(configFileOrDir)
	if err != nil {
		return err
	}

	// place to store a combined report
	allExifReport := exifReport{
		Files:  make(map[string]trackInfo),
		Albums: make([]albumInfo, 0),
	}
	// we need a map of 'artist|title': 'path'
	artistBarTitleToPathMap := make(map[string]string)
	// Get all the data
	for _, j := range maybeValidReports {
		ok, exr := isValidExifReport(j)
		if ok {
			// copy into the combined report
			maps.Copy(allExifReport.Files, exr.Files)
			allExifReport.Albums = append(allExifReport.Albums, exr.Albums...)

			// for each trackInfo in exr.Files
			for trckPath, trckInfo := range exr.Files {
				// create the trackArtist|trackTitle key
				key := fmt.Sprintf("%s|%s", trckInfo.Artist, trckInfo.Title)
				artistBarTitleToPathMap[key] = trckPath
			}
		}
	}

	// Now, walk all the files in the input folder
	wr, err := common.WalkAllMusicFiles(inputDir, createKeyWithTags)
	if err != nil {
		return err
	}

	// iterate over the tagged music files
	for key, drmFreePath := range wr.MapStringToString {
		// Check if there is a match in the encrypted data
		encPath, ok := artistBarTitleToPathMap[key]
		if ok {
			// Save the match in the report in Moved slice
			fmr := common.FileMovedResult{
				Source: encPath,
				Dest:   drmFreePath,
			}
			replacedReportResult.Moved = append(replacedReportResult.Moved, fmr)
		}
	}

	j, _ := json.Marshal(&replacedReportResult)
	jsonString := string(j)
	if isDryRun {
		fmt.Println(jsonString)
	}

	return nil
}

func isValidExifReport(path string) (bool, exifReport) {
	var r exifReport
	// open and read the whole file (json is usually tiny)
	data, err := os.ReadFile(path)
	if err != nil {
		return false, r
	}

	// attempt to unmarshal into a exifReport struct
	err = json.Unmarshal(data, &r)
	if err != nil {
		return false, r
	}
	return true, r
}

func createKeyWithTags(path string, info fs.FileInfo, results *common.WalkResults) error {
	// skip if the file is encrypted or a playlist
	if common.IsEncryptedFile(path) || common.IsPlaylistFile(path) {
		return nil
	}
	// Open the file to get more details
	file, err := os.Open(path)
	if err != nil {
		return nil
	} else {
		defer file.Close()

		var title string
		//var album string
		var artist string
		//var trackNumber string
		// Use dhowden/tag to read metadata
		m, err := tag.ReadFrom(file)
		if err != nil {
			return nil
		} else {
			isTagGood := true
			// Must have Album, Title, Track, Artist
			if m.Album() == "" {
				isTagGood = false
				//} else {
				//	album = m.Album()
			}

			if m.Title() == "" {
				isTagGood = false
			} else {
				title = m.Title()
			}

			// This is Track artist NOT album artist
			if m.Artist() == "" {
				isTagGood = false
			} else {
				artist = m.Artist()
			}

			trackNum, _ := m.Track()
			if trackNum == 0 {
				isTagGood = false
				//} else {
				//	trackNumber = fmt.Sprintf("%02d", trackNum)
			}

			if !isTagGood {
				return nil
			} else {
				//  create the 'artist|title' key
				key := fmt.Sprintf("%s|%s", artist, title)
				// save to the mapStringToString in the result
				results.MapStringToString[key] = path
				// append to the Files in the result (might need it)
				results.Files = append(results.Files, path)
			}
		}
	}

	return nil
}
