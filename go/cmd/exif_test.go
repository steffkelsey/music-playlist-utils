package cmd

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestExifTrackToTrackInfo(t *testing.T) {
	c := qt.New(t)
	tests := []struct {
		input    exifTrack
		expected trackInfo
	}{
		{
			exifTrack{
				Title:       "track 1",
				Artist:      "artist 1",
				AlbumArtist: "88 artists",
				Album:       "artist 1 live",
				TrackNumber: "1 of 18",
			},
			trackInfo{
				Title:       "track 1",
				Artist:      "artist 1",
				TrackNumber: "1",
				TotalTracks: "18",
				Album:       "artist 1 live",
			},
		},
		{
			exifTrack{
				Title:       "track 19",
				Artist:      "The Boss",
				AlbumArtist: "The Boss and the Band",
				Album:       "That Sad One",
				TrackNumber: "19 of 19",
			},
			trackInfo{
				Title:       "track 19",
				Artist:      "The Boss",
				TrackNumber: "19",
				TotalTracks: "19",
				Album:       "That Sad One",
			},
		},
		{
			exifTrack{
				Title:       "track 19",
				Artist:      "The Boss",
				AlbumArtist: "The Boss and the Band",
				Album:       "That Sad One",
				TrackNumber: 19,
			},
			trackInfo{
				Title:       "track 19",
				Artist:      "The Boss",
				TrackNumber: "19",
				TotalTracks: "",
				Album:       "That Sad One",
			},
		},
	}

	for _, test := range tests {
		c.Assert(exifTrackToTrackInfo(test.input), qt.Equals, test.expected)
	}
}
