package common

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
)

type FileMovedResult struct {
	Source string `json:"source"`
	Dest   string `json:"dest"`
}

func PromptAndMaybeSaveFile(path string, data []byte, message string) (bool, error) {
	fmt.Println(message)
	p := promptui.Prompt{
		Label:     "Confirm",
		IsConfirm: true,
	}

	// If NOT confirmed, p.Run() returns an error
	_, err := p.Run()
	if err != nil {
		return false, nil
	}
	// Confirmed, code path continues

	// create the file at the path
	f, err := os.Create(path)
	if err != nil {
		return false, err
	}

	// close the file when done
	defer f.Close()

	// write to the file
	_, err = f.Write(data)
	if err != nil {
		return false, err
	}

	return true, nil
}

func FindFileNameNoOverWrite(path string) string {
	return findFileNameNoOverWriteWithNums(path, 0)
}

// When num is 0, returns the path if the file does not exist
// When num > 0, returns Dir(path)/filename(num).Ext(path)
// For example:
// path = /home/Music/file.json
// If it does exist, then try
// path = /home/Music/file(1).json
func findFileNameNoOverWriteWithNums(path string, num int) string {
	suffix := ""
	if num > 0 {
		suffix = fmt.Sprintf("(%d)", num)
	}
	newFilename := filepath.Base(path)
	newFilename = strings.TrimSuffix(newFilename, filepath.Ext(path))
	newFilename = fmt.Sprintf("%s%s%s", newFilename, suffix, filepath.Ext(path))
	maybeGoodPath := filepath.Join(filepath.Dir(path), newFilename)
	_, err := os.Stat(maybeGoodPath)
	if err != nil {
		return maybeGoodPath
	}
	num++
	return findFileNameNoOverWriteWithNums(path, num)
}

func DeleteFiles(paths []string) error {
	for _, path := range paths {
		err := os.Remove(path)
		if err != nil {
			return err
		}
		fmt.Printf("- %s\n", path)
	}
	return nil
}

func CopyFile(from string, to string) error {
	// create the directory in the "to" location
	err := os.MkdirAll(filepath.Dir(to), 0755)
	if err != nil {
		return err
	}
	// create and open the file at the "to" location
	t, err := os.Create(to)
	if err != nil {
		return err
	}
	// open the file at the "from" path
	f, err := os.Open(from)
	if err != nil {
		return err
	}
	// copy "from" -> "to"
	_, err = io.Copy(t, f)
	if err != nil {
		return err
	}
	// close all open files
	err = t.Close()
	if err != nil {
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}
	return nil
}

func SwapRoot(path string, oldRoot string, newRoot string) string {
	rel, err := filepath.Rel(oldRoot, path)
	if err != nil {
		fmt.Printf("error finding relative path %v\n", err)
		fmt.Printf("Dir(): %s\n", filepath.Dir(path))
	}

	return filepath.Join(newRoot, rel)
}

func Sanitize(path *string) {
	*path = strings.TrimPrefix(*path, "\"")
	*path = strings.TrimSuffix(*path, "\"")
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func CreateAbsPath(dest string, listRoot string) string {
	return filepath.Join(listRoot, dest)
}

func MoveRelativePath(source string, curPlaylistDir string, destPlaylistDir string) string {
	var rel string
	if !filepath.IsAbs(source) {
		absMusicFilePath := CreateAbsPath(source, curPlaylistDir)
		rel, _ = filepath.Rel(destPlaylistDir, absMusicFilePath)
	} else {
		rel, _ = filepath.Rel(destPlaylistDir, source)
	}

	if !strings.HasPrefix(rel, "../") {
		rel = fmt.Sprintf("./%s", rel)
	}

	return rel
}

func RemoveEmptyDirectories(root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// os.Remove only deletes if the directory is empty
			os.Remove(path)
		}
		return nil
	})
}
