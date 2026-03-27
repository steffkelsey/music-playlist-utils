package cmd

import (
	"slices"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestRankDuplicates(t *testing.T) {
	c := qt.New(t)
	tests := []struct {
		rootPath string
		input    []string
		expected duplicateResult
	}{
		{
			"/home/user/Music",
			[]string{
				"/home/user/Music/album/file.mp4",
				"/home/user/Music/file.mp4",
				"/home/user/Music/file 1.mp4",
			},
			duplicateResult{
				Keep: []string{
					"/home/user/Music/album/file.mp4",
				},
				Delete: []string{
					"/home/user/Music/file.mp4",
					"/home/user/Music/file 1.mp4",
				},
			},
		},
		{
			"/home/user/Music",
			[]string{
				"/home/user/Music/file 2.mp4",
				"/home/user/Music/file 1.mp4",
				"/home/user/Music/file.mp4",
			},
			duplicateResult{
				Keep: []string{
					"/home/user/Music/file.mp4",
				},
				Delete: []string{
					"/home/user/Music/file 1.mp4",
					"/home/user/Music/file 2.mp4",
				},
			},
		},
		{
			"/home/user/Music",
			[]string{
				"/home/user/Music/file 2.mp4",
				"/home/user/Music/album/file 1.mp4",
				"/home/user/Music/album/file.mp4",
			},
			duplicateResult{
				Keep: []string{
					"/home/user/Music/album/file.mp4",
				},
				Delete: []string{
					"/home/user/Music/album/file 1.mp4",
					"/home/user/Music/file 2.mp4",
				},
			},
		},
	}

	for _, test := range tests {
		r := rankDuplicates(test.input, test.rootPath)
		c.Assert(r.Keep, qt.CmpEquals(), test.expected.Keep)
		c.Assert(r.Delete, qt.CmpEquals(), test.expected.Delete)
	}
}

func TestFilepathBaseAlphaAscSort(t *testing.T) {
	c := qt.New(t)
	tests := []struct {
		input    []string
		expected []string
	}{
		{
			[]string{
				"/home/user/Music/a.mp4",
				"/home/user/Music/a 1.mp4",
				"/home/user/Music/album/a 2.mp4",
			},
			[]string{
				"/home/user/Music/a.mp4",
				"/home/user/Music/a 1.mp4",
				"/home/user/Music/album/a 2.mp4",
			},
		},
	}

	for _, test := range tests {
		slices.SortFunc(test.input, filepathBaseAlphaAscCmp)
		c.Assert(test.input, qt.CmpEquals(), test.expected)
	}
}

func TestFilepathBaseLengthAscSort(t *testing.T) {
	c := qt.New(t)
	tests := []struct {
		input    []string
		expected []string
	}{
		{
			[]string{
				"/home/user/Music/file 1.mp4",
				"/home/user/Music/file.mp4",
				"/home/user/Music/album/file.mp4",
			},
			[]string{
				"/home/user/Music/file.mp4",
				"/home/user/Music/album/file.mp4",
				"/home/user/Music/file 1.mp4",
			},
		},
	}

	for _, test := range tests {
		slices.SortFunc(test.input, filepathBaseLengthAscCmp)
		c.Assert(test.input, qt.CmpEquals(), test.expected)
	}
}

func TestFilepathDirLengthDescSort(t *testing.T) {
	c := qt.New(t)
	tests := []struct {
		input    []string
		expected []string
	}{
		{
			[]string{
				"/home/user/Music/file 1.mp4",
				"/home/user/Music/artist/album/file.mp4",
				"/home/user/Music/album/file.mp4",
			},
			[]string{
				"/home/user/Music/artist/album/file.mp4",
				"/home/user/Music/album/file.mp4",
				"/home/user/Music/file 1.mp4",
			},
		},
	}

	for _, test := range tests {
		slices.SortFunc(test.input, filepathDirLengthDescCmp)
		c.Assert(test.input, qt.CmpEquals(), test.expected)
	}
}
