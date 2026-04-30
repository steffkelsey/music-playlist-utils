package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"music-utils/common"
)

type compareTracksRequest struct {
	Track1 common.TrackInfo `json:"track1"`
	Track2 common.TrackInfo `json:"track2"`
}

type compareTracksResponse struct {
	common.TrackMatch
}

var tracksCmd = &cobra.Command{
	Use:   "tracks",
	Short: "Compares two Tracks and returns a score on how well they match",
	Long:  `Compares two Tracks and returns a score on how well they match.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		decodedData, err = common.FlagBase64DataIsGood(data)
		if err != nil {
			return err
		}
		return nil
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return compareTracks()
	},
}

func init() {
	compareCmd.AddCommand(tracksCmd)
}

func compareTracks() error {
	var r compareTracksRequest
	// unmarshal the request data
	err := json.Unmarshal(decodedData, &r)
	if err != nil {
		return err
	}
	score := common.CmpAlbumTracks(r.Track1, r.Track2)

	var response compareTracksResponse
	response.TrackMatch = common.FmtTrackMatch(r.Track1, r.Track2, score, score > 0.85)

	j, err := json.MarshalIndent(&response, "", "  ")
	if err != nil {
		return err
	}
	jsonString := string(j)
	fmt.Println(jsonString)

	return nil
}
