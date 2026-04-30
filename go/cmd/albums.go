package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"music-utils/common"
)

type compareAlbumsRequest struct {
	Album1 common.AlbumInfo `json:"album1"`
	Album2 common.AlbumInfo `json:"album2"`
}

type compareAlbumsResponse struct {
	common.AlbumMatch
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
	score := common.CmpAlbums(r.Album1, r.Album2)

	var response compareAlbumsResponse
	response.AlbumMatch = common.FmtAlbumMatch(r.Album1, r.Album2, score, score > 0.85)

	j, err := json.MarshalIndent(&response, "", "  ")
	if err != nil {
		return err
	}
	jsonString := string(j)
	fmt.Println(jsonString)

	return nil
}
