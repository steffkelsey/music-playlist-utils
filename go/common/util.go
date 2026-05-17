package common

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unsafe"

	"github.com/lizc2003/audioduration"
)

func Bool2Float(b bool) float64 {
	return float64(Bool2int(b))
}

func Bool2int(b bool) int {
	return int(*(*byte)(unsafe.Pointer(&b)))
}

func MaxInt(a, b int) int {
	return int(math.Max(float64(a), float64(b)))
}

func MinInt(a, b int) int {
	return int(math.Min(float64(a), float64(b)))
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
		fallthrough
	case ExtMp4:
		return audioduration.Duration(f, audioduration.TypeMp4)
	case ExtOpus:
		// The audioduration lib does not seem to support Opus at this time.
		// Tried an OGG container, but got all negative values.
		return 0.0, fmt.Errorf("audioduration lib does not support OpusOgg")
	}

	return 0.0, fmt.Errorf("cannot find duration for that filetype")
}

func GetDurationAndBitRate(path string) (float64, int, error) {
	var duration float64
	var bitRate int

	d, err := ffprobe(path)
	if err != nil {
		return duration, bitRate, err
	}

	var f FFProbeFormatResponse
	// unmarshal the response data
	err = json.Unmarshal(d, &f)
	if err != nil {
		return duration, bitRate, err
	}
	duration, _ = strconv.ParseFloat(f.Format.Duration, 64)
	bitRate, _ = strconv.Atoi(f.Format.BitRate)

	return duration, bitRate, nil
}

// Uses the fast audioduration lib and falls back to the slower ffprobe method
// on error. Bundled to return a string in the format of FFProbeFormatResponse.
func GetDurationString(path string) (string, error) {
	var response string
	// try the fast method
	d, err := GetDuration(path)
	if err != nil {
		// use FFProbe
		response, err = FFProbeForString(path)
		if err != nil {
			return response, err
		}
	} else {
		// put the duration into a FFProbeFromatResponse
		r := FFProbeFormatResponse{
			Format: FFProbeDurationAndBitRateResponse{
				Duration: fmt.Sprintf("%.2f", d),
				Filename: path,
			},
		}
		// marshal into the json
		j, _ := json.Marshal(&r)
		// convert to string
		response = string(j)
	}
	return response, nil
}

func FmtAlbumMatch(a1, a2 AlbumInfo, score float64, success bool) AlbumMatch {
	return AlbumMatch{
		Score:       score,
		Titles:      fmt.Sprintf("%s | %s", a1.Album, a2.Album),
		Artists:     fmt.Sprintf("%s | %s", a1.Artist, a2.Artist),
		TotalDiscs:  fmt.Sprintf("%d | %d", a1.TotalDiscs, a2.TotalDiscs),
		TotalTracks: fmt.Sprintf("%d | %d", a1.TotalTracks, a2.TotalTracks),
		Success:     success,
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

// StripToughToMatchChars removes different versions of single
// and double quotes. Best used when making map keys
func StripToughToMatchChars(s string) string {
	re := regexp.MustCompile(`(\p{Pi}|\p{Pf}|'|"){1}`)
	return string(re.ReplaceAll([]byte(s), []byte("")))
}
