package cmd

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/google/go-cmp/cmp/cmpopts"

	"music-utils/common"
)

func TestRankDuplicates(t *testing.T) {
	c := qt.New(t)
	tests := []struct {
		input    []trackInfoExtended
		expected duplicateResult
	}{
		// Same tracks, same album, one has higher bitrate
		{
			[]trackInfoExtended{
				{
					AlbumTrackPercentage: 2.0,
					TrackInfo: common.TrackInfo{
						Title:           "Zimzallabim",
						Artist:          "Mos Def",
						DiscNumber:      1,
						TotalDiscs:      1,
						TrackNumber:     4,
						TotalTracks:     18,
						Album:           "The New Danger",
						AlbumArtist:     "Mos Def",
						DurationSeconds: 221,
						BitRate:         250,
						Path:            "/home/Music/Mos Def/The New Danger/04 - Mos Def - Zimzallabim.opus",
					},
				},
				{
					AlbumTrackPercentage: 2.0,
					TrackInfo: common.TrackInfo{
						Title:           "Zimzallabim",
						Artist:          "Mos Def",
						DiscNumber:      1,
						TotalDiscs:      1,
						TrackNumber:     4,
						TotalTracks:     18,
						Album:           "The New Danger",
						AlbumArtist:     "Mos Def",
						DurationSeconds: 221,
						BitRate:         130,
						Path:            "/home/Music/Mos Def/The New Danger/04 - Mos Def - Zimzallabim.mp3",
					},
				},
			},
			duplicateResult{ // Delete the one with the lower bitrate
				Keep: []trackInfoExtended{
					{
						TrackInfo: common.TrackInfo{
							Path: "/home/Music/Mos Def/The New Danger/04 - Mos Def - Zimzallabim.opus",
						},
					},
				},
				Delete: []trackInfoExtended{
					{
						TrackInfo: common.TrackInfo{
							Path: "/home/Music/Mos Def/The New Danger/04 - Mos Def - Zimzallabim.mp3",
						},
					},
				},
			},
		},
		// Same tracks, same album, one has the filename of a dupelicate
		{
			[]trackInfoExtended{
				{
					AlbumTrackPercentage: 0.9,
					TrackInfo: common.TrackInfo{
						Title:           "Xanadu",
						Artist:          "Olivia Newton‐John",
						DiscNumber:      1,
						TotalDiscs:      1,
						TrackNumber:     17,
						TotalTracks:     21,
						Album:           "The Singles Collection 1971-1992",
						AlbumArtist:     "Olivia Newton‐John",
						DurationSeconds: 206,
						BitRate:         130,
						Path:            "/home/Music/Olivia Newton‐John/The Singles Collection 1971-1992/17 - Olivia Newton‐John - Xanadu(2).mp3",
					},
				},
				{
					AlbumTrackPercentage: 0.9,
					TrackInfo: common.TrackInfo{
						Title:           "Xanadu",
						Artist:          "Olivia Newton‐John",
						DiscNumber:      1,
						TotalDiscs:      1,
						TrackNumber:     17,
						TotalTracks:     21,
						Album:           "The Singles Collection 1971-1992",
						AlbumArtist:     "Olivia Newton‐John",
						DurationSeconds: 208,
						BitRate:         130,
						Path:            "/home/Music/Olivia Newton‐John/The Singles Collection 1971-1992/17 - Olivia Newton‐John - Xanadu.mp3",
					},
				},
			},
			duplicateResult{ // Delete the one with the worst filename
				Keep: []trackInfoExtended{
					{
						TrackInfo: common.TrackInfo{
							Path: "/home/Music/Olivia Newton‐John/The Singles Collection 1971-1992/17 - Olivia Newton‐John - Xanadu.mp3",
						},
					},
				},
				Delete: []trackInfoExtended{
					{
						TrackInfo: common.TrackInfo{
							Path: "/home/Music/Olivia Newton‐John/The Singles Collection 1971-1992/17 - Olivia Newton‐John - Xanadu(2).mp3",
						},
					},
				},
			},
		},
		{
			[]trackInfoExtended{
				{
					AlbumTrackPercentage: 1.1538461538461537,
					TrackInfo: common.TrackInfo{
						Path:            "/home/Music/Keane/Hopes and Fears (deluxe edition)/09 - Keane - This Is the Last Time.mp3",
						Title:           "This Is the Last Time",
						Artist:          "Keane",
						DiscNumber:      1,
						TotalDiscs:      1,
						TrackNumber:     9,
						TotalTracks:     26,
						Album:           "Hopes and Fears (deluxe edition)",
						AlbumArtist:     "Keane",
						DurationSeconds: 209,
						BitRate:         130,
					},
				},
				{
					AlbumTrackPercentage: 0.09090909090909091,
					TrackInfo: common.TrackInfo{
						Path:            "/home/Music/Keane/Hopes and Fears/02 - Keane - This Is the Last Time.mp3",
						Title:           "This Is the Last Time",
						Artist:          "Keane",
						DiscNumber:      1,
						TotalDiscs:      1,
						TrackNumber:     2,
						TotalTracks:     11,
						Album:           "Hopes and Fears",
						AlbumArtist:     "Keane",
						DurationSeconds: 211,
						BitRate:         190,
					},
				},
				{
					AlbumTrackPercentage: 1.1538461538461537,
					TrackInfo: common.TrackInfo{
						Path:            "/home/Music/Keane/Hopes and Fears (deluxe edition)/25 - Keane - This Is the Last Time (live at Mill St Brewery, Toronto _ 2004 _ acoustic).mp3",
						Title:           "This Is the Last Time (live at Mill St Brewery, Toronto / 2004 / acoustic)",
						Artist:          "Keane",
						DiscNumber:      1,
						TotalDiscs:      1,
						TrackNumber:     25,
						TotalTracks:     26,
						Album:           "Hopes and Fears (deluxe edition)",
						AlbumArtist:     "Keane",
						DurationSeconds: 207,
						BitRate:         110,
					},
				},
			},
			duplicateResult{ // Delete the one from the sparse album even though the bitrate is highest
				Keep: []trackInfoExtended{
					{
						TrackInfo: common.TrackInfo{
							Path: "/home/Music/Keane/Hopes and Fears (deluxe edition)/09 - Keane - This Is the Last Time.mp3",
						},
					},
					{
						TrackInfo: common.TrackInfo{
							Path: "/home/Music/Keane/Hopes and Fears (deluxe edition)/25 - Keane - This Is the Last Time (live at Mill St Brewery, Toronto _ 2004 _ acoustic).mp3",
						},
					},
				},
				Delete: []trackInfoExtended{
					{
						TrackInfo: common.TrackInfo{
							Path: "/home/Music/Keane/Hopes and Fears/02 - Keane - This Is the Last Time.mp3",
						},
					},
				},
			},
		},
		{
			[]trackInfoExtended{
				{
					AlbumTrackPercentage: 0.055555555555555,
					TrackInfo: common.TrackInfo{
						Path:            "/home/Music/Various Artists/One Shot ’80, Volume 12: Movies/18 - Joe Cocker \u0026 Jennifer Warnes - Up Where We Belong.m4a",
						Title:           "Up Where We Belong",
						Artist:          "Joe Cocker \u0026 Jennifer Warnes",
						DiscNumber:      1,
						TotalDiscs:      1,
						TrackNumber:     18,
						TotalTracks:     18,
						Album:           "One Shot ’80, Volume 12: Movies",
						AlbumArtist:     "Various Artists",
						DurationSeconds: 236,
						BitRate:         130,
					},
				},
				{
					AlbumTrackPercentage: 0.16666666666666,
					TrackInfo: common.TrackInfo{
						Path:            "/home/Music/Various Artists/The One and Only Love Album/Disc 1/11 - Joe Cocker \u0026 Jennifer Warnes - Up Where We Belong.mp3",
						Title:           "Up Where We Belong",
						Artist:          "Joe Cocker \u0026 Jennifer Warnes",
						DiscNumber:      1,
						TotalDiscs:      2,
						TrackNumber:     11,
						TotalTracks:     12,
						Album:           "The One and Only Love Album",
						AlbumArtist:     "Various Artists",
						DurationSeconds: 237,
						BitRate:         130,
					},
				},
			},
			duplicateResult{ // Delete the one from the sparse album even though the bitrate is highest
				Keep: []trackInfoExtended{
					{
						TrackInfo: common.TrackInfo{
							Path: "/home/Music/Various Artists/The One and Only Love Album/Disc 1/11 - Joe Cocker \u0026 Jennifer Warnes - Up Where We Belong.mp3",
						},
					},
				},
				Delete: []trackInfoExtended{
					{
						TrackInfo: common.TrackInfo{
							Path: "/home/Music/Various Artists/One Shot ’80, Volume 12: Movies/18 - Joe Cocker \u0026 Jennifer Warnes - Up Where We Belong.m4a",
						},
					},
				},
			},
		},
		{
			[]trackInfoExtended{
				{
					AlbumTrackPercentage: 1.0,
					TrackInfo: common.TrackInfo{
						Path:            "/home/Music/Steely Dan/Citizen Steely Dan: 1972–1980/Disc 2/16 - Steely Dan - Bad Sneakers.m4a",
						Title:           "Bad Sneakers",
						Artist:          "Steely Dan",
						DiscNumber:      2,
						TotalDiscs:      4,
						TrackNumber:     16,
						TotalTracks:     21,
						Album:           "Citizen Steely Dan: 1972–1980",
						AlbumArtist:     "Steely Dan",
						DurationSeconds: 199,
						BitRate:         130,
					},
				},
				{
					AlbumTrackPercentage: 1.0,
					TrackInfo: common.TrackInfo{
						Path:            "/home/Music/Steely Dan/A Decade of Steely Dan/14 - Steely Dan - Bad Sneakers.mp3",
						Title:           "Bad Sneakers",
						Artist:          "Steely Dan",
						DiscNumber:      1,
						TotalDiscs:      1,
						TrackNumber:     14,
						TotalTracks:     14,
						Album:           "A Decade of Steely Dan",
						AlbumArtist:     "Steely Dan",
						DurationSeconds: 199,
						BitRate:         130,
					},
				},
			},
			duplicateResult{ // 2 complete albums, keep them all
				Keep: []trackInfoExtended{
					{
						TrackInfo: common.TrackInfo{
							Path: "/home/Music/Steely Dan/Citizen Steely Dan: 1972–1980/Disc 2/16 - Steely Dan - Bad Sneakers.m4a",
						},
					},
					{
						TrackInfo: common.TrackInfo{
							Path: "/home/Music/Steely Dan/A Decade of Steely Dan/14 - Steely Dan - Bad Sneakers.mp3",
						},
					},
				},
				Delete: []trackInfoExtended{},
			},
		},
		{
			[]trackInfoExtended{
				{
					AlbumTrackPercentage: 1.0,
					TrackInfo: common.TrackInfo{
						Path:            "/home/Music/Yazoo/Upstairs at Eric’s/11 - Yazoo - Bring Your Love Down (Didn’t I).mp3",
						Title:           "Bring Your Love Down (Didn’t I)",
						Artist:          "Yazoo",
						DiscNumber:      1,
						TotalDiscs:      1,
						TrackNumber:     11,
						TotalTracks:     11,
						Album:           "Upstairs at Eric’s",
						AlbumArtist:     "Yazoo",
						DurationSeconds: 280,
						BitRate:         130,
					},
				},
				{
					AlbumTrackPercentage: 0.36363636,
					TrackInfo: common.TrackInfo{
						Path:            "/home/Music/Yazoo/Upstairs at Eric's/11 - Yazoo - Bring Your Love Down (Didn’t I).mp3",
						Title:           "Bring Your Love Down (Didn’t I)",
						Artist:          "Yazoo",
						DiscNumber:      1,
						TotalDiscs:      1,
						TrackNumber:     11,
						TotalTracks:     11,
						Album:           "Upstairs at Eric's",
						AlbumArtist:     "Yazoo",
						DurationSeconds: 280,
						BitRate:         130,
					},
				},
			},
			duplicateResult{ // 1 album is way MORE sparse than the other (almost same album name)
				Keep: []trackInfoExtended{
					{
						TrackInfo: common.TrackInfo{
							Path: "/home/Music/Yazoo/Upstairs at Eric’s/11 - Yazoo - Bring Your Love Down (Didn’t I).mp3",
						},
					},
				},
				Delete: []trackInfoExtended{
					{
						TrackInfo: common.TrackInfo{
							Path: "/home/Music/Yazoo/Upstairs at Eric's/11 - Yazoo - Bring Your Love Down (Didn’t I).mp3",
						},
					},
				},
			},
		},
		{
			[]trackInfoExtended{
				{
					AlbumTrackPercentage: 0.066666666667,
					TrackInfo: common.TrackInfo{
						Path:            "/home/Music/Various Artists/The Encyclopaedia of Music: Best of 50’s, Volume 2/12 - Peggy Lee - Baubles Bangles and Beads.MP3",
						Title:           "Baubles Bangles and Beads",
						Artist:          "Peggy Lee",
						DiscNumber:      1,
						TotalDiscs:      1,
						TrackNumber:     12,
						TotalTracks:     15,
						Album:           "The Encyclopaedia of Music: Best of 50’s, Volume 2",
						AlbumArtist:     "Various Artists",
						DurationSeconds: 199,
						BitRate:         60,
					},
				},
				{
					AlbumTrackPercentage: 0.047619047619,
					TrackInfo: common.TrackInfo{
						Path:            "/home/Music/Peggy Lee/The Legendary Peggy Lee/Disc 2/07 - Peggy Lee - Baubles, Bangles and Beads.mp3",
						Title:           "Baubles, Bangles and Beads",
						Artist:          "Peggy Lee",
						DiscNumber:      2,
						TotalDiscs:      3,
						TrackNumber:     7,
						TotalTracks:     21,
						Album:           "The Legendary Peggy Lee",
						AlbumArtist:     "Peggy Lee",
						DurationSeconds: 200,
						BitRate:         130,
					},
				},
			},
			duplicateResult{ // two closely sparse albums, one with way better bitrate, take the high bitrate
				Keep: []trackInfoExtended{
					{
						TrackInfo: common.TrackInfo{
							Path: "/home/Music/Peggy Lee/The Legendary Peggy Lee/Disc 2/07 - Peggy Lee - Baubles, Bangles and Beads.mp3",
						},
					},
				},
				Delete: []trackInfoExtended{
					{
						TrackInfo: common.TrackInfo{
							Path: "/home/Music/Various Artists/The Encyclopaedia of Music: Best of 50’s, Volume 2/12 - Peggy Lee - Baubles Bangles and Beads.MP3",
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		c.Assert(rankDuplicates(test.input), qt.CmpEquals(cmpopts.IgnoreFields(trackInfoExtended{}, "Title", "Artist", "DiscNumber", "TotalDiscs", "TrackNumber", "TotalTracks", "Album", "AlbumArtist", "DurationSeconds", "BitRate", "AlbumTrackPercentage")), test.expected)
	}
}
