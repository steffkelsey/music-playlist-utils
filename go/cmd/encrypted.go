package cmd

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"music-utils/common"
)

var encryptedCmd = &cobra.Command{
	Use:   "encrypted",
	Short: "Copies only the encrypted files to the output directory",
	Long: `Copies only encrypted music files in the given input directory 
to the given output directory preserving subfolders. Dry run shows a 
JSON report of where files would have been copied to.
Encrypted music files must have a m4p extension.
eg:
Check all the files in the root folder $HOME/Music and export any encrypted 
to $HOME/encrypted

music-utils check -i "$HOME/Music" -o "$HOME/encrypted"`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		// Verify that input dir exists
		inputDir, err = common.FlagDirectoryExists(inputDir)
		if err != nil {
			return err
		}
		// Verify the output directory exists
		outputDir, err = common.FlagDirectoryExists(outputDir)
		if err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		err := findEncryptedFiles(inputDir)
		return err
	},
}

func init() {
	rootCmd.AddCommand(encryptedCmd)
}

func findEncryptedFiles(rootPath string) error {
	res, err := common.WalkAllMusicFiles(rootPath, isFileEncrypted)
	if err != nil {
		return err
	}
	// go through the files and copy each one to the target destination
	// create the report for each
	final := make([]string, len(res.Files))
	for i, path := range res.Files {
		common.Sanitize(&path)
		newPath := common.SwapRoot(path, rootPath, outputDir)
		if !isDryRun {
			// copy to new location
			err := common.CopyFile(path, newPath)
			if err != nil {
				fmt.Printf("error copying file, err: %v\n", err)
				return err
			}
			fmt.Printf("+ %s\n", newPath)
		}
		// save for the report
		final[i] = newPath
	}
	var sb strings.Builder
	sb.WriteString(`{"encrypted": `)
	j, _ := json.Marshal(&final)
	sb.Write(j)
	// dump the string
	jsonString := sb.String()
	// close the object
	jsonString += "}"
	if isDryRun {
		fmt.Println(jsonString)
	}
	return nil
}

func isFileEncrypted(path string, info fs.FileInfo, results *common.WalkResults) error {
	// if the file is NOT encrypted, skip
	if !common.IsEncryptedFile(path) {
		return nil
	}
	// Open the file to be sure it exists
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
	} else {
		defer file.Close()

		results.Files = append(results.Files, path)
		results.Count++
	}

	return nil
}
