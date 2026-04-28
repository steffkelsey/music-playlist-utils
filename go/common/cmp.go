package common

import (
	"math"
	"strings"
	"unicode"
)

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

func CmpAlbumTracks(t1 TrackInfo, t2 TrackInfo) float64 {
	titleScore := Bool2Float(IsExactMatch(t1.Title, t2.Title))
	albumScore := Bool2Float(IsExactMatch(t1.Album, t2.Album))
	artistScore := Bool2Float(IsExactMatch(t1.Artist, t2.Artist))
	albumArtistScore := Bool2Float(IsExactMatch(t1.AlbumArtist, t2.AlbumArtist))
	trackNumberScore := Bool2Float(t1.TrackNumber == t2.TrackNumber)
	totalTracksScore := Bool2Float(t1.TotalTracks == t2.TotalTracks)

	sum := titleScore + albumScore + artistScore + albumArtistScore + trackNumberScore + totalTracksScore
	perfect := 6.0

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

	// both track number and total tracks must match or throw them out
	if trackNumberScore < 1.0 || totalTracksScore < 1.0 {
		trackNumberScore = 0
		totalTracksScore = 0
	}

	// how to weight the rest?
	// AlbumArtist is least critical

	// Since Titles by Artists can match across albums because
	// of compilations (greatest hits, soundtracks, etc)
	// this matters the most if we are just trying to find the same
	// track BUT i would really like the duration to be close because
	// there can be very different versions.
	// The question is, when should duration get involved because
	// we have to calculate it (not commonly found in the tags)

	// If we are leaning toward just the matching sound file, then the
	// above matters most

	// If we are leaning toward matching the track to a specific album,
	// then the weight is more equal

	// Rubber duck says to make this function be about matching best
	// to an album and we should make another function that includes
	// calulating duration to check if the track is a match across
	// albums

	// A Very Likely match
	//Title - exact
	//TrackNumber - exact
	//TotalTracks - exact
	//
	//Artist - high fuzzy
	//Album - high fuzzy
	//AlbumArtist - high fuzzy

	// A likely match
	//Title - high fuzzy
	//Artist - high fuzzy
	//Album - high fuzzy
	//and the rest is bonus

	//fmt.Printf("'%s' | '%s': %.2f\n", t1.Title, t2.Title, titleScore)
	//fmt.Printf("'%s' | '%s': %.2f\n", t1.Album, t2.Album, albumScore)
	//fmt.Printf("'%s' | '%s': %.2f\n", t1.Artist, t2.Artist, artistScore)
	//fmt.Printf("'%s' | '%s': %.2f\n", t1.AlbumArtist, t2.AlbumArtist, albumArtistScore)
	//fmt.Printf("%d | %d: %.2f\n", t1.TrackNumber, t2.TrackNumber, trackNumberScore)
	//fmt.Printf("%d | %d: %.2f\n", t1.TotalTracks, t2.TotalTracks, totalTracksScore)

	sum = titleScore + albumScore + artistScore + albumArtistScore + trackNumberScore + totalTracksScore

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
func CmpTracks(t1 TrackInfo, t2 TrackInfo) float64 {
	titleScore := Bool2Float(IsExactMatch(t1.Title, t2.Title))
	artistScore := Bool2Float(IsExactMatch(t1.Artist, t2.Artist))

	// duration score is linear falloff where 10 seconds difference (and greater) in duration
	// results in 0.0 and 0 seconds diff in duration is a 1.0
	var durationScore float64
	diff := t1.DurationSeconds - t2.DurationSeconds
	if diff == 0 {
		durationScore = 1.0
	} else {
		durationScore = 1 - math.Min((math.Abs(float64(diff))/10.0), 1.0)
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

func CmpAlbumInfoAlbumTitle(a, b AlbumInfo) int {
	return strings.Compare(strings.ToLower(a.Album), strings.ToLower(b.Album))
}

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
