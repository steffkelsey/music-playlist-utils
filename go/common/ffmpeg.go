package common

import (
	"os/exec"
	"strings"
)

func IsFFProbeInstalled() bool {
	output, err := ffprobe("./bad-file-name")
	if err != nil {
		return strings.Contains(strings.ToLower(string(output)), "no such file")
	}
	return false
}

func ffprobe(path string) ([]byte, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=filename,bit_rate,duration", "-of", "json", path)
	return cmd.CombinedOutput()
}

func FFProbeForString(path string) (string, error) {
	s, err := ffprobe(path)
	if err != nil {
		return "", err
	}
	return string(s), nil
}

type FFProbeFormatResponse struct {
	Format FFProbeDurationAndBitRateResponse `json:"format"`
}

type FFProbeDurationAndBitRateResponse struct {
	Filename string `json:"filename"`
	Duration string `json:"duration"`
	BitRate  string `json:"bit_rate"`
}
