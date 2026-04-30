package cmd

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/alitto/pond/v2"
	"github.com/spf13/cobra"

	"music-utils/common"
)

type exifReport struct {
	Files  map[string]common.TrackInfo `json:"files"`
	Albums []common.AlbumInfo          `json:"albums"`
}

type exifTrack struct {
	Title       string `json:"Title"`
	Artist      string `json:"Artist"`
	AlbumArtist string `json:"AlbumArtist"`
	Album       string `json:"Album"`
	DiscNumber  any    `json:"DiskNumber"`
	TrackNumber any    `json:"TrackNumber"`
	Duration    string `json:"Duration"`
}

var exifCmd = &cobra.Command{
	Use:   "exif",
	Short: "Exports exif metadata into a json report",
	Long: `Exports exif metadata into a json report.
Exports for one or more encrypted music files found
in the inputDir and exports json optimized for finding
the same tracks online (grouped by album).

To export exif data for all encrypted files in the inputDir:

music-utils encrypted exif -i $HOME/Music/encrypted -o $HOME/Music/encrypted
`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		// Verify that input dir exists (also expands the path)
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
		// make sure exiftool is installed
		if !isExiftoolInstalled() {
			return fmt.Errorf("exiftool required for this command")
		}
		return findExifData()
	},
}

func init() {
	encryptedCmd.AddCommand(exifCmd)
}

func findExifData() error {
	// find all the encrypted files
	w, err := common.WalkAllMusicFiles(inputDir, isFileEncrypted)
	if err != nil {
		return err
	}

	// create a var to hold results
	results := exifReport{
		Files:  make(map[string]common.TrackInfo),
		Albums: make([]common.AlbumInfo, 0),
	}

	// Create a pool with a result type of string
	pool := pond.NewResultPool[string](100)

	// Create a task group
	group := pool.NewGroup()

	// iterate over the results, calling exiftool on each concurrently
	for _, p := range w.Files {
		group.SubmitErr(func() (string, error) {
			return exiftoolForString(p)
		})
	}

	// wait for all the responses to complete
	responses, err := group.Wait()

	if err != nil {
		fmt.Printf("Failed to get exif data: %v", err)
		return err
	}

	albumNameToSliceIndexMap := make(map[string]int)

	for i, str := range responses {
		p := w.Files[i]
		t := []exifTrack{}
		err = json.Unmarshal([]byte(str), &t)
		if err != nil {
			return err
		}

		// create TrackInfo from the exif data
		ti := exifTrackToTrackInfo(t[0])
		results.Files[p] = ti

		// see if the album exists in the Albums slice
		i, ok := albumNameToSliceIndexMap[ti.GetAlbumKey()]
		if ok {
			// add the TrackInfo to the AlbumInfo
			results.Albums[i].Tracks = append(results.Albums[i].Tracks, ti)
			// Default to the biggest one for total discs
			if results.Albums[i].TotalDiscs < ti.TotalDiscs {
				results.Albums[i].TotalDiscs = ti.TotalDiscs
			}
		} else {
			// save the index where we added the album into the name map
			albumNameToSliceIndexMap[ti.GetAlbumKey()] = len(results.Albums)
			tr := []common.TrackInfo{ti}
			// create the new album in the results
			results.Albums = append(results.Albums, common.AlbumInfo{
				Album:      ti.Album,
				Artist:     ti.AlbumArtist,
				TotalDiscs: ti.TotalDiscs,
				Tracks:     tr,
			})
		}
	}

	// sort the tracks in each album
	for _, album := range results.Albums {
		slices.SortFunc(album.Tracks, common.CmpTrackInfoDiscAndTrackNum)
	}

	// marshal the report to []byte
	j, _ := json.MarshalIndent(&results, "", "  ")

	if isDryRun {
		// print the json report to stdout
		fmt.Println(string(j))
		return nil
	}

	// create a destination for the report
	reportPath := filepath.Join(outputDir, "encrypted-exif.json")
	// We don't want to overwrite reports, so make sure the path is unique
	reportPath = common.FindFileNameNoOverWrite(reportPath)
	// ask if they want to save the json report and save it if so
	msg := fmt.Sprintf(`Save report at:
%s		
`, reportPath)
	didSave, err := common.PromptAndMaybeSaveFile(reportPath, j, msg)
	if err != nil {
		fmt.Printf("Error writing json report, %v\n", err)
		return err
	}
	if didSave {
		fmt.Println("Report saved")
	}

	return nil
}

