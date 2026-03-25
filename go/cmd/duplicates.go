package cmd

import (
	"cmp"
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"music-utils/common"
)

type duplicateResult struct {
    Keep   []string  `json:"keep"`
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
			d := rankDuplicates(v, rootPath)
			j, _ := json.Marshal(&d)
			sb.WriteString(string(j))
			sb.WriteString(",")
	    //sb.WriteString(fmt.Sprintf("[%s],", strings.Join(v, ",")))
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
	results.MapSizeStringSlices[info.Size()] = append(results.MapSizeStringSlices[info.Size()], path)
	if len(results.MapSizeStringSlices[info.Size()]) > 1 {
		results.Count++
	}
	return nil
}

func rankDuplicates(dupes []string, rootPath string) duplicateResult {
	r := duplicateResult{}
	// make a copy of the input
  d := slices.Clone(dupes)
	// sort by filename alphabetically
	slices.SortFunc(d, filepathBaseAlphaAscCmp)
	// sort by filename length
	slices.SortFunc(d, filepathBaseLengthAscCmp)
	// sort by folder length
	slices.SortFunc(d, filepathDirLengthDescCmp)

	for i, path := range d {
		// always keep the top entry
		if i == 0 {
			r.Keep = append(r.Keep, path)
			continue
		}
		// If entry is in the same folder as the top one, delete it
		if filepath.Dir(r.Keep[0]) == filepath.Dir(path) {
			r.Delete = append(r.Delete, path)			
			continue
		}
		// if the entry is NOT in the root path folder, mark as keep
		if isInRootPath(path, rootPath) {
			r.Keep = append(r.Keep, path)
		} else {
			// otherwise mark delete
			r.Delete = append(r.Delete, path)
		}
	}

	return r
}

func filepathBaseAlphaAscCmp(a, b string) int {
	// remove the file extension before comparing
	baseA := strings.TrimSuffix(a, filepath.Ext(a))
	baseB := strings.TrimSuffix(b, filepath.Ext(b))
  return cmp.Compare(baseA, baseB)
}

func filepathBaseLengthAscCmp(a, b string) int {
  return cmp.Compare(len(filepath.Base(a)), len(filepath.Base(b)))
}

func filepathDirLengthDescCmp(a, b string) int {
  return -cmp.Compare(len(filepath.Dir(a)), len(filepath.Dir(b)))
}

func isInRootPath(path string, rootPath string) bool {
	rel, _ := filepath.Rel(rootPath, path)
	if filepath.Dir(rel) == "." {
		return false
	} 
	return true
}

