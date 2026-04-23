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

func TestCmpTracks(t *testing.T) {
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
	}

	for _, test := range tests {
		c.Assert(CmpTracks(test.t1, test.t2), qt.CmpEquals(cmpopts.EquateApprox(0, 0.01)), test.expected)
	}
}
