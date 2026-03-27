package common

import (
	"path/filepath"
	"strings"
)

const (
	ParamDryRun = "dry-run"
	ParamInputDir = "inputDir"
	ParamOutputDir = "outputDir"

	ExtMp3 = ".mp3"
	ExtMp4 = ".mp4"
	ExtM4a = ".m4a"
	ExtM4p = ".m4p"

	Continue = 0
	ConfirmAll = 1
	Abort = -1
)

// IsMusicFile determines if the given path points to a music file based
// solely on file extension.
//
// A successful IsMusicFile returns true.
func IsMusicFile(path string) bool {
	// Get file extension
	ext := strings.ToLower(filepath.Ext(path))
	// Check if it fits the type we care about
	if ext == ExtMp3 || ext == ExtMp4 || ext == ExtM4a || ext == ExtM4p {
		return true
	}
	return false
}
