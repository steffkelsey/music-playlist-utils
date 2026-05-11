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
		{"Gold: Bob Marley & The Wailers", "Gold", 0.2, 0.42},
		{"Gold: Bob Marley & The Wailers", "Blue", 0.0, 0.0},
		{"The Stranger", "The Stranger (Remastered)", 0.66, 1.0},
		{"Rufus & Chaka Khan", "Rufus feat. Chaka Khan", 0.75, 1.0},
		{"Yazoo", "Yaz", 0.0, 0.85},
		{"Rufus featuring Chaka Khan", "Rufus feat. Chaka Khan", 0.75, 0.68},
		{"Little Children (Original Motion Picture Score)", "Little Children (Orginal Motion Picture Score)", 0.83, 0.81}, //nolint:misspell
	}

	for _, test := range tests {
		a1, a2 := IsFuzzyMatch(test.s1, test.s2)
		c.Assert(a1, qt.CmpEquals(cmpopts.EquateApprox(0, 0.01)), test.expected1)
		c.Assert(a2, qt.CmpEquals(cmpopts.EquateApprox(0, 0.01)), test.expected2)
	}
}

func TestCmpAlbums(t *testing.T) {
	c := qt.New(t)
	tests := []struct {
		a1       AlbumInfo
		a2       AlbumInfo
		expected float64
	}{
		{
			AlbumInfo{
				Album:       "Album 1",
				Artist:      "artist 1",
				TotalTracks: 1,
				TotalDiscs:  1,
			},
			AlbumInfo{
				Album:       "Album 1",
				Artist:      "artist 1",
				TotalTracks: 1,
				TotalDiscs:  1,
			},
			1.0,
		},
		{
			AlbumInfo{
				Album:       "Totally Different",
				Artist:      "some guy",
				TotalTracks: 6,
				TotalDiscs:  6,
			},
			AlbumInfo{
				Album:       "Album 1",
				Artist:      "artist 1",
				TotalTracks: 1,
				TotalDiscs:  1,
			},
			0.0,
		},
		{
			AlbumInfo{
				Album:       "Little Children (Original Motion Picture Score)",
				Artist:      "Thomas Newman",
				TotalTracks: 19,
				TotalDiscs:  1,
			},
			AlbumInfo{
				Album:       "Little Children (Original Motion Picture Score)",
				Artist:      "Thomas Newman",
				TotalTracks: 0,
				TotalDiscs:  0,
			},
			0.67,
		},
		{
			AlbumInfo{
				Album:       "Little Children (Original Motion Picture Score)",
				Artist:      "Thomas Newman",
				TotalTracks: 19,
				TotalDiscs:  1,
			},
			AlbumInfo{
				Album:       "Little Children (Orginal Motion Picture Score)", //nolint:misspell
				Artist:      "Thomas Newman",
				TotalTracks: 0,
				TotalDiscs:  0,
			},
			0.61,
		},
		{
			AlbumInfo{
				Album:       "Alright, Still",
				Artist:      "Lily Allen",
				TotalTracks: 14,
				TotalDiscs:  1,
			},
			AlbumInfo{
				Album:  "Alright, Still",
				Artist: "Lily Allen",
			},
			0.67,
		},
		{
			AlbumInfo{
				Album:       "Alright, Still",
				Artist:      "Lily Allen",
				TotalTracks: 14,
				TotalDiscs:  1,
			},
			AlbumInfo{
				Album:       "Alright, Still",
				Artist:      "Lily Allen",
				TotalTracks: 11,
			},
			0.67,
		},
		{
			AlbumInfo{
				Album:       "Alright, Still",
				Artist:      "Lily Allen",
				TotalTracks: 14,
				TotalDiscs:  1,
			},
			AlbumInfo{
				Album:       "Alright, Still (Bonus Track Version)",
				Artist:      "Lily Allen",
				TotalTracks: 13,
			},
			0.75,
		},
	}

	for _, test := range tests {
		c.Assert(CmpAlbums(test.a1, test.a2), qt.CmpEquals(cmpopts.EquateApprox(0, 0.01)), test.expected)
	}
}

