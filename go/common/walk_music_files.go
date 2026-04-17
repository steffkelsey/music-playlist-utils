package common

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type AlbumInfo struct {
	Album         string      `json:"album"`
	Artist        string      `json:"artist"`
	Tracks        []TrackInfo `json:"tracks"`
	TotalDiscs    int         `json:"totalDiscs"`
	TotalTracks   int         `json:"totalTracks"`
	AccurateCount bool        `json:"accurateCount"`
}

func (a *AlbumInfo) FlattenTrackNumbers() {
	// map for discNumber to totalTracks for that disc
	m := make(map[int]int)
	// iterate over the tracks
	for index, t := range a.Tracks {
		// save the total tracks for this disc number if needed
		if _, ok := m[t.DiscNumber]; !ok {
			m[t.DiscNumber] = t.TotalTracks
		}
		n := t.TrackNumber
		for i := range t.DiscNumber {
			if tot, ok := m[i]; ok {
				n += tot
			}
		}
		a.Tracks[index].TrackNumber = n
	}
}

// This will only work on albums where the Tracks are sorted
// by disc number and then track number
func (a *AlbumInfo) IsComplete() bool {
	for i := 1; i <= a.TotalDiscs; i++ {
		if !a.IsDiscComplete(i) {
			return false
		}
	}
	return true
}

func (a *AlbumInfo) CalcTotalTracks() bool {
	trackCount := 0
	foundAllDiscs := true
	// For the total number of possible discs
	for i := 1; i <= a.TotalDiscs; i++ {
		j, found := slices.BinarySearchFunc(a.Tracks,
			TrackInfo{DiscNumber: i},
			CmpTrackInfoDiscNum,
		)
		if found {
			// get the totalTracks for this disc off of the returned track index
			trackCount += a.Tracks[j].TotalTracks
		} else {
			foundAllDiscs = false
		}
	}
	// sum of totalTracks for each disc is the album totaltracks
	a.TotalTracks = trackCount
	a.AccurateCount = foundAllDiscs
	return foundAllDiscs
}

// This will only work on albums where the Tracks are sorted
// by disc number and then track number
func (a *AlbumInfo) IsDiscComplete(n int) bool {
	i, found := slices.BinarySearchFunc(a.Tracks,
		TrackInfo{DiscNumber: n},
		CmpTrackInfoDiscNum,
	)
	if found {
		// we now have the starting track index for this disc
		// we need the ending index
		// if this is the final disc, we can use the length of the Tracks slice
		// if not, we need to know the index of the first track of the next disc
		if n < a.TotalDiscs {
			j, f := slices.BinarySearchFunc(a.Tracks,
				TrackInfo{DiscNumber: n + 1},
				CmpTrackInfoDiscNum,
			)
			if f {
				return a.Tracks[i].TotalTracks == j-i
			}
		} else {
			return a.Tracks[i].TotalTracks == len(a.Tracks)-i
		}
	}

	return false
}

func (a *AlbumInfo) GetKey() string {
	return StripToughToMatchChars(strings.ToLower(fmt.Sprintf("%s|%s", a.Artist, a.Album)))
}

type TrackInfo struct {
	Path            string `json:"-"`
	Title           string `json:"title"`
	Artist          string `json:"artist"`
	DiscNumber      int    `json:"discNumber"`
	TotalDiscs      int    `json:"totalDiscs"`
	TrackNumber     int    `json:"trackNumber"`
	TotalTracks     int    `json:"totalTracks"`
	Album           string `json:"album"`
	AlbumArtist     string `json:"albumArtist"`
	DurationSeconds int    `json:"durationSeconds"`
}

func (t TrackInfo) GetAlbumKey() string {
	return StripToughToMatchChars(strings.ToLower(fmt.Sprintf("%s|%s", t.AlbumArtist, t.Album)))
}

func (t TrackInfo) GetKey() string {
	return StripToughToMatchChars(strings.ToLower(fmt.Sprintf("%s|%s", t.Artist, t.Title)))
}

type WalkResults struct {
	Count                     int64
	MapSizeStringSlices       map[int64][]string
	Files                     []string
	RootPath                  string
	Albums                    []AlbumInfo
	Tracks                    []TrackInfo
	AlbumArtistBarNameToIndex map[string]int
	TrackPathToIndex          map[string]int
}

func WalkAllMusicFiles(folder string, processFunc func(path string, info fs.FileInfo, results *WalkResults) error) (WalkResults, error) {
	results := WalkResults{
		Count:                     0,
		MapSizeStringSlices:       make(map[int64][]string),
		Files:                     make([]string, 0),
		RootPath:                  folder,
		Albums:                    make([]AlbumInfo, 0),
		Tracks:                    make([]TrackInfo, 0),
		AlbumArtistBarNameToIndex: make(map[string]int),
		TrackPathToIndex:          make(map[string]int),
	}

	err := filepath.Walk(folder, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error accessing path %s: %v\n", path, err)
			return nil // Continue walking
		}

		// Skip directories
		if info.IsDir() && path != folder {
			return nil
		}

		// Skip the root folder itself
		if path == folder {
			return nil
		}

		// If it's an music file, run the processFunc
		if IsMusicFile(path) {
			err = processFunc(path, info, &results)
			if err != nil {
				fmt.Printf("Error on file at in processFunc at path: %s\n, err: %v", path, err)
			}
			// TODO return the error? Does that stop walking?
			return err
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking through folder: %v\n", err)
		return results, err
	}

	return results, nil
}

func WalkAllMusicFilesNotRecursive(folder string, processFunc func(path string, info fs.FileInfo, results *WalkResults) error) (WalkResults, error) {
	results := WalkResults{
		Count:                     0,
		MapSizeStringSlices:       make(map[int64][]string),
		Files:                     make([]string, 0),
		RootPath:                  folder,
		Albums:                    make([]AlbumInfo, 0),
		Tracks:                    make([]TrackInfo, 0),
		AlbumArtistBarNameToIndex: make(map[string]int),
		TrackPathToIndex:          make(map[string]int),
	}

	entries, err := os.ReadDir(folder)
	if err != nil {
		return results, err
	}

	for _, entry := range entries {
		// Skip subdirectories
		if entry.IsDir() {
			continue
		}

		path := filepath.Join(folder, entry.Name())
		// only act on music files by extension
		if IsMusicFile(path) {
			info, err := entry.Info()
			if err != nil {
				fmt.Printf("error getting FileInfo: %v\n", err)
			}
			err = processFunc(path, info, &results)
			if err != nil {
				fmt.Printf("Error on file at in processFunc at path: %s\n, err: %v", path, err)
			}
		}
	}
	return results, err
}
