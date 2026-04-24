package common

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestAlbumIsComplete(t *testing.T) {
	c := qt.New(t)
	tests := []struct {
		input    AlbumInfo
		expected bool
	}{
		{
			AlbumInfo{
				Album:  "The Best of",
				Artist: "Spongebob",
				Tracks: []TrackInfo{
					{},
				},
				TotalTracks: 1,
			},
			true,
		},
		{
			AlbumInfo{
				Album:       "The Best of",
				Artist:      "Spongebob",
				TotalTracks: 1,
			},
			false,
		},
	}

	for _, test := range tests {
		c.Assert(test.input.IsComplete(), qt.Equals, test.expected)
	}
}
