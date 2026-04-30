package common

import (
	"encoding/base64"
	"fmt"
	"os"
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

func FlagBase64DataIsGood(flagData string) ([]byte, error) {
	var res []byte
	// Verify that data exists
	if len(flagData) == 0 {
		return res, fmt.Errorf("data is empty")
	}

	// Verify it is valid base64 encoding
	var err error
	res, err = base64.StdEncoding.DecodeString(flagData)
	if err != nil {
		return res, err
	}
	return res, nil
}
