package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// thanks to https://mholt.github.io/json-to-go/ for helping create this struct
type Event struct {
	GamePk   int    `json:"gamePk"` // game_id
	Link     string `json:"link"`   // game link
	GameData struct {
		Game struct {
			Pk int `json:"pk"` // game_id
		} `json:"game"`
	} `json:"gameData"`
	LiveData struct {
		Plays struct {
			AllPlays []struct {
				Result struct {
					Event       string `json:"event"`
					EventCode   string `json:"eventCode"`
					EventTypeID string `json:"eventTypeId"` //event type
					Description string `json:"description"`
				} `json:"result"`
				About struct {
					EventIdx            int       `json:"eventIdx"` // event_id
					EventID             int       `json:"eventId"`
					Period              int       `json:"period"` // period
					PeriodType          string    `json:"periodType"`
					OrdinalNum          string    `json:"ordinalNum"`
					PeriodTime          string    `json:"periodTime"` // period_time
					PeriodTimeRemaining string    `json:"periodTimeRemaining"`
					DateTime            time.Time `json:"dateTime"`
					Goals               struct {
						Away int `json:"away"`
						Home int `json:"home"`
					} `json:"goals"`
				} `json:"about"`
				Coordinates struct { //coords need to test
					X float32 `json:"x"` // coord_x originally not in generated struct
					Y float32 `json:"y"` //coord_y originally not in generated struct
				} `json:"coordinates"`
				Players []struct { // player 1 = Players[0], player 2 = Players[-1]
					Player struct {
						ID       int    `json:"id"` // playerx_id
						FullName string `json:"fullName"`
						Link     string `json:"link"`
					} `json:"player"`
					PlayerType string `json:"playerType"` // playerx_type
				} `json:"players,omitempty"`
			} `json:"allPlays"`
		} `json:"plays"`
	} `json:"liveData"`
}

type dbPlayers struct {
	player1ID   int
	player1Type string
	player2ID   int
	player2Type string
}

func GetEvents(gameID string) error {
	apiURL := fmt.Sprintf("https://statsapi.web.nhl.com/api/v1/game/%s/feed/live", gameID)

	client := &http.Client{}
	request, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return err
	}

	response, err := client.Do(request)
	if err != nil {
		return err
	}

	data := Event{}
	dataDec := json.NewDecoder(response.Body)
	err = dataDec.Decode(&data)
	if err != nil {
		return err
	}

	var player1ID int
	var player2ID int
	var player1Type string
	var player2Type string

	for _, cur := range data.LiveData.Plays.AllPlays {
		player1ID = 0
		player1Type = ""
		player2ID = 0
		player2Type = ""

		if len(cur.Players) >= 1 {
			player1ID = cur.Players[0].Player.ID
			player1Type = cur.Players[0].PlayerType
			if len(cur.Players) >= 2 {
				player2ID = cur.Players[len(cur.Players)-1].Player.ID
				player2Type = cur.Players[len(cur.Players)-1].PlayerType
			}
		}

		q := `INSERT INTO event (event_id, event_type, player1_id, player2_id,
		                player1_type, player2_type, coord_x, coord_y, period, period_time, game_id)
		                        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
		_, err := Db.Exec(q, cur.About.EventIdx, cur.Result.EventTypeID, player1ID, player2ID,
			player1Type, player2Type, cur.Coordinates.X, cur.Coordinates.Y, cur.About.Period, cur.About.PeriodTime, data.GamePk)
		if err != nil {
			if IsUniqueViolation(err) {
				continue
			}
			log.Fatal(err)
		}
	}

	return nil
}