package common

import (
	"cmp"
	"math"
	"strings"
	"unicode"
)

type AlbumMatch struct {
	Score      float64 `json:"score"`
	Titles     string  `json:"titles"`
	Artists    string  `json:"artists"`
	TotalDiscs string  `json:"totalDiscs"`
	Success    bool    `json:"success"`
}

type TrackMatch struct {
	TrackPaths   []string `json:"tracks"`
	Score        float64  `json:"score"`
	Titles       string   `json:"titles"`
	Artists      string   `json:"artists"`
	Albums       string   `json:"albums"`
	AlbumArtists string   `json:"albumArtists"`
	Durations    string   `json:"durations"`
	TrackNumbers string   `json:"trackNumbers"`
	TotalTracks  string   `json:"totalTracks"`
	DiscNumbers  string   `json:"discNumbers"`
	TotalDiscs   string   `json:"totalDiscs"`
	Success      bool     `json:"success"`
}

func IsExactMatch(s1 string, s2 string) bool {
	return strings.EqualFold(s1, s2)
}

func IsFuzzyMatch(s1 string, s2 string) (float64, float64) {
	f := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	}

	if len(s1) == 0 || len(s2) == 0 {
		return 0.0, 0.0
	}

	s1 = strings.ToLower(s1)
	s2 = strings.ToLower(s2)

	f1 := strings.FieldsFunc(s1, f)
	f2 := strings.FieldsFunc(s2, f)

	// sort them so the longest is in f1
	if len(f1) < len(f2) {
		f1 = strings.FieldsFunc(s2, f)
		f2 = strings.FieldsFunc(s1, f)
	}

	// map for storing all the words in f1
	map1 := make(map[string]struct{})

	// iterate over f1 to fill in the map
	for _, w := range f1 {
		map1[w] = struct{}{}
	}

	// find the percentage of words from f2 that are found in f1
	matches := 0
	for _, w := range f2 {
		_, ok := map1[w]
		if ok {
			matches++
		}
	}

	r1 := float64(matches) / float64(len(f1))

	r2 := SubstrMagic(f1, f2)

	return r1, r2
}

func CmpAlbums(a1, a2 AlbumInfo) float64 {
	titleScore := Bool2Float(IsExactMatch(a1.Album, a2.Album))
	artistScore := Bool2Float(IsExactMatch(a1.Artist, a2.Artist))
	totalDiscsScore := Bool2Float(a1.TotalDiscs == a2.TotalDiscs)

	sum := titleScore + artistScore + totalDiscsScore
	perfect := 3.0

	// A perfect match = exact matches on all
	if sum < perfect {
		// A perfect WORST match, return 0.0
		if sum-0.001 < 0 {
			return 0.0
		}
	} else {
		// A perfect match, return 1.0
		return 1.0
	}

	// The only scores that can be fuzzy are string based
	// keep perfect scores
	if titleScore < 1.0 {
		s1, s2 := IsFuzzyMatch(a1.Album, a2.Album)
		titleScore = (s1 + s2) * 0.5
	}
	if artistScore < 1.0 {
		s1, s2 := IsFuzzyMatch(a1.Artist, a2.Artist)
		artistScore = (s1 + s2) * 0.5
	}

	sum = titleScore + artistScore + totalDiscsScore
	return sum / perfect
}

func CmpAlbumTracks(t1, t2 TrackInfo) float64 {
	titleScore := Bool2Float(IsExactMatch(t1.Title, t2.Title))
	albumScore := Bool2Float(IsExactMatch(t1.Album, t2.Album))
	artistScore := Bool2Float(IsExactMatch(t1.Artist, t2.Artist))
	albumArtistScore := Bool2Float(IsExactMatch(t1.AlbumArtist, t2.AlbumArtist))
	trackNumberScore := Bool2Float(t1.TrackNumber == t2.TrackNumber)
	totalTracksScore := Bool2Float(t1.TotalTracks == t2.TotalTracks)
	discNumberScore := Bool2Float(t1.DiscNumber == t2.DiscNumber)
	totalDiscsScore := Bool2Float(t1.TotalDiscs == t2.TotalDiscs)

	sum := titleScore + albumScore + artistScore + albumArtistScore + trackNumberScore + totalTracksScore + discNumberScore + totalDiscsScore
	perfect := 8.0

	// A perfect match = exact matches on all
	if sum < perfect {
		// A perfect WORST match, return 0.0
		if sum-0.001 < 0 {
			return 0.0
		}
	} else {
		// A perfect match, return 1.0
		return 1.0
	}

	// The only scores that can be fuzzy are string based
	// keep perfect scores
	if titleScore < 1.0 {
		s1, s2 := IsFuzzyMatch(t1.Title, t2.Title)
		titleScore = (s1 + s2) * 0.5
	}
	if albumScore < 1.0 {
		s1, s2 := IsFuzzyMatch(t1.Album, t2.Album)
		albumScore = (s1 + s2) * 0.5
	}
	if artistScore < 1.0 {
		s1, s2 := IsFuzzyMatch(t1.Artist, t2.Artist)
		artistScore = (s1 + s2) * 0.5
	}
	if albumArtistScore < 1.0 {
		s1, s2 := IsFuzzyMatch(t1.AlbumArtist, t2.AlbumArtist)
		albumArtistScore = (s1 + s2) * 0.5
	}

	sum = titleScore + albumScore + artistScore + albumArtistScore + trackNumberScore + totalTracksScore + discNumberScore + totalDiscsScore

	//fmt.Printf("score: %.2f\n", sum/6.0)
	return sum / perfect
}