func TestCmpAlbumsWithTracks(t *testing.T) {
	c := qt.New(t)
	tests := []struct {
		a1            AlbumInfo
		a2            AlbumInfo
		expectedScore float64
		expectedMap   map[int]int
	}{
		{
			AlbumInfo{
				Album:       "Album 1",
				Artist:      "artist 1",
				TotalTracks: 1,
				TotalDiscs:  1,
				Tracks: []TrackInfo{
					{
						DiscNumber:      1,
						TotalDiscs:      1,
						TrackNumber:     1,
						TotalTracks:     1,
						Title:           "track 1",
						Artist:          "artist 1 feat two",
						Album:           "Album 1",
						AlbumArtist:     "artist 1",
						DurationSeconds: 10,
					},
				},
			},
			AlbumInfo{
				Album:       "Album 1",
				Artist:      "artist 1",
				TotalTracks: 1,
				TotalDiscs:  1,
				Tracks: []TrackInfo{
					{
						DiscNumber:      1,
						TotalDiscs:      1,
						TrackNumber:     1,
						TotalTracks:     1,
						Title:           "track 1",
						Artist:          "artist 1 feat two",
						Album:           "Album 1",
						AlbumArtist:     "artist 1",
						DurationSeconds: 10,
					},
				},
			},
			1.0,
			map[int]int{0: 0}, // match is for slice index
		},
		{
			AlbumInfo{
				Album:       "Totally Different",
				Artist:      "some guy",
				TotalTracks: 6,
				TotalDiscs:  6,
				Tracks: []TrackInfo{
					{
						DiscNumber:      1,
						TotalDiscs:      6,
						TrackNumber:     1,
						TotalTracks:     2,
						Title:           "track 1",
						Artist:          "some guy",
						Album:           "Totally Different",
						AlbumArtist:     "some Guy",
						DurationSeconds: 60,
					},
				},
			},
			AlbumInfo{
				Album:       "Album 1",
				Artist:      "artist 1",
				TotalTracks: 1,
				TotalDiscs:  1,
				Tracks: []TrackInfo{
					{
						DiscNumber:      1,
						TotalDiscs:      1,
						TrackNumber:     1,
						TotalTracks:     1,
						Title:           "track 1",
						Artist:          "artist 1",
						Album:           "Album 1",
						AlbumArtist:     "artist 1",
						DurationSeconds: 120,
					},
				},
			},
			0.0,
			map[int]int{},
		},
		{
			AlbumInfo{
				Album:       "Alright, Still",
				Artist:      "Lily Allen",
				TotalTracks: 14,
				TotalDiscs:  1,
				Tracks: []TrackInfo{
					{
						DiscNumber:      1,
						TotalDiscs:      1,
						TrackNumber:     1,
						TotalTracks:     14,
						Title:           "Smile",
						Artist:          "Lily Allen",
						Album:           "Alright, Still",
						AlbumArtist:     "Lily Allen",
						DurationSeconds: 197,
					},
					{
						Title:           "Knock 'Em Out",
						Artist:          "Lily Allen",
						DiscNumber:      1,
						TotalDiscs:      1,
						TrackNumber:     2,
						TotalTracks:     14,
						Album:           "Alright, Still",
						AlbumArtist:     "Lily Allen",
						DurationSeconds: 174,
					},
					{
						Title:           "LDN",
						Artist:          "Lily Allen",
						DiscNumber:      1,
						TotalDiscs:      1,
						TrackNumber:     3,
						TotalTracks:     14,
						Album:           "Alright, Still",
						AlbumArtist:     "Lily Allen",
						DurationSeconds: 191,
					},
					{
						Title:           "Everythings Just Wonderful",
						Artist:          "Lily Allen",
						DiscNumber:      1,
						TotalDiscs:      1,
						TrackNumber:     4,
						TotalTracks:     14,
						Album:           "Alright, Still",
						AlbumArtist:     "Lily Allen",
						DurationSeconds: 209,
					},
					{
						Title:           "Not Big",
						Artist:          "Lily Allen",
						DiscNumber:      1,
						TotalDiscs:      1,
						TrackNumber:     5,
						TotalTracks:     14,
						Album:           "Alright, Still",
						AlbumArtist:     "Lily Allen",
						DurationSeconds: 197,
					},
					{
						Title:           "Friday Night",
						Artist:          "Lily Allen",
						DiscNumber:      1,
						TotalDiscs:      1,
						TrackNumber:     6,
						TotalTracks:     14,
						Album:           "Alright, Still",
						AlbumArtist:     "Lily Allen",
						DurationSeconds: 187,
					},
					{
						Title:           "Shame for You",
						Artist:          "Lily Allen",
						DiscNumber:      1,
						TotalDiscs:      1,
						TrackNumber:     7,
						TotalTracks:     14,
						Album:           "Alright, Still",
						AlbumArtist:     "Lily Allen",
						DurationSeconds: 246,
					},
					{
						Title:           "Littlest Things",
						Artist:          "Lily Allen",
						DiscNumber:      1,
						TotalDiscs:      1,
						TrackNumber:     8,
						TotalTracks:     14,
						Album:           "Alright, Still",
						AlbumArtist:     "Lily Allen",
						DurationSeconds: 182,
					},
					{
						Title:           "Take What You Take",
						Artist:          "Lily Allen",
						DiscNumber:      1,
						TotalDiscs:      1,
						TrackNumber:     9,
						TotalTracks:     14,
						Album:           "Alright, Still",
						AlbumArtist:     "Lily Allen",
						DurationSeconds: 246,
					},
					{
						Title:           "Friend of Mine",
						Artist:          "Lily Allen",
						DiscNumber:      1,
						TotalDiscs:      1,
						TrackNumber:     10,
						TotalTracks:     14,
						Album:           "Alright, Still",
						AlbumArtist:     "Lily Allen",
						DurationSeconds: 238,
					},
					{
						Title:           "Alfie",
						Artist:          "Lily Allen",
						DiscNumber:      1,
						TotalDiscs:      1,
						TrackNumber:     11,
						TotalTracks:     14,
						Album:           "Alright, Still",
						AlbumArtist:     "Lily Allen",
						DurationSeconds: 165,
					},
					{
						Title:           "Nan You're a Window Shopper",
						Artist:          "Lily Allen",
						DiscNumber:      1,
						TotalDiscs:      1,
						TrackNumber:     12,
						TotalTracks:     14,
						Album:           "Alright, Still",
						AlbumArtist:     "Lily Allen",
						DurationSeconds: 178,
					},
					{
						Title:           "Smile (Version Revisited) [Mark Ronson Remix]",
						Artist:          "Lily Allen",
						DiscNumber:      1,
						TotalDiscs:      1,
						TrackNumber:     13,
						TotalTracks:     14,
						Album:           "Alright, Still",
						AlbumArtist:     "Lily Allen",
						DurationSeconds: 193,
					},
					{
						Title:           "Blank Expression",
						Artist:          "Lily Allen",
						DiscNumber:      1,
						TotalDiscs:      1,
						TrackNumber:     14,
						TotalTracks:     14,
						Album:           "Alright, Still",
						AlbumArtist:     "Lily Allen",
						DurationSeconds: 154,
					},
				},
			},
			AlbumInfo{
				Album:       "Alright, Still (Bonus Track Version)",
				Artist:      "Lily Allen",
				TotalTracks: 13,
				Tracks: []TrackInfo{
					{
						TrackNumber:     1,
						TotalTracks:     13,
						Title:           "Smile",
						Artist:          "Lily Allen",
						Album:           "Alright, Still (Bonus Track Version)",
						AlbumArtist:     "Lily Allen",
						DurationSeconds: 197,
					},
					{
						Title:           "Knock 'Em Out",
						Artist:          "Lily Allen",
						TrackNumber:     2,
						TotalTracks:     13,
						Album:           "Alright, Still (Bonus Track Version)",
						AlbumArtist:     "Lily Allen",
						DurationSeconds: 174,
					},
					{
						Title:           "LDN",
						Artist:          "Lily Allen",
						TrackNumber:     3,
						TotalTracks:     13,
						Album:           "Alright, Still (Bonus Track Version)",
						AlbumArtist:     "Lily Allen",
						DurationSeconds: 193,
					},
					{
						Title:           "Everything's Just Wonderful",
						Artist:          "Lily Allen",
						TrackNumber:     4,
						TotalTracks:     13,
						Album:           "Alright, Still (Bonus Track Version)",
						AlbumArtist:     "Lily Allen",
						DurationSeconds: 209,
					},
					{
						Title:           "Not Big",
						Artist:          "Lily Allen",
						TrackNumber:     5,
						TotalTracks:     13,
						Album:           "Alright, Still (Bonus Track Version)",
						AlbumArtist:     "Lily Allen",
						DurationSeconds: 197,
					},
					{
						Title:           "Friday Night",
						Artist:          "Lily Allen",
						TrackNumber:     6,
						TotalTracks:     13,
						Album:           "Alright, Still (Bonus Track Version)",
						AlbumArtist:     "Lily Allen",
						DurationSeconds: 187,
					},
					{
						Title:           "Shame for You",
						Artist:          "Lily Allen",
						TrackNumber:     7,
						TotalTracks:     13,
						Album:           "Alright, Still (Bonus Track Version)",
						AlbumArtist:     "Lily Allen",
						DurationSeconds: 247,
					},
					{
						Title:           "Littlest Things",
						Artist:          "Lily Allen",
						TrackNumber:     8,
						TotalTracks:     13,
						Album:           "Alright, Still (Bonus Track Version)",
						AlbumArtist:     "Lily Allen",
						DurationSeconds: 188,
					},
					{
						Title:           "Take What You Take",
						Artist:          "Lily Allen",
						TrackNumber:     9,
						TotalTracks:     13,
						Album:           "Alright, Still (Bonus Track Version)",
						AlbumArtist:     "Lily Allen",
						DurationSeconds: 247,
					},
					{
						Title:           "Friend of Mine",
						Artist:          "Lily Allen",
						TrackNumber:     10,
						TotalTracks:     13,
						Album:           "Alright, Still (Bonus Track Version)",
						AlbumArtist:     "Lily Allen",
						DurationSeconds: 238,
					},
					{
						Title:           "Alfie",
						Artist:          "Lily Allen",
						TrackNumber:     11,
						TotalTracks:     13,
						Album:           "Alright, Still (Bonus Track Version)",
						AlbumArtist:     "Lily Allen",
						DurationSeconds: 167,
					},
					{
						Title:           "Nan You're a Window Shopper",
						Artist:          "Lily Allen",
						TrackNumber:     12,
						TotalTracks:     13,
						Album:           "Alright, Still (Bonus Track Version)",
						AlbumArtist:     "Lily Allen",
						DurationSeconds: 172,
					},
					{
						Title:           "Smile Version Revisited (Mark Ronson Remix)",
						Artist:          "Lily Allen",
						TrackNumber:     13,
						TotalTracks:     13,
						Album:           "Alright, Still (Bonus Track Version)",
						AlbumArtist:     "Lily Allen",
						DurationSeconds: 196,
					},
				},
			},
			0.80,
			map[int]int{0: 0, 1: 1, 2: 2, 3: 3, 4: 4, 5: 5, 6: 6, 7: 7, 8: 8, 9: 9, 10: 10, 11: 11, 12: 12},
		},
	}

	for _, test := range tests {
		actualScore, actualMap := CmpAlbumsWithTracks(test.a1, test.a2, 0.70)
		c.Assert(actualScore, qt.CmpEquals(cmpopts.EquateApprox(0, 0.01)), test.expectedScore)
		c.Assert(actualMap, qt.DeepEquals, test.expectedMap)
	}
}

