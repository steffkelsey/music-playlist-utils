package common

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestIsExactMatch(t *testing.T) {
	c := qt.New(t)
	tests := []struct {
		s1       string
		s2       string
		expected bool
	}{
		{"Gold", "Gold", true},
		{"Gold: Bob Marley & The Wailers", "gold: bob marley & the wailers", true},
		{"Gold: Bob Marley & The Wailers", "Gold", false},
	}

	for _, test := range tests {
		c.Assert(IsExactMatch(test.s1, test.s2), qt.Equals, test.expected)
	}
}

func TestIsFuzzyMatch(t *testing.T) {
	c := qt.New(t)
	tests := []struct {
		s1        string
		s2        string
		expected1 float64
		expected2 float64
	}{
		{"Gold", "Gold", 1.0, 1.0},
		{"Gold: Bob Marley & The Wailers", "gold: bob marley & the wailers", 1.0, 1.0},
		{"Gold: Bob Marley & The Wailers", "Gold", 0.2, 0.45},
		{"Gold: Bob Marley & The Wailers", "Blue", 0.0, 0.0},
		{"The Stranger", "The Stranger (Remastered)", 0.66, 0.91},
	}

	for _, test := range tests {
		a1, a2 := IsFuzzyMatch(test.s1, test.s2)
		c.Assert(a1, qt.CmpEquals(cmpopts.EquateApprox(0, 0.01)), test.expected1)
		c.Assert(a2, qt.CmpEquals(cmpopts.EquateApprox(0, 0.01)), test.expected2)
	}
}

func TestCmpAlbumTracks(t *testing.T) {
	c := qt.New(t)
	tests := []struct {
		t1       TrackInfo
		t2       TrackInfo
		expected float64
	}{
		{
			TrackInfo{
				Title:           "track 1",
				Artist:          "artist 1",
				TrackNumber:     1,
				TotalTracks:     18,
				Album:           "artist 1 live",
				DurationSeconds: 293,
			},
			TrackInfo{
				Title:           "track 1",
				Artist:          "artist 1",
				TrackNumber:     1,
				TotalTracks:     18,
				Album:           "artist 1 live",
				DurationSeconds: 293,
			},
			1.0,
		},
		{
			TrackInfo{
				Title:           "track 1",
				Artist:          "artist 1",
				TrackNumber:     1,
				TotalTracks:     18,
				Album:           "artist 1 live",
				AlbumArtist:     "artist 1 live",
				DurationSeconds: 293,
			},
			TrackInfo{
				Title:           "other song",
				Artist:          "other person",
				TrackNumber:     2,
				TotalTracks:     10,
				Album:           "totally different",
				AlbumArtist:     "totally different",
				DurationSeconds: 120,
			},
			0.0,
		},
	}

	for _, test := range tests {
		c.Assert(CmpAlbumTracks(test.t1, test.t2), qt.CmpEquals(cmpopts.EquateApprox(0, 0.01)), test.expected)
	}
}

//func TestGetDuration(t *testing.T) {
//	c := qt.New(t)
//	tests := []struct {
//		input    string
//		expected float64
//	}{
//		{"/samples/The New Danger/01 - The Boogie Man Song.mp3", 143.06},
//		{"/samples/Togetherness/01 - L.T.D. - Holding On (When Love Is Gone).m4a", 238.75},
//	}
//
//	for _, test := range tests {
//		actual, err := GetDuration(test.input)
//		c.Assert(actual, qt.CmpEquals(cmpopts.EquateApprox(0, 0.01)), test.expected)
//	}
//}
