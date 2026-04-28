package cmd

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"music-utils/common"
)

type replacedReport struct {
	Matches []common.TrackMatch `json:"matches"`
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
	// Get all the data
	for _, j := range maybeValidReports {
		ok, exr := isValidExifReport(j)
		if ok {
			// copy into the combined report
			maps.Copy(allExifReport.Files, exr.Files)
			allExifReport.Albums = append(allExifReport.Albums, exr.Albums...)
		} else {
			fmt.Printf("invalid report at %s\n", j)
		}
	}

	// Now, walk all the files in the input folder
	wr, err := common.WalkAllMusicFiles(inputDir, createKeyWithTags)
	if err != nil {
		return err
	}

	// Sort the Tracks slice
	slices.SortFunc(wr.Tracks, common.CmpTrackInfoTitle)
	// create map for searching on artist|title for TrackInfo
	// create map for searching on title for TrackInfo
	artistTitleToIndices := make(map[string][]int)
	titleToIndices := make(map[string][]int)
	for i, t := range wr.Tracks {
		k := strings.ToLower(fmt.Sprintf("%s|%s", t.Artist, t.Title))
		_, ok := artistTitleToIndices[k]
		if ok {
			artistTitleToIndices[k] = append(artistTitleToIndices[k], i)
		} else {
			artistTitleToIndices[k] = make([]int, 1)
			artistTitleToIndices[k][0] = i
		}
		k = strings.ToLower(t.Title)
		_, ok = titleToIndices[k]
		if ok {
			titleToIndices[k] = append(titleToIndices[k], i)
		} else {
			titleToIndices[k] = make([]int, 1)
			titleToIndices[k][0] = i
		}
	}

	// Sort the Albums slice
	slices.SortFunc(wr.Albums, common.CmpAlbumInfoAlbumTitle)
	// create map for searching on album title for AlbumInfo
	albumToIndices := make(map[string][]int)
	for i, a := range wr.Albums {
		k := strings.ToLower(a.Album)
		_, ok := albumToIndices[k]
		if ok {
			albumToIndices[k] = append(albumToIndices[k], i)
		} else {
			albumToIndices[k] = make([]int, 1)
			albumToIndices[k][0] = i
		}
	}

	// At this point, we have multiple ways to search in our DRM-Free music
	// for matches in the slices of DRM Tracks and DRM albums.
	// We can look for matching artist|title in one map
	// We can look for matching title in another map
	// We can binary search by track title in the Tracks slice for fuzzy matches
	// And for albums, we can look for matches using the Album title as a key
	// And we can binary search the Albums slice for fuzzy matches

	// iterate over the DRM tracks
OUTER:
	for drmPath, drmTrack := range allExifReport.Files {
		drmTrack.Path = drmPath
		// spot for the current leader in matching the track
		bestScore := 0.0
		bestIndex := -1
		// quickest way to matching tracks in well tagged lib, is for
		// artist|title to match
		k := strings.ToLower(fmt.Sprintf("%s|%s", drmTrack.Artist, drmTrack.Title))
		freeTrackIndices, ok := artistTitleToIndices[k]
		if ok {
			// for each matching index, do we have a perfect match?
			// if yes, mark and continue
			// if partial match, is it close enough?
			for _, i := range freeTrackIndices {
				// check for perfect match
				score := common.CmpAlbumTracks(drmTrack, wr.Tracks[i])
				if score+0.001 > 1.0 {
					// is perfect! search for this track is over!
					// Save the match in the report in Moved slice
					fmr := common.FileMovedResult{
						Source: drmPath,
						Dest:   wr.Tracks[i].Path,
					}
					replacedReportResult.Moved = append(replacedReportResult.Moved, fmr)
					replacedReportResult.Matches = append(replacedReportResult.Matches,
						common.FmtTrackMatch(drmTrack, wr.Tracks[i], score, true))
					continue OUTER
				} else if score > 0.5 {
					// Is good enough to hold onto but keep looking
					bestScore = score
					bestIndex = i
				}
			}
		}

		// we never found a perfect match, we can decide to stop looking
		// and take a very good match
		if bestScore > 0.85 {
			fmr := common.FileMovedResult{
				Source: drmPath,
				Dest:   wr.Tracks[bestIndex].Path,
			}
			replacedReportResult.Moved = append(replacedReportResult.Moved, fmr)
			replacedReportResult.Matches = append(replacedReportResult.Matches,
				common.FmtTrackMatch(drmTrack, wr.Tracks[bestIndex], bestScore, true))
			continue OUTER
		}

		// if we're here, we either had no matches at all or a score for
		// matching an album specific track less than 85% (which might be too high)
		// We could be in the situation where the artist did not match (often the
		// case when a track has multiple artist but the tags did not include them
		// all for both or described multiple artists without exact matching language).
		// Below we will solve for matching mathcing titles but not perfectly matching
		// artists.
		// The long tail being where the target is that one or both of the tracks is
		// from Greatets Hits or a Soundtrack album but the recordings are close enough
		freeTrackIndices, ok = titleToIndices[strings.ToLower(drmTrack.Title)]
		if ok {
			// for each matching index, do we have a perfect match?
			// if yes, mark and continue
			// if partial match, is it close enough?
			for _, i := range freeTrackIndices {
				// check using album information first
				// check for really good to perfect match
				// we are still matching across album info here
				// where track number and total tracks is heavily weighted
				score := common.CmpAlbumTracks(drmTrack, wr.Tracks[i])
				if score > 0.85 {
					// is good enough! search for this track is over!
					// Save the match in the report in Moved slice
					fmr := common.FileMovedResult{
						Source: drmPath,
						Dest:   wr.Tracks[i].Path,
					}
					replacedReportResult.Moved = append(replacedReportResult.Moved, fmr)
					replacedReportResult.Matches = append(replacedReportResult.Matches,
						common.FmtTrackMatch(drmTrack, wr.Tracks[i], score, true))
					continue OUTER
				} else if score > 0.5 {
					// Is good enough to hold onto but keep looking
					bestScore = score
					bestIndex = i
				}

				// That didn't work, let's get the Duration and go
				// for a match where we don't include Album, track number,
				// and total tracks.
				// Get the duration if wr.Tracks[i] does not have it
				if wr.Tracks[i].DurationSeconds == 0 {
					d, err := common.GetDuration(wr.Tracks[i].Path)
					if err != nil {
						fmt.Printf("error getting duration. %+v\n", err)
					} else {
						wr.Tracks[i].DurationSeconds = int(d)
					}
				}
				// check for perfect match
				score = common.CmpTracks(drmTrack, wr.Tracks[i])
				if score+0.001 > 1.0 {
					// is perfect! search for this track is over!
					// Save the match in the report in Moved slice
					fmr := common.FileMovedResult{
						Source: drmPath,
						Dest:   wr.Tracks[i].Path,
					}
					replacedReportResult.Moved = append(replacedReportResult.Moved, fmr)
					replacedReportResult.Matches = append(replacedReportResult.Matches,
						common.FmtTrackMatch(drmTrack, wr.Tracks[i], score, true))
					continue OUTER
				} else if score > 0.5 {
					//fmt.Printf("score: %.2f\n", score)
					//fmt.Printf("%s | %s\n", drmTrack.Title, wr.Tracks[i].Title)
					//fmt.Printf("%s | %s\n", drmTrack.Artist, wr.Tracks[i].Artist)
					//fmt.Printf("%d | %d\n", drmTrack.DurationSeconds, wr.Tracks[i].DurationSeconds)
					//fmt.Printf("%s | %s\n", drmTrack.Album, wr.Tracks[i].Album)
					//fmt.Printf("%s | %s\n", drmTrack.AlbumArtist, wr.Tracks[i].AlbumArtist)
					//fmt.Printf("%d | %d\n", drmTrack.TrackNumber, wr.Tracks[i].TrackNumber)
					//fmt.Printf("%d | %d\n", drmTrack.TotalTracks, wr.Tracks[i].TotalTracks)
					//fmt.Printf("%s\n", wr.Tracks[i].Path)
					// Is good enough to hold onto but keep looking
					bestScore = score
					bestIndex = i
				}
			}
		}

		// we're only taking over 85% on this. Might be too low for this type of match?
		if bestScore > 0.85 {
			fmr := common.FileMovedResult{
				Source: drmPath,
				Dest:   wr.Tracks[bestIndex].Path,
			}
			replacedReportResult.Moved = append(replacedReportResult.Moved, fmr)
			replacedReportResult.Matches = append(replacedReportResult.Matches,
				common.FmtTrackMatch(drmTrack, wr.Tracks[bestIndex], bestScore, true))
			continue OUTER
		} else if bestScore > 0.5 {
			replacedReportResult.Matches = append(replacedReportResult.Matches,
				common.FmtTrackMatch(drmTrack, wr.Tracks[bestIndex], bestScore, false))
		}

		// if we're here, there is no possible match where the title is exactly the same
		// TODO is it worth it to find similar titles using binarySearch on the Tracks slice?
	}

	// Sort the matches from best to worst by score
	slices.SortFunc(replacedReportResult.Matches, common.CmpTrackMatchScore)

	if !isDryRun {
		for _, m := range replacedReportResult.Moved {
			// delete the source file (the encrypted one in this case)
			os.Remove(m.Source)
			fmt.Printf("- %s\n", m.Source)
		}
	}

	j, _ := json.MarshalIndent(&replacedReportResult, "", "  ")
	jsonString := string(j)
	if isDryRun {
		fmt.Println(jsonString)
	} else {
		// save the report without overwriting
		// create a destination for the report
		reportPath := filepath.Join(outputDir, "replaced.json")
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
