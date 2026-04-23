package cmd

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"maps"
	"os"

	"github.com/spf13/cobra"

	"music-utils/common"
)

type replacedReport struct {
	movedReport
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
		movedReport: movedReport{Moved: make([]common.FileMovedResult, 0)},
	}

	maybeValidReports, err := getReportsToValidate(configFileOrDir)
	if err != nil {
		return err
	}

	// place to store a combined report
	allExifReport := exifReport{
		Files:  make(map[string]common.TrackInfo),
		Albums: make([]common.AlbumInfo, 0),
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
		// Check if there is a match in the encrypted data on just artist|title
		encPath, ok := artistBarTitleToPathMap[key]
		if ok {
			// Once here, we know that the track.Title and track.Artist match
			// Get the TrackInfo for the drmFree track
			freeTrack := wr.Tracks[wr.TrackPathToIndex[drmFreePath]]
			// Get the trackInfo for the DRM track
			drmTrack := allExifReport.Files[encPath]
			// check that the match is exact
			if freeTrack.Album == drmTrack.Album && freeTrack.TrackNumber == drmTrack.TrackNumber {
				fmt.Println("Got stuff to do!")
			}

			// Save the match in the report in Moved slice
			fmr := common.FileMovedResult{
				Source: encPath,
				Dest:   drmFreePath,
			}
			replacedReportResult.Moved = append(replacedReportResult.Moved, fmr)
		}
	}

	//if !isDryRun {
	// delete the encrypted files
	//}

	j, _ := json.MarshalIndent(&replacedReportResult, "", "  ")
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

	ok, track, _ := common.CreateTrackInfoFromPath(path)
	if !ok {
		return nil
	} else {
		//  create the 'artist|title' key
		key := fmt.Sprintf("%s|%s", track.Artist, track.Title)
		// save to the mapStringToString in the result
		results.MapStringToString[key] = path
		// append to the Files in the result (might need it)
		results.Files = append(results.Files, path)
		// update the count
		results.Count++
		// Update the map of path to index in the Tracks slice
		results.TrackPathToIndex[path] = len(results.Tracks)
		// append the track
		results.Tracks = append(results.Tracks, track)
		// see if we have a new album
		// see if the album exists in the Albums slice
		i, ok := results.AlbumNameToIndex[track.Album]
		if ok {
			// add the TrackInfo to the AlbumInfo
			results.Albums[i].Tracks = append(results.Albums[i].Tracks, track)
			// Default to the largest number of total tracks
			if results.Albums[i].TotalTracks < track.TotalTracks {
				results.Albums[i].TotalTracks = track.TotalTracks
			}
			// Default to having an AlbumArtist
			if results.Albums[i].Artist == "" {
				results.Albums[i].Artist = track.AlbumArtist
			}
		} else {
			// save the index where we added the album into the name map
			results.AlbumNameToIndex[track.Album] = len(results.Albums)
			a := []common.TrackInfo{track}
			// create the new album in the results
			results.Albums = append(results.Albums, common.AlbumInfo{
				Album:       track.Album,
				Artist:      track.AlbumArtist,
				TotalTracks: track.TotalTracks,
				Tracks:      a,
			})
		}
	}

	return nil
}
