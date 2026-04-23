package common

import (
	"os"

	"github.com/dhowden/tag"
)

func CreateTrackInfoFromPath(path string) (bool, TrackInfo, []string) {
	r := make([]string, 0)
	// Open the file to get more details
	file, err := os.Open(path)
	if err != nil {
		r = append(r, "Could not open file")
		return false, TrackInfo{}, r
	} else {
		defer file.Close()

		// Use dhowden/tag to read metadata
		m, err := tag.ReadFrom(file)
		if err != nil {
			r = append(r, "Could not read tags")
			return false, TrackInfo{}, r
		}
		t := CreateTrackInfoFromTags(path, m)
		isTagGood, r := t.IsGood()
		return isTagGood, t, r
	}
}

func CreateTrackInfoFromTags(path string, m tag.Metadata) TrackInfo {
	t := TrackInfo{
		Path:        path,
		Title:       m.Title(),
		Artist:      m.Artist(),
		Album:       m.Album(),
		AlbumArtist: m.AlbumArtist(),
	}

	t.TrackNumber, t.TotalTracks = m.Track()

	return t
}

func (t TrackInfo) IsGood() (bool, []string) {
	r := make([]string, 0)
	isTagGood := true

	// Must have Album, Title, TrackNumber, Artist
	if t.Album == "" {
		r = append(r, "Missing Album tag")
		isTagGood = false
	}

	if t.Title == "" {
		r = append(r, "Missing Title tag")
		isTagGood = false
	}

	// This is Track artist NOT album artist
	if t.Artist == "" {
		r = append(r, "Missing Artist tag")
		isTagGood = false
	}

	if t.TrackNumber == 0 {
		r = append(r, "Missing Track Number tag")
		isTagGood = false
	}

	return isTagGood, r
}
