package cmd

import (
	"cmp"
	"encoding/json"
	"fmt"
	"io/fs"
	"math"
	"path/filepath"
	"slices"
	"strconv"

	"github.com/alitto/pond/v2"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	"music-utils/common"
)

type trackInfoExtended struct {
	AlbumTrackPercentage float64 `json:"albumTrackPercentage"`
	common.TrackInfo
}

type duplicatesReport struct {
	Duplicates []duplicateResult `json:"duplicates"`
	movedReport
}

type duplicateResult struct {
	Keep   []trackInfoExtended `json:"keep"`
	Delete []trackInfoExtended `json:"delete"`
}

var duplicatesCmd = &cobra.Command{
	Use:   "duplicates",
	Short: "Finds duplicate music files",
	Long: `Finds duplicate music files in the given
input folder. To find duplicates recursively in ~/Music/tmp 
and save the report in ~/Music/reports:

music-utils duplicates -i $HOME/Music/tmp -o $HOME/Music/Reports`,
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
		if !common.IsFFProbeInstalled() {
			return fmt.Errorf("ffprobe is not installed")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return findDuplicateFiles(inputDir, isDryRun)
	},
}

func init() {
	rootCmd.AddCommand(duplicatesCmd)
}

func findDuplicateFiles(rootPath string, isDryRun bool) error {
	wr, err := common.WalkAllMusicFiles(rootPath, collateSizeAlbumAndTracks)
	if err != nil {
		return err
	}

	dupesToRank := make([][]common.TrackInfo, 0)
	mapTrackIndexToDupesToRankIndex := make(map[int]int)

	// Create a pool with a result type of string
	pool := pond.NewResultPool[string](100)

	// Create a task group
	group := pool.NewGroup()

	// Sort the Tracks slice by title asc
	slices.SortFunc(wr.Tracks, common.CmpTrackInfoTitle)
	for i, t := range wr.Tracks {
		// set the now sorted index in the path map
		wr.TrackPathToIndex[t.Path] = i

		group.SubmitErr(func() (string, error) {
			fmt.Printf("\rgetting duration of all tracks! %d%% complete...", 100*i/len(wr.Tracks))
			return common.GetDurationString(t.Path)
		})
	}

	// wait for all the responses to complete
	responses, err := group.Wait()

	if err != nil {
		fmt.Printf("Failed to get duration and bitrate data: %v", err)
		return err
	}

	// write a new line
	fmt.Println()

	for _, str := range responses {
		d := []byte(str)
		var f common.FFProbeFormatResponse
		// unmarshal the response data
		err = json.Unmarshal(d, &f)
		if err != nil {
			fmt.Printf("error unmarshalling json.%v\n", err)
			continue
		}
		// look up the index using the path
		index, ok := wr.TrackPathToIndex[f.Format.Filename]
		if !ok {
			fmt.Printf("error finding track index at path %s\n", f.Format.Filename)
			continue
		}
		// set the duration on the track at the index
		durFl, _ := strconv.ParseFloat(f.Format.Duration, 64)
		wr.Tracks[index].DurationSeconds = int(durFl)
		// set the bitrate on the track at the index
		wr.Tracks[index].BitRate, _ = strconv.Atoi(f.Format.BitRate)
	}

	// Sort the Albums slice by title asc
	slices.SortFunc(wr.Albums, common.CmpAlbumInfoAlbumTitle)
	// Sort the tracks in each AlbumInfo by Disc Number then Track Number
	for i, album := range wr.Albums {
		slices.SortFunc(album.Tracks, common.CmpTrackInfoDiscAndTrackNum)
		// add up the total tracks for all discs in the album
		wr.Albums[i].CalcTotalTracks()
	}
	// re-init the map for searching on albumArtist|title for AlbumInfo
	wr.AlbumArtistBarNameToIndex = make(map[string]int)
	for i, a := range wr.Albums {
		k := a.GetExactKey()
		_, ok := wr.AlbumArtistBarNameToIndex[k]
		if !ok {
			wr.AlbumArtistBarNameToIndex[k] = i
		} else {
			fmt.Printf("album key clash at %s\n", k)
		}
	}

	// deal with the tracks that have exactly the same file size
	for _, v := range wr.MapSizeStringSlices {
		// v is an array of paths that have the same file size
		// only look at the ones where the array has more than 1 path
		if len(v) > 1 {
			tracks := make([]common.TrackInfo, 0)
			// convert to an array of Tracks
			for _, p := range v {
				i, ok := wr.TrackPathToIndex[p]
				if ok {
					tracks = append(tracks, wr.Tracks[i])
				} else {
					fmt.Printf("error finding track index at path %s\n", p)
				}
			}

			fmt.Printf("comparing %+v\n", tracks)

			// iterate over the array, comparing the tracks to each other
			// use a cutoff of 1.0 so all tracks in the slice are compared
			for k := range tracks {
				// must find the index in the wr.Tracks slice
				i := wr.TrackPathToIndex[tracks[k].Path]
				// Has the track at this index in wr.Tracks already been matched?
				_, matched := mapTrackIndexToDupesToRankIndex[i]
				if matched {
					continue
				}
				matchingIndexes := returnAllTracksWithAlbumInfoOverThresholdExcluding(tracks, k, 1.0, 0.87)
				if len(matchingIndexes) > 0 {
					// create an slice of TrackInfo objects to be ranked
					a := []common.TrackInfo{wr.Tracks[i]}
					newDupesToRankIndex := len(dupesToRank)
					for z := range matchingIndexes {
						a = append(a, tracks[z])
						// find the index in the wr.Tracks slice
						j := wr.TrackPathToIndex[tracks[z].Path]
						// mark that the indexes have duplicates
						mapTrackIndexToDupesToRankIndex[j] = newDupesToRankIndex
					}
					dupesToRank = append(dupesToRank, a)
				}
			}
		}
	}

	// #1 Do I have an album match without knowing it because of a two
	// different spellings of the title (or different special character?)
	//for i := range wr.Albums {
	//	matchIndex, matchInfo := returnBestAlbumExcluding(wr.Albums, i, 0.7, 0.95)
	//	if matchIndex > -1 {
	//		// TODO If yes, then combine the albums and move on, dealing with the tracks later
	//		fmt.Printf("MATCH! %+v\n", matchInfo)
	//	}
	//}

	// Now we have all the sorted albums in a slice with tracks in order
	// AND all the sorted tracks in a slice.

	for i := range wr.Tracks {
		// Has the track at this index already been matched?
		_, matched := mapTrackIndexToDupesToRankIndex[i]
		if matched {
			continue
		}
		matchingIndexes := returnAllTracksWithAlbumInfoOverThresholdExcluding(wr.Tracks, i, 0.50, 0.87)
		if len(matchingIndexes) > 0 {
			// create an slice of indexes to be ranked
			a := []common.TrackInfo{wr.Tracks[i]}
			newDupesToRankIndex := len(dupesToRank)
			for _, z := range matchingIndexes {
				a = append(a, wr.Tracks[z])
				// mark that the indexes have duplicates
				mapTrackIndexToDupesToRankIndex[z] = newDupesToRankIndex
			}
			dupesToRank = append(dupesToRank, a)
		}
	}

	// create arrays of trackInfoExtended from the dupesToRank arrays of wr.Tracks indexes
	// calculate the percentage of tracks that are present in each album
	// get the BitRate for each track while we are at it
	dupesToRankExt := make([][]trackInfoExtended, 0)

	for _, a := range dupesToRank {
		out := make([]trackInfoExtended, 0)
		for _, t := range a {
			et := trackInfoExtended{
				TrackInfo: t,
			}
			// get the bitrate and updated duration using the pool
			group.SubmitErr(func() (string, error) {
				return common.FFProbeForString(t.Path)
			})

			// Get the index to the album in wr.Albums using the exactAlbumKey
			ak := t.GetExactAlbumKey()
			albumIndex, ok := wr.AlbumArtistBarNameToIndex[ak]
			if !ok {
				fmt.Printf("error finding albumInfo index with key %s\n", ak)
			} else {
				if wr.Albums[albumIndex].TotalTracks > 0 {
					et.AlbumTrackPercentage = float64(len(wr.Albums[albumIndex].Tracks)) / float64(wr.Albums[albumIndex].TotalTracks)
				}
			}

			out = append(out, et)
		}
		dupesToRankExt = append(dupesToRankExt, out)
	}

	// wait for all the responses to complete
	responses, err = group.Wait()

	if err != nil {
		fmt.Printf("Failed to get duration and bitrate data: %v", err)
		return err
	}

	for _, str := range responses {
		d := []byte(str)
		var f common.FFProbeFormatResponse
		// unmarshal the response data
		err = json.Unmarshal(d, &f)
		if err != nil {
			fmt.Printf("error unmarshalling json.%v\n", err)
			continue
		}
		// look up the index using the path
		index, ok := wr.TrackPathToIndex[f.Format.Filename]
		if !ok {
			fmt.Printf("error finding track index at path %s\n", f.Format.Filename)
			continue
		}
		// get the duration float from the string
		df, _ := strconv.ParseFloat(f.Format.Duration, 64)
		// set the duration on the track at the index
		wr.Tracks[index].DurationSeconds = int(df)
		// set the bitrate on the track at the index
		brf, _ := strconv.ParseFloat(f.Format.BitRate, 64)
		// round to nearest 10 kbps but show as kbps
		brf = 10.0 * math.Round(brf/10000.0)
		// set the bitrate on the track at the index
		wr.Tracks[index].BitRate = int(brf)
	}

	// we need to add the updated bitrate and duration before we rank
	for i, a := range dupesToRankExt {
		// updates for each trackInfoExt
		for k, et := range a {
			// look up the wr.Tracks index using the path
			index, ok := wr.TrackPathToIndex[et.Path]
			if !ok {
				fmt.Printf("error finding track index at path %s\n", et.Path)
				continue
			}
			dupesToRankExt[i][k].DurationSeconds = wr.Tracks[index].DurationSeconds
			dupesToRankExt[i][k].BitRate = wr.Tracks[index].BitRate
		}
	}

	report := duplicatesReport{
		Duplicates:  make([]duplicateResult, 0),
		movedReport: movedReport{Moved: make([]common.FileMovedResult, 0)},
	}
	act := common.Continue
loop:
	for _, v := range dupesToRankExt {
		// create a duplicateResult struct
		d := rankDuplicates(v)
		if !isDryRun {
			// Only prompt the user if something will be deleted.
			// Otherwise there is not a meaningful decision to make.
			if len(d.Delete) > 0 {
				switch act {
				case common.Continue:
					act, err = promptAndMaybeDelete(&d)
					if err != nil {
						fmt.Printf("Error deleting file: %v", err)
						return err
					}
					fmt.Println()
				case common.ConfirmAll:
					// delete the files without asking
					common.DeleteFiles(getPathsToDelete(d.Delete))
				case common.Abort:
					break loop
				}
			}
		}
		// Only write it out if one of the slices has paths
		if len(d.Keep) > 0 || len(d.Delete) > 0 {
			report.Duplicates = append(report.Duplicates, d)
		}
		if len(d.Keep) > 0 && len(d.Delete) > 0 {
			// For each deleted file create a FileMovedResult
			// where the source is a deleted file path and the destination
			// is one of the keep file paths
			for _, p := range d.Delete {
				m := common.FileMovedResult{
					Source: p.Path,
					Dest:   d.Keep[0].Path,
				}
				report.Moved = append(report.Moved, m)
			}
		}
	}

	// marshal the report to []byte
	j, _ := json.MarshalIndent(&report, "", "  ")
	if isDryRun {
		fmt.Println(string(j))
	} else {
		// create a destination for the report
		reportPath := filepath.Join(outputDir, "deleted-duplicates.json")
		// We don't want to overwrite reports, so make sure the path is unique
		reportPath = common.FindFileNameNoOverWrite(reportPath)
		// ask if they want to save the json report and save it if so
		msg := fmt.Sprintf(`Save report at:
%s		
`, reportPath)
		didSave, err := common.PromptAndMaybeSaveFile(reportPath, j, msg)
		if err != nil {
			fmt.Printf("Error writing json report, %v\n", err)
			return err
		}
		if didSave {
			fmt.Println("Report saved")
		}
	}
	return nil
}

