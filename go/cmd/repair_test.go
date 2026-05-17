package cmd

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"music-utils/common"
)

func TestMovedReportRemoveDuplicates(t *testing.T) {
	c := qt.New(t)
	tests := []struct {
		input    movedReport
		expected movedReport
	}{
		{
			// expect one dupe removed
			movedReport{Moved: []common.FileMovedResult{
				{
					Source: "/a/b/c.mp3",
					Dest:   "/c/d/f.mp3",
				},
				{
					Source: "/a/b/c.mp3",
					Dest:   "/c/d/f.mp3",
				},
			}},
			movedReport{Moved: []common.FileMovedResult{
				{
					Source: "/a/b/c.mp3",
					Dest:   "/c/d/f.mp3",
				},
			}},
		},
		{
			// expect sorted but no deletions when no dupes and unsorted
			movedReport{Moved: []common.FileMovedResult{
				{
					Source: "/a/b/z.mp3",
					Dest:   "/c/d/f.mp3",
				},
				{
					Source: "/a/b/c.mp3",
					Dest:   "/c/d/f.mp3",
				},
			}},
			movedReport{Moved: []common.FileMovedResult{
				{
					Source: "/a/b/c.mp3",
					Dest:   "/c/d/f.mp3",
				},
				{
					Source: "/a/b/z.mp3",
					Dest:   "/c/d/f.mp3",
				},
			}},
		},
	}

	for _, test := range tests {
		test.input.RemoveDuplicates()
		c.Assert(test.input, qt.DeepEquals, test.expected)
	}
}
