package common

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestIsJsonFile(t *testing.T) {
	c := qt.New(t)
	tests := []struct {
		input    string
		expected bool
	}{
		{"/home/user/Music/report.json", true},
		{"/home/user/Music/FILE.JSON", true},
		{"filename with spaces.JSON", true},
		{"filename no extension", false},
		{"./relative folder/track1.mp3", false},
		{"./relative folder/playlist.m3u", false},
	}

	for _, test := range tests {
		c.Assert(IsJsonFile(test.input), qt.Equals, test.expected)
	}
}

func TestIsMusicFile(t *testing.T) {
	c := qt.New(t)
	tests := []struct {
		input    string
		expected bool
	}{
		{"/home/user/Music/file.mp4", true},
		{"/home/user/Music/FILE.MP4", true},
		{"filename with spaces.MP3", true},
		{"test.m4a", true},
		{"test2.m4p", true},
		{"filename no extension", false},
		{"./relative folder/track1.mp3", true},
	}

	for _, test := range tests {
		c.Assert(IsMusicFile(test.input), qt.Equals, test.expected)
	}
}
