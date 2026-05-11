package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"music-utils/common"
)

type compareAlbumsRequest struct {
	Album1    common.AlbumInfo `json:"album1"`
	Album2    common.AlbumInfo `json:"album2"`
	CmpType   string           `json:"compareType"`
	Threshold float64          `json:"threshold"`
}

type compareAlbumsResponse struct {
	common.AlbumMatch
	TrackMatchIndexes map[int]int `json:"trackMatchIndexes"`
}

var albumsCmd = &cobra.Command{
	Use:   "albums",
	Short: "Compares two Albums and returns a score on how well they match",
	Long:  `Compares two Albums and returns a score on how well they match.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		decodedData, err = common.FlagBase64DataIsGood(data)
		if err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return compareAlbums()
	},
}

func init() {
	compareCmd.AddCommand(albumsCmd)
}

func compareAlbums() error {
	var r compareAlbumsRequest
	// unmarshal the request data
	err := json.Unmarshal(decodedData, &r)
	if err != nil {
		return err
	}
	if r.Threshold < 0.001 {
		r.Threshold = 0.85
	}
	var score float64
	var indexMatchMap map[int]int
	switch r.CmpType {
	case "ignoretracks":
		score = common.CmpAlbums(r.Album1, r.Album2)
	case "tracks":
		fallthrough
	default:
		score, indexMatchMap = common.CmpAlbumsWithTracks(r.Album1, r.Album2, 0.7)
	}

	response := compareAlbumsResponse{
		TrackMatchIndexes: indexMatchMap,
	}
	response.AlbumMatch = common.FmtAlbumMatch(r.Album1, r.Album2, score, score > r.Threshold)

	j, err := json.MarshalIndent(&response, "", "  ")
	if err != nil {
		return err
	}
	jsonString := string(j)
	fmt.Println(jsonString)

	return nil
}
