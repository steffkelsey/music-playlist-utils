package common

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"github.com/hcl/audioduration"
)

func Bool2Float(b bool) float64 {
	return float64(Bool2int(b))
}

func Bool2int(b bool) int {
	return int(*(*byte)(unsafe.Pointer(&b)))
}

func GetDuration(path string) (float64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0.0, err
	}
	defer f.Close()
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ExtMp3:
		return audioduration.Duration(f, audioduration.TypeMp3)
	case ExtM4a:
		return audioduration.Duration(f, audioduration.TypeMp4)
	case ExtMp4:
		return audioduration.Duration(f, audioduration.TypeMp4)
	}

	return 0.0, fmt.Errorf("cannot find duration for that filetype")
}
