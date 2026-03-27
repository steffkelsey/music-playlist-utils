package common

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
)

func PromptAndMaybeSaveFile(path string, data []byte) error {
	msg := `Save results at:
%s
`
	fmt.Printf(msg, path)
	p := promptui.Prompt{
		Label:     "Confirm",
		IsConfirm: true,
	}

	_, err := p.Run()
	if err != nil {
		return nil
	}
	// create the file at the path
	f, err := os.Create(path)
	if err != nil {
		fmt.Printf("Error creating file, %v\n", err)
		return err
	}

	// close the file when done
	defer f.Close()

	// write to the file
	_, err = f.Write(data)
	if err != nil {
		fmt.Printf("Error writing to file, %v\n", err)
		return err
	}

	fmt.Println("Report saved")

	return nil
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
			//log.Fatal(err)
		}
		fmt.Printf("- %s\n", path)
	}
	return nil
}