func collateSizeAlbumAndTracks(path string, info fs.FileInfo, results *common.WalkResults) error {
	// Ignore encrypted files and playlist files
	if common.IsEncryptedFile(path) || common.IsPlaylistFile(path) {
		return nil
	}
	// Look for files with the same file size first (extremely naive)
	results.MapSizeStringSlices[info.Size()] = append(results.MapSizeStringSlices[info.Size()], path)

	// Build out a model of the tracks into Track and Albums
	ok, track, _ := common.CreateTrackInfoFromPath(path)
	if !ok {
		return nil
	} else {
		// Set the map of path to index in the Tracks slice
		results.TrackPathToIndex[path] = len(results.Tracks)
		// append the track
		results.Tracks = append(results.Tracks, track)
		// see if the album already exists in the Albums slice
		i, ok := results.AlbumArtistBarNameToIndex[track.GetExactAlbumKey()]
		if ok {
			// add the TrackInfo to the AlbumInfo.Tracks slice
			results.Albums[i].Tracks = append(results.Albums[i].Tracks, track)
		} else {
			// save the index where we added the album into the name map
			results.AlbumArtistBarNameToIndex[track.GetExactAlbumKey()] = len(results.Albums)
			a := []common.TrackInfo{track}
			// create the new album in the results
			results.Albums = append(results.Albums, common.AlbumInfo{
				Album:      track.Album,
				Artist:     track.AlbumArtist,
				TotalDiscs: track.TotalDiscs,
				Tracks:     a,
			})
		}
	}

	return nil
}

