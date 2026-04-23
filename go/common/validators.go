package common

import (
	"fmt"
	"os"
	"strings"
	"unicode"
)

// FlagDirectoryExists is a cobra flag validator to be included
// in PreRunE or RunE functions to check if a flag input
// that points to a directory exists in the file system.
func FlagDirectoryExists(flagDir string) (string, error) {
	// Use ExpandEnv to substitute any possible OS environment vars
	flagDir = os.ExpandEnv(flagDir)
	_, err := os.Stat(flagDir)
	if err != nil {
		return "", fmt.Errorf("file or directory does not exist at %s", flagDir)
	}
	return flagDir, nil
}

func IsExactMatch(s1 string, s2 string) bool {
	return strings.EqualFold(s1, s2)
}

func IsFuzzyMatch(s1 string, s2 string) (float64, float64) {
	f := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	}

	r2 := 0.0

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
	map1 := make(map[string]bool)

	// iterate over f1 to fill in the map
	for _, w := range f1 {
		map1[w] = true
	}

	// find the percentage of words from f2 that are found in f1
	matches := 0
	for _, w := range f2 {
		_, ok := map1[w]
		if ok {
			matches++
		}
	}

	r1 := float64(matches) / float64(len(f2))

	// now find what percentage a substring that the lowercase, joined string of f2 is of f1
	f1str := strings.Join(f1, "")
	f2str := strings.Join(f2, "")

	i := float64(strings.Index(f1str, f2str))
	if i > -1 {
		r2 = float64(len(f2)) / float64(len(f1))
	}

	return r1, r2
}