func isExiftoolInstalled() bool {
	output, err := exiftool("./bad-file-name")
	if err != nil {
		return strings.Contains(string(output), "File not found")
	}
	return false
}

func exiftool(path string) ([]byte, error) {
	cmd := exec.Command("exiftool", "-j", path)
	cmd.Dir = inputDir
	return cmd.CombinedOutput()
}

func exiftoolForBytes(path string) ([]byte, error) {
	s, err := exiftool(path)
	if err != nil {
		return s, err
	}
	return s, nil
}

func exiftoolForString(path string) (string, error) {
	s, err := exiftoolForBytes(path)
	if err != nil {
		return "", err
	}
	return string(s), err
}

func exifTrackToTrackInfo(i exifTrack) common.TrackInfo {
	t := common.TrackInfo{
		Title:           i.Title,
		Artist:          i.Artist,
		Album:           i.Album,
		AlbumArtist:     i.AlbumArtist,
		DiscNumber:      0,
		TotalDiscs:      0,
		TrackNumber:     0,
		TotalTracks:     0,
		DurationSeconds: 0,
	}

	switch v := i.TrackNumber.(type) {
	case float64:
		t.TrackNumber = int(v)
	case int:
		t.TrackNumber = v
	case string:
		// might be in the format "<track> of <total>"
		if strings.Contains(v, "of") {
			pattern := regexp.MustCompile(`(?P<track>\w+)\sof\s+(?P<total>\w+)$`)
			match := pattern.FindSubmatch([]byte(v))

			for i, name := range pattern.SubexpNames() {
				if i != 0 && name != "" {
					switch name {
					case "track":
						t.TrackNumber, _ = strconv.Atoi(string(match[i]))
					case "total":
						t.TotalTracks, _ = strconv.Atoi(string(match[i]))
					}
				}
			}

		} else {
			t.TrackNumber, _ = strconv.Atoi(v)
		}
	default:
		fmt.Printf("Unknown type %T!\n", v)
	}

	switch disc := i.DiscNumber.(type) {
	case float64:
		t.DiscNumber = int(disc)
	case int:
		t.DiscNumber = disc
	case string:
		// might be in the format "<disc> of <total>"
		if strings.Contains(disc, "of") {
			pattern := regexp.MustCompile(`(?P<disc>\w+)\sof\s+(?P<total>\w+)$`)
			match := pattern.FindSubmatch([]byte(disc))

			for i, name := range pattern.SubexpNames() {
				if i != 0 && name != "" {
					switch name {
					case "disc":
						t.DiscNumber, _ = strconv.Atoi(string(match[i]))
					case "total":
						t.TotalDiscs, _ = strconv.Atoi(string(match[i]))
					}
				}
			}

		} else {
			t.DiscNumber, _ = strconv.Atoi(disc)
		}
	case nil:
		t.DiscNumber = 0
	default:
		fmt.Printf("Unknown type %T!\n", disc)
	}

	d, err := strconvToDuration(i.Duration)
	if err != nil {
		fmt.Printf("error converting string to duration. %v\n", err)
	} else {
		t.DurationSeconds = int(d.Seconds())
	}

	return t
}

func strconvToDuration(s string) (time.Duration, error) {
	// assume s is in format "hh:mm:ss"
	// split the string on the ":" to get components
	c := strings.Split(s, ":")
	return time.ParseDuration(fmt.Sprintf("%sh%sm%ss", c[0], c[1], c[2]))
}
