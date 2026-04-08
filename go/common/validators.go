package common

import (
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
		return "", fmt.Errorf("directory does not exist at %s", flagDir)
	}
	return flagDir, nil
}
