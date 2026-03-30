package common

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestSwapRoot(t *testing.T) {
	c := qt.New(t)
	tests := []struct {
		inputPath    string
		inputOldRoot string
		inputNewRoot string
		expected     string
	}{
		{"/home/user/Music/tmp/file.m4a", "/home/user/Music/tmp", "/home/user/Music/untagged", "/home/user/Music/untagged/file.m4a"},
		{"/home/user/Music/FILE.MP4", "/home/user/Music", "/a", "/a/FILE.MP4"},
		{"/a/track.mp3", "/", "/b/c", "/b/c/a/track.mp3"},
		{"/home/user/Music/tmp/filename with spaces.mp3", "/home/user/Music/tmp", "/c/d/e", "/c/d/e/filename with spaces.mp3"},
		{"/home/user/Music/tmp/Unknown Artist/Unknown Album/Yes - 90125 - 06 - Leave It.mp3", "/home/user/Music/tmp", "/home/user/Music/untagged", "/home/user/Music/untagged/Unknown Artist/Unknown Album/Yes - 90125 - 06 - Leave It.mp3"},
		{"/home/steff/Music/tmp/Unknown Artist/Unknown Album/05 - T.B. Sheets - Van Morrison - T.B. Sheets.mp3", "/home/steff/Music/tmp", "/home/steff/Music/untagged", "/home/steff/Music/untagged/Unknown Artist/Unknown Album/05 - T.B. Sheets - Van Morrison - T.B. Sheets.mp3"},
		{"/home/steff/Music/tmp/Unknown Artist/Unknown Album/track.mp3", "/home/steff/Music/tmp", "/", "/Unknown Artist/Unknown Album/track.mp3"},
	}

	for _, test := range tests {
		c.Assert(SwapRoot(test.inputPath, test.inputOldRoot, test.inputNewRoot), qt.Equals, test.expected)
	}
}
