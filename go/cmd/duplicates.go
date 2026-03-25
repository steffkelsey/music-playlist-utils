package cmd

import (
	"fmt"
	"io/fs"
	//"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"music-utils/common"
)

type duplicateResult struct {
    Keep   string    `json:"keep"`
    Delete []string  `json:"delete"`
}

var duplicatesCmd = &cobra.Command{
	Use:   "duplicates",
	Short: "Finds duplicate music files",
	Long: `Finds duplicate music files in the given
input folder.`,
  PreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		// Verify that input dir exists
		inputDir, err = common.FlagDirectoryExists(inputDir)
		if err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		err := findDuplicateFiles(inputDir)
		return err
	},
}

func init() {
	rootCmd.AddCommand(duplicatesCmd)
}

func findDuplicateFiles(rootPath string) error {
	res, err := common.WalkAllMusicFiles(rootPath, dupeBySize)
	if err != nil {
		return err
	}
	var sb strings.Builder
	sb.WriteString(`{"duplicates": [`)
	for _, v := range res.MapSizeStringSlices {
    if len(v) > 1 {
			// create a duplicateResult struct
			// find the keeper out of all the duplicates
			// when in the same folder, keep the one with the shorter filename
			// when in different folders, keep the one that is most deeply nested
	    sb.WriteString(fmt.Sprintf("[%s],", strings.Join(v, ",")))
		}
  }
	// dump the string
	jsonString := sb.String()
	// trim any dangling comma
  jsonString = strings.TrimSuffix(jsonString, ",")
	// close the duplicates array
	jsonString += "]"
	// close the object
	jsonString += "}"
	fmt.Println(jsonString)
	return nil
}

func dupeBySize(path string, info fs.FileInfo, results *common.WalkResults) error {
	results.MapSizeStringSlices[info.Size()] = append(results.MapSizeStringSlices[info.Size()], fmt.Sprintf("\"%s\"", path))
	if len(results.MapSizeStringSlices[info.Size()]) > 1 {
		results.Count++
	}
	return nil
}

//func traceWalk(path string, info fs.FileInfo, results *common.WalkResults) error {
//	// Open the file to get more details
//	file, err := os.Open(path)
//	if err != nil {
//		fmt.Printf("Error opening file path: %s, err: %v\n", path, err)
//		return err // TODO does this stop the walking?
//	} else {
//		defer file.Close()
//		if results.MapSizeString[info.Size()] == "" {
//			results.MapSizeString[info.Size()] = path
//		} else {
//			// If the filename is an exact match
//			thisName := filepath.Base(path)
//			maybeMatchName := filepath.Base(results.MapSizeString[info.Size()])
//		  thisExt := strings.ToLower(filepath.Ext(path))
//			maybeMatchExt := strings.ToLower(filepath.Ext(results.MapSizeString[info.Size()]))
//			if thisName == maybeMatchName {
//			  //fmt.Printf("Size: %d\n", info.Size())
//		    //fmt.Printf("  Path: %s\n", path)
//			  //fmt.Printf("  Dupe: %s\n", results.MapSizeString[info.Size()])
//				results.Count++
//			} else if thisExt == maybeMatchExt {
//				// What if extensions match?
//			  //fmt.Printf("Size: %d\n", info.Size())
//		    //fmt.Printf("  Path: %s\n", path)
//			  //fmt.Printf("  Dupe: %s\n", results.MapSizeString[info.Size()])
//				results.Count++
//			}
//		}
//		//fmt.Printf("File: %s\n", filepath.Base(path))
//		//fmt.Printf("  Path: %s\n", path)
//		//fmt.Printf("  Size: %d bytes\n", info.Size())
//	}
//	return nil
//}

