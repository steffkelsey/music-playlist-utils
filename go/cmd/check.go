package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/dhowden/tag"

	"github.com/spf13/cobra"

	"music-utils/common"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Indicates if the required metadata exist in each file",
	Long: `Indicates if the required metadata exists in each music file 
in the given input directory.	
Music files must have a mp3, mp4, m4a, or m4p extension.
eg:
Check all the files in the root folder $HOME/Music and export any without tags to $HOME/no-tags
music-utils check -i "$HOME/Music" -o "$HOME/no-tags"`,
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
		err := findUntaggedFiles(inputDir)
		return err
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}


func getMusicMetaDataForAllFilesInFolder(folder string) (map[string]bool, error) {
	fmt.Printf("Reading files from folder: %s\n", folder)
	fmt.Println("------------------")

	results := make(map[string]bool)
	numTracks := 0

	// Walk through all files in the folder and subfolders
	err := filepath.Walk(folder, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error accessing path %s: %v\n", path, err)
			return nil // Continue walking
		}

		// Skip directories
		if info.IsDir() && path != folder {
			//fmt.Printf("Directory: %s\n", path)
			//fmt.Println("------------------")
			return nil
		}

		// Skip the root folder itself
		if path == folder {
			return nil
		}

		// Get the most likely Title and Artist from the file name
		// which uses the convention SpotiDownloader.com - Title - Artist.extension
		//expectedTitle, expectedArtist := getTitleAndArtistFromFilename(filepath.Base(path))
		// Replace '_' with '/' in both the title and the artist
		//expectedTitle = strings.ReplaceAll(expectedTitle, "_", "/")
		//expectedArtist = strings.ReplaceAll(expectedArtist, "_", "/")

		//if results[path] {
		//	fmt.Printf("  File already processed with title: %s\n", path)
		//	return nil
		//}

		// If it's an MP3 or MP4 file, extract metadata using dhowden/tag
		if common.IsMusicFile(path) {
			numTracks++
			// Open the file to get more details
			file, err := os.Open(path)
			if err != nil {
				fmt.Printf("  Error opening file: %v\n", err)
			} else {
				defer file.Close()

				// Use dhowden/tag to read metadata
				m, err := tag.ReadFrom(file)
				if err != nil {
					// Print file information
					fmt.Printf("File: %s\n", filepath.Base(path))
					fmt.Printf("  Path: %s\n", path)
					fmt.Printf("  Size: %d bytes\n", info.Size())
					//fmt.Printf("  Mode: %s\n", info.Mode())
					//fmt.Printf("  Modified: %s\n", info.ModTime().Format(time.RFC1123))

					fmt.Printf("  Error reading metadata: %v\n", err)
					fmt.Println("------------------")
					results[path] = true
				} else {
					tagErrors := ""
					// Must have Album, Title, Track, Artist
					if m.Album() == "" {
						tagErrors += "| Album "
						results[path] = true
					}

					if m.Title() == "" {
						tagErrors += "| Title "
						results[path] = true
					}

					if m.Artist() == "" {
						tagErrors += "| Artist "
						results[path] = true
					}

					trackNum, _ := m.Track()
					if trackNum == 0 {
						tagErrors += "| Track "
						results[path] = true
					}

					if results[path] {
					  fmt.Printf("File: %s\n", filepath.Base(path))
					  fmt.Printf("  Path: %s\n", path)
					  fmt.Printf("  TagErrors: %s\n", tagErrors)
					}



					//if m.Title() != expectedTitle || m.Artist() != expectedArtist {
					//	// Print file information
					//	fmt.Printf("File: %s\n", filepath.Base(path))
					//	fmt.Printf("  Path: %s\n", path)
					//	fmt.Printf("  Size: %d bytes\n", info.Size())
					//	fmt.Printf("  Mode: %s\n", info.Mode())
					//	fmt.Printf("  Modified: %s\n", info.ModTime().Format(time.RFC1123))

					//	fmt.Printf("  Warning: Title and Artist do not match the expected values\n")
					//	// Display metadata
					//	fmt.Printf("  Title: %s != %s\n", m.Title(), expectedTitle)
					//	fmt.Printf("  Artist: %s != %s\n", m.Artist(), expectedArtist)
					//	fmt.Printf("  Album: %s\n", m.Album())
					//	fmt.Printf("  Year: %d\n", m.Year())
					//	fmt.Printf("  Genre: %s\n", m.Genre())
					//	fmt.Println("------------------")
					//}
				}
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking through folder: %v\n", err)
	}

	fmt.Printf("\n Found %d files WITHOUT metadata out of %d tracks\n", len(results), numTracks)
	return results, nil
}

func findUntaggedFiles(rootPath string) error {
	res, err := common.WalkAllMusicFiles(rootPath, isFileUntagged)
	if err != nil {
		return err
	}
	var sb strings.Builder
	sb.WriteString(`{"untagged": `)
	sb.WriteString(fmt.Sprintf("[%s]", strings.Join(res.Files, ",")))
	// dump the string
	jsonString := sb.String()
	// close the object
	jsonString += "}"
	fmt.Println(jsonString)
	return nil
}

func isFileUntagged(path string, info fs.FileInfo, results *common.WalkResults) error {
	// Open the file to get more details
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("  Error opening file: %v\n", err)
	} else {
		defer file.Close()

		// Use dhowden/tag to read metadata
		m, err := tag.ReadFrom(file)
		if err != nil {
			results.Files = append(results.Files, fmt.Sprintf("\"%s\"", path))
			results.Count++
		} else {
			isTagGood := true
			// Must have Album, Title, Track, Artist
			if m.Album() == "" {
				isTagGood = false
			}

			if m.Title() == "" {
				isTagGood = false
			}

			if m.Artist() == "" {
				isTagGood = false
			}

			trackNum, _ := m.Track()
			if trackNum == 0 {
				isTagGood = false
			}

			if !isTagGood {
				results.Files = append(results.Files, fmt.Sprintf("\"%s\"", path))
				results.Count++
			}
		}
	}

	return nil
}
