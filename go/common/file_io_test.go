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

func TestCreateAbsPath(t *testing.T) {
	c := qt.New(t)
	tests := []struct {
		inputPath         string
		inputPlaylistRoot string
		expected          string
	}{
		{"../file.m4a", "/home/user/Music/playlists", "/home/user/Music/file.m4a"},
		{"../tmp/file.m4a", "/home/user/Music/playlists", "/home/user/Music/tmp/file.m4a"},
	}

	for _, test := range tests {
		c.Assert(CreateAbsPath(test.inputPath, test.inputPlaylistRoot), qt.Equals, test.expected)
	}
}

func TestMoveRelativePath(t *testing.T) {
	c := qt.New(t)
	tests := []struct {
		inputSource          string
		inputCurPlaylistDir  string
		inputDestPlaylistDir string
		expected             string
	}{
		{"./file.m4a", "/home/user/Music", "/home/user/Music/Playlists", "../file.m4a"},
		{"./file.m4a", "/home/user/Music", "/home/user/Playlists", "../Music/file.m4a"},
		{"./file.m4a", "/home/user/Music/xmas-list", "/home/user/Music", "./xmas-list/file.m4a"},
		{"./artist/album/file.m4a", "/home/user/Music/tmp", "/home/user/Music/Playlists", "../tmp/artist/album/file.m4a"},
		{"../tmp/artist/album/file.m4a", "/home/user/Music/Playlists", "/home/user/Music/tmp", "./artist/album/file.m4a"},
		{"/home/user/Music/tmp/artist/album/file.m4a", "/home/user/Music/Playlists", "/home/user/Music/tmp", "./artist/album/file.m4a"},
		{"/home/user/Music/tmp/artist/album/file.m4a", "/home/user/Music/tmp", "/home/user/Music/tmp", "./artist/album/file.m4a"},
	}

	for _, test := range tests {
		c.Assert(MoveRelativePath(test.inputSource, test.inputCurPlaylistDir, test.inputDestPlaylistDir), qt.Equals, test.expected)
	}
}
