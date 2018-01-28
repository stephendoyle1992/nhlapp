package main

import (
	"fmt"
)

func getTeamShots(gameID string, team string) (int, error) {
	if gameID == "" || team == "" {
		return 0, fmt.Errorf("game or team not given")
	}
	q := `select count (*) from event where game_id = $1 and player1_team = $2 and (event_type = 'SHOT' or event_type = 'GOAL')`

	row := Db.QueryRow(q, gameID, team)
	/*if err != nil {
		return 0, err
	}*/

	var count int
	row.Scan(&count)

	return count, nil
}

func getLineShots(gameID string) error {
	q := `select distinct roster.player_id, roster.event_id
            from event_roster as roster,
                (select event_id, player1_team as team
                    from event
                    where game_id = $1 and (event_type = 'SHOT' or event_type = 'GOAL')
                ) as sub_q
            where roster.game_id = $1 and roster.event_id = sub_q.event_id and roster.team = sub_q.team
            order by roster.event_id asc`

	_, err := Db.Query(q, gameID)
	if err != nil {
		return err
	}

	return nil
}
