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
				Album:      "Single Disc, Single Track",
				Artist:     "Mister Test",
				TotalDiscs: 1,
				Tracks: []TrackInfo{
					{
						DiscNumber:  1,
						TotalDiscs:  1,
						TrackNumber: 1,
						TotalTracks: 1,
						Album:       "Single Disc, Single Track",
						AlbumArtist: "Mister Test",
					},
				},
			},
			true,
		},
		{
			AlbumInfo{
				Album:      "Two Discs, Three Tracks Total",
				Artist:     "Test",
				TotalDiscs: 2,
				Tracks: []TrackInfo{
					{
						DiscNumber:  1,
						TotalDiscs:  2,
						TrackNumber: 1,
						TotalTracks: 1,
						Album:       "Two Discs, Three Tracks Total",
						AlbumArtist: "Test",
					},
					{
						DiscNumber:  2,
						TotalDiscs:  2,
						TrackNumber: 1,
						TotalTracks: 2,
						Album:       "Two Discs, Three Tracks Total",
						AlbumArtist: "Test",
					},
					{
						DiscNumber:  2,
						TotalDiscs:  2,
						TrackNumber: 2,
						TotalTracks: 2,
						Album:       "Two Discs, Three Tracks Total",
						AlbumArtist: "Test",
					},
				},
			},
			true,
		},
		{
			AlbumInfo{
				Album:      "Two Discs, Three Tracks Total, Missing Track 1",
				Artist:     "Test",
				TotalDiscs: 2,
				Tracks: []TrackInfo{
					{
						DiscNumber:  1,
						TotalDiscs:  2,
						TrackNumber: 1,
						TotalTracks: 1,
						Album:       "Two Discs, Three Tracks Total, Missing Track 1",
						AlbumArtist: "Test",
					},
					{
						DiscNumber:  2,
						TotalDiscs:  2,
						TrackNumber: 2,
						TotalTracks: 2,
						Album:       "Two Discs, Three Tracks Total, Missing Track 1",
						AlbumArtist: "Test",
					},
				},
			},
			false,
		},
		{
			AlbumInfo{
				Album:      "Single Disc, No Tracks",
				Artist:     "Mister Test",
				Tracks:     make([]TrackInfo, 0),
				TotalDiscs: 1,
			},
			false,
		},
	}

	for _, test := range tests {
		c.Assert(test.input.IsComplete(), qt.Equals, test.expected)
	}
}

func TestAlbumIsDiscComplete(t *testing.T) {
	c := qt.New(t)
	tests := []struct {
		album    AlbumInfo
		disc     int
		expected bool
	}{
		{
			AlbumInfo{
				Album:      "Single Disc, No Tracks",
				Artist:     "Mister Test",
				Tracks:     make([]TrackInfo, 0),
				TotalDiscs: 1,
			},
			1,
			false,
		},
		{
			AlbumInfo{
				Album:  "Single Disc, Single Track",
				Artist: "Mister Test",
				Tracks: []TrackInfo{
					{
						DiscNumber:  1,
						TotalDiscs:  1,
						TrackNumber: 1,
						TotalTracks: 1,
						Album:       "Single Disc, Single Track",
						AlbumArtist: "Mister Test",
					},
				},
				TotalDiscs: 1,
			},
			1,
			true,
		},
		{
			AlbumInfo{
				Album:      "Two Discs, Three Tracks Total",
				Artist:     "Test",
				TotalDiscs: 2,
				Tracks: []TrackInfo{
					{
						DiscNumber:  1,
						TotalDiscs:  2,
						TrackNumber: 1,
						TotalTracks: 1,
						Album:       "Single Disc, Three Tracks Total",
						AlbumArtist: "Test",
					},
					{
						DiscNumber:  2,
						TotalDiscs:  2,
						TrackNumber: 1,
						TotalTracks: 2,
						Album:       "Single Disc, Three Tracks Total",
						AlbumArtist: "Test",
					},
				},
			},
			2,
			false,
		},
		{
			AlbumInfo{
				Album:      "Two Discs, Three Tracks Total",
				Artist:     "Test",
				TotalDiscs: 2,
				Tracks: []TrackInfo{
					{
						DiscNumber:  1,
						TotalDiscs:  2,
						TrackNumber: 1,
						TotalTracks: 1,
						Album:       "Single Disc, Three Tracks Total",
						AlbumArtist: "Test",
					},
					{
						DiscNumber:  2,
						TotalDiscs:  2,
						TrackNumber: 1,
						TotalTracks: 2,
						Album:       "Single Disc, Three Tracks Total",
						AlbumArtist: "Test",
					},
					{
						DiscNumber:  2,
						TotalDiscs:  2,
						TrackNumber: 2,
						TotalTracks: 2,
						Album:       "Single Disc, Three Tracks Total",
						AlbumArtist: "Test",
					},
				},
			},
			2,
			true,
		},
		{
			AlbumInfo{
				Album:      "Two Discs, Three Tracks Total, Missing Track 1",
				Artist:     "Test",
				TotalDiscs: 2,
				Tracks: []TrackInfo{
					{
						DiscNumber:  1,
						TotalDiscs:  2,
						TrackNumber: 1,
						TotalTracks: 1,
						Album:       "Single Disc, Three Tracks Total, Missing Track 1",
						AlbumArtist: "Test",
					},
					{
						DiscNumber:  2,
						TotalDiscs:  2,
						TrackNumber: 2,
						TotalTracks: 3,
						Album:       "Single Disc, Three Tracks Total, Missing Track 1",
						AlbumArtist: "Test",
					},
					{
						DiscNumber:  2,
						TotalDiscs:  2,
						TrackNumber: 3,
						TotalTracks: 3,
						Album:       "Single Disc, Three Tracks Total, Missing Track 1",
						AlbumArtist: "Test",
					},
				},
			},
			2,
			false,
		},
	}

	for _, test := range tests {
		c.Assert(test.album.IsDiscComplete(test.disc), qt.Equals, test.expected)
	}
}