func TestCmpTracksWithAlbumInfo(t *testing.T) {
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
				DiscNumber:      1,
				TotalDiscs:      1,
				TrackNumber:     1,
				TotalTracks:     18,
				Album:           "artist 1 live",
				AlbumArtist:     "artist 1 live",
				DurationSeconds: 293,
			},
			TrackInfo{
				Title:           "other song",
				Artist:          "other person",
				DiscNumber:      2,
				TotalDiscs:      2,
				TrackNumber:     2,
				TotalTracks:     10,
				Album:           "totally different",
				AlbumArtist:     "totally different",
				DurationSeconds: 120,
			},
			0.0,
		},
		{
			TrackInfo{
				Title:           "Master Blaster (Jammin')",
				Artist:          "Stevie Wonder",
				TrackNumber:     3,
				TotalTracks:     8,
				Album:           "Stevie Wonder's Original Musiquarium I (Reissue)",
				AlbumArtist:     "Stevie Wonder",
				DurationSeconds: 308,
			},
			TrackInfo{
				Title:           "Master Blaster (Jammin')",
				Artist:          "Stevie Wonder",
				TrackNumber:     3,
				TotalTracks:     8,
				Album:           "Original Musiquarium I",
				AlbumArtist:     "Stevie Wonder",
				DurationSeconds: 308,
			},
			0.91,
		},
		{
			TrackInfo{
				Title:           "Moonlight Feels Right",
				Artist:          "Starbuck",
				TrackNumber:     24,
				TotalTracks:     24,
				Album:           "The Very Best",
				AlbumArtist:     "Starbuck",
				DurationSeconds: 219,
			},
			TrackInfo{
				Title:           "Moonlight Feels Right",
				Artist:          "Starbuck",
				TrackNumber:     5,
				TotalTracks:     10,
				Album:           "Moonlight Feels Right",
				AlbumArtist:     "Starbuck",
				DurationSeconds: 218,
			},
			0.5,
		},
		{
			TrackInfo{
				Title:           "Hollywood Swinging",
				Artist:          "Kool & The Gang",
				DiscNumber:      1,
				TotalDiscs:      2,
				TrackNumber:     9,
				TotalTracks:     16,
				Album:           "Gold",
				AlbumArtist:     "Kool & The Gang",
				DurationSeconds: 278,
			},
			TrackInfo{
				Title:           "Hollywood Swinging",
				Artist:          "Kool & The Gang",
				DiscNumber:      5,
				TotalDiscs:      6,
				TrackNumber:     2,
				TotalTracks:     23,
				Album:           "Can You Dig It? The '70s Soul Experience (Disc 5)",
				AlbumArtist:     "Kool & The Gang",
				DurationSeconds: 270,
			},
			0.375,
		},
	}

	for _, test := range tests {
		c.Assert(CmpTracksWithAlbumInfo(test.t1, test.t2), qt.CmpEquals(cmpopts.EquateApprox(0, 0.01)), test.expected)
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
		{
			TrackInfo{
				Title:           "Master Blaster (Jammin')",
				Artist:          "Stevie Wonder",
				TrackNumber:     3,
				TotalTracks:     8,
				Album:           "Stevie Wonder's Original Musiquarium I (Reissue)",
				AlbumArtist:     "Stevie Wonder",
				DurationSeconds: 308,
			},
			TrackInfo{
				Title:           "Master Blaster (Jammin')",
				Artist:          "Stevie Wonder",
				TrackNumber:     3,
				TotalTracks:     8,
				Album:           "Original Musiquarium I",
				AlbumArtist:     "Stevie Wonder",
				DurationSeconds: 308,
			},
			1.0,
		},
		{
			TrackInfo{
				Title:           "Moonlight Feels Right",
				Artist:          "Starbuck",
				TrackNumber:     24,
				TotalTracks:     24,
				Album:           "The Very Best",
				AlbumArtist:     "Starbuck",
				DurationSeconds: 219,
			},
			TrackInfo{
				Title:           "Moonlight Feels Right",
				Artist:          "Starbuck",
				TrackNumber:     5,
				TotalTracks:     10,
				Album:           "Moonlight Feels Right",
				AlbumArtist:     "Starbuck",
				DurationSeconds: 218,
			},
			0.97,
		},
		{
			TrackInfo{
				Title:           "Hollywood Swinging",
				Artist:          "Kool and The Gang",
				TrackNumber:     9,
				TotalTracks:     16,
				Album:           "Gold",
				AlbumArtist:     "Kool & The Gang",
				DurationSeconds: 278,
			},
			TrackInfo{
				Title:           "Hollywood Swinging",
				Artist:          "Kool & The Gang",
				TrackNumber:     2,
				TotalTracks:     23,
				Album:           "Can You Dig It? The '70s Soul Experience (Disc 5)",
				AlbumArtist:     "",
				DurationSeconds: 270,
			},
			0.72,
		},
	}

	for _, test := range tests {
		c.Assert(CmpTracks(test.t1, test.t2), qt.CmpEquals(cmpopts.EquateApprox(0, 0.01)), test.expected)
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

func TestSubstrMagic(t *testing.T) {
	c := qt.New(t)
	tests := []struct {
		a1       []string
		a2       []string
		expected float64
	}{
		{[]string{"rufus", "chaka", "khan"}, []string{"rufus", "feat", "chaka", "khan"}, 1.0},
		{[]string{"gold", "bob", "marley", "the", "wailers"}, []string{"blue"}, 0.0},
	}

	for _, test := range tests {
		c.Assert(SubstrMagic(test.a1, test.a2), qt.CmpEquals(cmpopts.EquateApprox(0, 0.01)), test.expected)
	}
}