// CmpTracks is concerned with if two tracks match that do
// NOT come from the same album yet are still the same track.
// The use cases this covers are one track is from the studio
// album and the other is from a Greatest Hits compilation
// or a movie soundtrack where the album and album artist
// are completely different but the track is the same.
// Basically, the Track.Title, Track.Artist and DurationSeconds
// should be very close to matching.
func CmpTracks(t1, t2 TrackInfo) float64 {
	titleScore := Bool2Float(IsExactMatch(t1.Title, t2.Title))
	artistScore := Bool2Float(IsExactMatch(t1.Artist, t2.Artist))

	// duration score is linear falloff where 10 seconds difference (and greater) in duration
	// results in 0.0 and 0 seconds diff in duration is a 1.0
	var durationScore float64
	diff := t1.DurationSeconds - t2.DurationSeconds
	if diff == 0 {
		durationScore = 1.0
	} else {
		durationScore = 1 - math.Min((math.Abs(float64(diff))/11.0), 1.0)
	}

	sum := titleScore + artistScore + durationScore
	perfect := 3.0

	// A perfect match = exact matches on all
	if sum < perfect {
		// A perfect WORST match, return 0.0
		if sum-0.001 < 0 {
			return 0.0
		}
	} else {
		// A perfect match, return 1.0
		return 1.0
	}

	// The only scores that can be fuzzy are string based
	// keep perfect scores
	if titleScore < 1.0 {
		s1, s2 := IsFuzzyMatch(t1.Title, t2.Title)
		titleScore = (s1 + s2) * 0.5
	}
	if artistScore < 1.0 {
		s1, s2 := IsFuzzyMatch(t1.Artist, t2.Artist)
		artistScore = (s1 + s2) * 0.5
	}

	sum = titleScore + artistScore + durationScore

	return sum / perfect
}

// For sorting
func CmpAlbumInfoAlbumTitle(a, b AlbumInfo) int {
	return strings.Compare(strings.ToLower(a.Album), strings.ToLower(b.Album))
}

// For sorting
func CmpTrackInfoDiscAndTrackNum(a, b TrackInfo) int {
	// Sort by discNumber first
	if n := cmp.Compare(a.DiscNumber, b.DiscNumber); n != 0 {
		return n
	}
	// If disc number is equal, compare by track number
	return cmp.Compare(a.TrackNumber, b.TrackNumber)
}

// For sorting
func CmpTrackInfoTitle(a, b TrackInfo) int {
	return strings.Compare(strings.ToLower(a.Title), strings.ToLower(b.Title))
}

func SubstrMagic(a1, a2 []string) float64 {
	a1str := strings.Join(a1, "")
	a2str := strings.Join(a2, "")

	// record the score if we check whole string versus whole string
	r := scoreSub(a1str, a2str)

	if len(a1) < 2 || len(a2) < 2 {
		return r
	}

	// iterate over the words in a1
	for i := range a1 {
		t := make([]string, 0)
		t = append(t, a1[:i]...)
		t = append(t, a1[i+1:]...)
		tstr := strings.Join(t, "")
		r2 := scoreSub(tstr, a2str)
		if r < r2 {
			r = r2
		}
	}

	// iterate over the words in a2
	for i := range a2 {
		t := make([]string, 0)
		t = append(t, a2[:i]...)
		t = append(t, a2[i+1:]...)
		tstr := strings.Join(t, "")
		r2 := scoreSub(tstr, a1str)
		if r < r2 {
			r = r2
		}
	}

	return r
}

func scoreSub(s1, s2 string) float64 {
	r := 0.0

	// sort so longest is in s1
	if len(s1) < len(s2) {
		t := s2
		s2 = s1
		s1 = t
	}

	i := strings.Index(s1, s2)
	if i > -1 {
		r = r + float64(len(s2))/float64(len(s1))
	}
	if i == 0 {
		// bonus for matching at index 0
		r = r + 0.25
	}
	// cap  at 1.0
	if r > 1.0 {
		r = 1.0
	}
	return r

}

// For sorting
func CmpTrackMatchScore(a, b TrackMatch) int {
	return -1 * cmp.Compare(a.Score, b.Score)
}