func returnAllTracksWithAlbumInfoOverThresholdExcluding(tracks []common.TrackInfo, i int, cutoff float64, threshold float64) []int {
	// set a default cutoff of 0.5
	if cutoff-0.001 < 0 {
		cutoff = 0.5
	}
	// set a default threshold of 0.85
	if threshold-0.001 < 0 {
		threshold = 0.85
	}

	// threshold for comparing tracks WITHOUT album context
	// eg: title, artist, duration
	alt_threshold := threshold - 0.1

	matchingIndexes := make([]int, 0)

	cnt := 0
	// start just before i and walk down
	score := 1.0
	j := common.MaxInt(0, i-1)
	for score > cutoff {
		if j == i {
			break
		}
		score = common.CmpTracksWithAlbumInfo(tracks[i], tracks[j])
		if score > threshold {
			matchingIndexes = append(matchingIndexes, j)
		} else {
			alt_score := common.CmpTracks(tracks[i], tracks[j])
			if alt_score > alt_threshold {
				matchingIndexes = append(matchingIndexes, j)
			}
		}

		j--
		if j < 0 {
			break
		}
		cnt++
	}

	// start just after i and walk up
	score = 1.0
	j = common.MinInt(len(tracks)-1, i+1)
	for score > cutoff {
		if j == i {
			break
		}
		score = common.CmpTracksWithAlbumInfo(tracks[i], tracks[j])
		if score > threshold {
			matchingIndexes = append(matchingIndexes, j)
		} else {
			alt_score := common.CmpTracks(tracks[i], tracks[j])
			if alt_score > alt_threshold {
				matchingIndexes = append(matchingIndexes, j)
			}
		}

		j++
		if j > len(tracks)-1 {
			break
		}
		cnt++
	}

	//if len(matchingIndexes) > 0 {
	//	fmt.Printf("checked %d indexes to find %d matches\n", cnt, len(matchingIndexes))
	//}
	return matchingIndexes
}

