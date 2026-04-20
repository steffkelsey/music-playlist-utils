package cmd

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/alitto/pond/v2"
	"github.com/spf13/cobra"

	"music-utils/common"
)

type exifReport struct {
	Files  []map[string]trackInfo `json:"files"`
	Albums []albumInfo            `json:"albums"`
}

type albumInfo struct {
	Album       string      `json:"album"`
	Artist      string      `json:"artist"`
	Tracks      []trackInfo `json:"tracks"`
	TotalTracks string      `json:"totalTracks"`
}

type trackInfo struct {
	Title       string `json:"title"`
	Artist      string `json:"artist"`
	TrackNumber string `json:"trackNumber"`
	TotalTracks string `json:"totalTracks"`
	Album       string `json:"album"`
}

type exifTrack struct {
	Title       string `json:"Title"`
	Artist      string `json:"Artist"`
	AlbumArtist string `json:"AlbumArtist"`
	Album       string `json:"Album"`
	TrackNumber any    `json:"TrackNumber"`
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
		Files:  make([]map[string]trackInfo, 0),
		Albums: make([]albumInfo, 0),
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

		// create the map entry for the path of the encrypted file
		m := make(map[string]trackInfo)
		// add the metadata to the trackInfo at m[p]
		ti := exifTrackToTrackInfo(t[0])
		m[p] = ti

		// see if the album exists in the Albums slice
		i, ok := albumNameToSliceIndexMap[ti.Album]
		if ok {
			// add the trackInfo to the albumInfo
			results.Albums[i].Tracks = append(results.Albums[i].Tracks, ti)
			// check that the album.totalTracks still looks good
			curAlbumTotalTracksInt, _ := strconv.Atoi(results.Albums[i].TotalTracks)
			curTrackTotalTracksInt, _ := strconv.Atoi(ti.TotalTracks)
			// Default to the biggest one
			if curAlbumTotalTracksInt < curTrackTotalTracksInt {
				results.Albums[i].TotalTracks = ti.TotalTracks
			}
		} else {
			// save the index where we added the album into the name map
			albumNameToSliceIndexMap[ti.Album] = len(results.Albums)
			tr := []trackInfo{ti}
			// create the new album in the results
			results.Albums = append(results.Albums, albumInfo{
				Album:       ti.Album,
				Artist:      t[0].AlbumArtist,
				TotalTracks: ti.TotalTracks,
				Tracks:      tr,
			})
		}

		// add the map to the files slice
		results.Files = append(results.Files, m)
	}

	// marshal the report to []byte
	j, _ := json.Marshal(&results)

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

func exifTrackToTrackInfo(i exifTrack) trackInfo {
	t := trackInfo{
		Title:       i.Title,
		Artist:      i.Artist,
		Album:       i.Album,
		TrackNumber: "",
		TotalTracks: "",
	}

	var tn string

	switch v := i.TrackNumber.(type) {
	case float64:
		tn = fmt.Sprintf("%1.f", v)
	case int:
		tn = fmt.Sprintf("%d", v)
	case string:
		tn = v
	default:
		fmt.Printf("Unknown type %T!\n", v)
	}

	// might be in the format "<track> of <total>"
	if strings.Contains(tn, "of") {
		pattern := regexp.MustCompile(`(?P<track>\w+)\sof\s+(?P<total>\w+)$`)
		match := pattern.FindSubmatch([]byte(tn))

		for i, name := range pattern.SubexpNames() {
			if i != 0 && name != "" {
				switch name {
				case "track":
					t.TrackNumber = string(match[i])
				case "total":
					t.TotalTracks = string(match[i])
				}
			}
		}

	} else {
		t.TrackNumber = tn
	}

	return t
}
