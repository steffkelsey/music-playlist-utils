package cmd

import (
	"cmp"
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	"music-utils/common"
)

type duplicateResult struct {
	Keep   []string `json:"keep"`
	Delete []string `json:"delete"`
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
		// Verify the output directory exists
		outputDir, err = common.FlagDirectoryExists(outputDir)
		if err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		err := findDuplicateFiles(inputDir, isDryRun)
		return err
	},
}

func init() {
	rootCmd.AddCommand(duplicatesCmd)
}

func findDuplicateFiles(rootPath string, isDryRun bool) error {
	res, err := common.WalkAllMusicFiles(rootPath, dupeBySize)
	if err != nil {
		return err
	}
	var sb strings.Builder
	sb.WriteString(`{"duplicates": [`)
	act := common.Continue
loop:
	for _, v := range res.MapSizeStringSlices {
		if len(v) > 1 {
			// create a duplicateResult struct
			d := rankDuplicates(v, rootPath)
			if !isDryRun {
				switch act {
				case common.Continue:
					act, err = promptAndMaybeDelete(&d)
					if err != nil {
						fmt.Printf("Error deleting file: %v", err)
						return err
					}
					fmt.Println()
				case common.ConfirmAll:
					// delete the files without asking
					common.DeleteFiles(d.Delete)
				case common.Abort:
					break loop
				}
			}
			// Only write if out one of the slices has paths
			if len(d.Keep) > 0 || len(d.Delete) > 0 {
				j, _ := json.Marshal(&d)
				s := string(j)
				sb.WriteString(s)
				sb.WriteString(",")
			}
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
	if isDryRun {
		fmt.Println(jsonString)
	} else {
		// create a destination for the report
		reportPath := filepath.Join(outputDir, "deleted-duplicates.json")
		// We don't want to overwrite reports, so make sure the path is unique
		reportPath = common.FindFileNameNoOverWrite(reportPath)
		// ask if they want to save the json report and save it if so
		err := common.PromptAndMaybeSaveFile(reportPath, []byte(jsonString))
		if err != nil {
			fmt.Printf("Error writing json report, %v\n", err)
			return err
		}
	}
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
	return filepath.Dir(rel) != "."
}

func promptAndMaybeDelete(d *duplicateResult) (int, error) {
	fmt.Printf(`Keep:
%s	

Delete:
%s

`,
		strings.Join(d.Keep, "\n"),
		strings.Join(d.Delete, "\n"))
	prompt := promptui.Select{
		Label: "Delete?",
		Items: []string{"Yes", "No", "Confirm All. Don't Ask Again", "Quit"},
	}

	_, result, err := prompt.Run()

	if err != nil {
		// User hit ctrl-c or something
		return common.Abort, err
	}

	switch result {
	case "No":
		fmt.Printf("\nSkipping...\n")
		// Empty the result so it is not recorded
		d.Keep = []string{}
		d.Delete = []string{}
	case "Yes":
		err = common.DeleteFiles(d.Delete)
	case "Confirm All. Don't Ask Again":
		// Delete the files we just presented to the user
		err = common.DeleteFiles(d.Delete)
		// Send back that all future deletions are confirmed
		return common.ConfirmAll, err
	case "Quit":
		// Empty the result so it is not recorded
		d.Keep = []string{}
		d.Delete = []string{}
		return common.Abort, nil
	}
	return common.Continue, err
}