// Duplicate rules
// A duplicate is when we have the same recording
// detected when track title, track artist, and duration are very good matches
// exact match of file size is also a naive indicator
// When we have duplicates, the times we want to delete are:
//  1. duplicate tracks that exist on the same album
//  2. duplicate tracks on two sparse albums
//  3. duplicate tracks where one album is sparse
//     eg: we have Billy Joel - Piano Man from the album 'Piano Man' with all other tracks
//     and Billy Joel - Piano Man from the album 'Greatest Hits of the 70s' with only 2 tracks
//  4. TODO duplicate tracks where neither album is sparse but one album is just the single (one or two tracks total). Left this at the moment because it seemed risky
//
// Rules for keeping:
// 1.Quality first!
//
//		If one track has a higher bandwidth
//		Calculate bandwidth kbit/s (kilobits per second) as (size on disk * 8000) / (duration seconds)
//
//	 2. Artists releases first!
//	    Intact Artist Albums over multi-artist Albums (Various Artists or Hits Compilations/Soundtracks)
//	 3. Keep albums whole! If we have a 2 mostly full albums with a duplicate track,
//	    leave both tracks in place. Most likely scenarios are with movie soundtracks
//
// BUT we should always shift to keeping Artist Albums intact
// So if the higher bandwidth track is part of
// We also might have poorly tagged tracks where the Album, track title,
// artist, or album artist was tagged with different spelling or
// different special characters and we want to put them together correctly.
func rankDuplicates(dupesToRank []trackInfoExtended) duplicateResult {
	r := duplicateResult{
		Keep:   make([]trackInfoExtended, 0),
		Delete: make([]trackInfoExtended, 0),
	}
	// sort by bitrate
	// and then by albumTrackPercentage proximity to 100
	// and then by filename score
	slices.SortFunc(dupesToRank, cmpTrackInfoExtendedBitRateAndTrackPercentage)

	// Pop off the top one to keep
	r.Keep = append(r.Keep, dupesToRank[0])
	// Remove it by reslicing
	dupesToRank = dupesToRank[1:]

	// iterate over the remaining ones and make decisions
	for len(dupesToRank) > 0 {
		keepTrack := r.Keep[0]
		// set a current index
		i := 0
		t := dupesToRank[i]
		// is this track from the same album as the keepTrack?
		if keepTrack.GetExactAlbumKey() == t.GetExactAlbumKey() {
			// same disc and track number?
			if keepTrack.DiscNumber == t.DiscNumber && keepTrack.TrackNumber == t.TrackNumber {
				// delete this track!
				r.Delete = append(r.Delete, t)
				dupesToRank = append(dupesToRank[:i], dupesToRank[i+1:]...)
				continue
			} else {
				// keep it!
				r.Keep = append(r.Keep, t)
				dupesToRank = append(dupesToRank[:i], dupesToRank[i+1:]...)
				continue
			}
		} else {
			// different albums
			// are both tracks from sparse albums?
			if t.AlbumTrackPercentage < 0.5 && keepTrack.AlbumTrackPercentage < 0.5 {
				if t.BitRate >= keepTrack.BitRate && t.AlbumTrackPercentage > keepTrack.AlbumTrackPercentage {
					// delete the keep track
					r.Delete = append(r.Delete, keepTrack)
					// move the current track to keep
					r.Keep[0] = t
					dupesToRank = append(dupesToRank[:i], dupesToRank[i+1:]...)
				} else {
					// delete the current track
					r.Delete = append(r.Delete, t)
					dupesToRank = append(dupesToRank[:i], dupesToRank[i+1:]...)
				}
				continue
			}
			// are both tracks from non sparse albums
			if t.AlbumTrackPercentage >= 0.5 && keepTrack.AlbumTrackPercentage >= 0.5 {
				// keep the current track
				r.Keep = append(r.Keep, t)
				dupesToRank = append(dupesToRank[:i], dupesToRank[i+1:]...)
				continue
			}
			// is the current track way MORE sparse than the keep track?
			if keepTrack.AlbumTrackPercentage > t.AlbumTrackPercentage && (keepTrack.AlbumTrackPercentage-t.AlbumTrackPercentage > 0.5 || t.AlbumTrackPercentage < 0.1) {
				// delete the current track
				r.Delete = append(r.Delete, t)
				dupesToRank = append(dupesToRank[:i], dupesToRank[i+1:]...)
				continue
			}
			// is current track from a much less sparse album than the keep track?
			if t.AlbumTrackPercentage > keepTrack.AlbumTrackPercentage && (t.AlbumTrackPercentage-keepTrack.AlbumTrackPercentage > 0.5 || keepTrack.AlbumTrackPercentage < 0.1) {
				// delete the keep track
				r.Delete = append(r.Delete, keepTrack)
				// move the current track to keep
				r.Keep[0] = t
				dupesToRank = append(dupesToRank[:i], dupesToRank[i+1:]...)
				continue
			}
		}
		// we fell through everything, just keep the current track
		r.Keep = append(r.Keep, t)
		dupesToRank = append(dupesToRank[:i], dupesToRank[i+1:]...)
	}

	return r
}

