package common

import (
	"path/filepath"
	"strings"
)

const (
	ParamDryRun    = "dry-run"
	ParamInputDir  = "inputDir"
	ParamInputFile = "input-file"
	ParamOutputDir = "outputDir"
	ParamRecursive = "recursive"
	ParamValidate  = "validate"

	ExtMp3 = ".mp3"
	ExtMp4 = ".mp4"
	ExtM4a = ".m4a"
	ExtM4p = ".m4p"
	ExtM3u = ".m3u"

	Continue   = 0
	ConfirmAll = 1
	Abort      = -1
)

// IsMusicFile determines if the given path points to a music file based
// solely on file extension.
//
// A successful IsMusicFile returns true.
func IsMusicFile(path string) bool {
	// Get file extension
	ext := strings.ToLower(filepath.Ext(path))
	// Check if it fits the type we care about
	if ext == ExtMp3 || ext == ExtMp4 || ext == ExtM4a || ext == ExtM4p || ext == ExtM3u {
		return true
	}
	return false
}

// IsEncryptedFile determines if the given path points to an encrypted music
// file based solely on file extension.
//
// A successful IsEncryptedFile returns true.
func IsEncryptedFile(path string) bool {
	// Get file extension
	ext := strings.ToLower(filepath.Ext(path))
	// return if is an encrypted type
	return ext == ExtM4p
}

// IsPlaylistFile determines if the given path points to a playlist
// file based solely on file extension.
//
// A successful IsPlaylistFile returns true.
func IsPlaylistFile(path string) bool {
	// Get file extension
	ext := strings.ToLower(filepath.Ext(path))
	// return if is a playlist type
	return ext == ExtM3u
}
