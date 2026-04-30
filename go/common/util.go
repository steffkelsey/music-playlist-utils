package common

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"github.com/hcl/audioduration"
)

func Bool2Float(b bool) float64 {
	return float64(Bool2int(b))
}

func Bool2int(b bool) int {
	return int(*(*byte)(unsafe.Pointer(&b)))
}

func GetDuration(path string) (float64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0.0, err
	}
	defer f.Close()
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ExtMp3:
		return audioduration.Duration(f, audioduration.TypeMp3)
	case ExtM4a:
		return audioduration.Duration(f, audioduration.TypeMp4)
	case ExtMp4:
		return audioduration.Duration(f, audioduration.TypeMp4)
	}

	return 0.0, fmt.Errorf("cannot find duration for that filetype")
}

func FmtAlbumMatch(a1, a2 AlbumInfo, score float64, success bool) AlbumMatch {
	return AlbumMatch{
		Score:      score,
		Titles:     fmt.Sprintf("%s | %s", a1.Album, a2.Album),
		Artists:    fmt.Sprintf("%s | %s", a1.Artist, a2.Artist),
		TotalDiscs: fmt.Sprintf("%d | %d", a1.TotalDiscs, a2.TotalDiscs),
		Success:    success,
	}
}

func FmtTrackMatch(t1, t2 TrackInfo, score float64, success bool) TrackMatch {
	return TrackMatch{
		TrackPaths:   []string{t1.Path, t2.Path},
		Score:        score,
		Titles:       fmt.Sprintf("%s | %s", t1.Title, t2.Title),
		Artists:      fmt.Sprintf("%s | %s", t1.Artist, t2.Artist),
		Albums:       fmt.Sprintf("%s | %s", t1.Album, t2.Album),
		AlbumArtists: fmt.Sprintf("%s | %s", t1.AlbumArtist, t2.AlbumArtist),
		Durations:    fmt.Sprintf("%d | %d", t1.DurationSeconds, t2.DurationSeconds),
		TrackNumbers: fmt.Sprintf("%d | %d", t1.TrackNumber, t2.TrackNumber),
		TotalTracks:  fmt.Sprintf("%d | %d", t1.TotalTracks, t2.TotalTracks),
		DiscNumbers:  fmt.Sprintf("%d | %d", t1.DiscNumber, t2.DiscNumber),
		TotalDiscs:   fmt.Sprintf("%d | %d", t1.TotalDiscs, t2.TotalDiscs),
		Success:      success,
	}
}