// for sorting/searching where we want to use bitrate as primary sort
// and proximity to 100% albumTrackPercentage
// and filename score (does it end in (#)
func cmpTrackInfoExtendedBitRateAndTrackPercentage(a, b trackInfoExtended) int {
	// Sort by bitrate first
	if n := -1 * cmp.Compare(a.BitRate, b.BitRate); n != 0 {
		return n
	}
	// If bitrate is equal, compare by how close the album track percentage is to 1.0
	if n := cmp.Compare(math.Abs(a.AlbumTrackPercentage-1.0), math.Abs(b.AlbumTrackPercentage-1.0)); n != 0 {
		return n
	}

	// If bitrate and album track percentage diff from 1.0 are equal,
	// then go by filename score. A filename with file(1).opus or
	// file(2).opus is worse than file.opus because it indicates
	// a duplicate from our organize cmd.
	return -1 * cmp.Compare(a.FilenameScore(), b.FilenameScore())
}

func promptAndMaybeDelete(d *duplicateResult) (int, error) {
	// show the keep and delete tracks as formatted json
	keepj, _ := json.MarshalIndent(&d.Keep, "", "  ")
	keepstr := string(keepj)
	deletej, _ := json.MarshalIndent(&d.Delete, "", "  ")
	deletestr := string(deletej)
	fmt.Printf(`Keep:
%s	

Delete:
%s

`,
		keepstr,
		deletestr)
	prompt := promptui.Select{
		Label: "Delete?",
		Items: []string{"Yes", "No", "Confirm All. Don't Ask Again", "Quit"},
	}

	_, result, err := prompt.Run()

	if err != nil {
		// User hit ctrl-c or something
		return common.Abort, err
	}

	switch result {
	case "No":
		fmt.Printf("\nSkipping...\n")
		// Empty the result so it is not recorded
		d.Keep = []trackInfoExtended{}
		d.Delete = []trackInfoExtended{}
	case "Yes":
		err = common.DeleteFiles(getPathsToDelete(d.Delete))
	case "Confirm All. Don't Ask Again":
		// Delete the files we just presented to the user
		err = common.DeleteFiles(getPathsToDelete(d.Delete))
		// Send back that all future deletions are confirmed
		return common.ConfirmAll, err
	case "Quit":
		// Empty the result so it is not recorded
		d.Keep = []trackInfoExtended{}
		d.Delete = []trackInfoExtended{}
		return common.Abort, nil
	}
	return common.Continue, err
}

func getPathsToDelete(tracks []trackInfoExtended) []string {
	pathsToDelete := make([]string, 0)
	for _, t := range tracks {
		pathsToDelete = append(pathsToDelete, t.Path)
	}
	return pathsToDelete
}
