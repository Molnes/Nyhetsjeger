package user_ranking

import (
	"database/sql"

	"github.com/google/uuid"
)

type UserRanking struct {
	Username  string
	Points    int
	Placement int
}

// Returns the ranking of all users who have opted in to the ranking.
func GetRanking(db *sql.DB) ([]UserRanking, error) {
	rows, err := db.Query(`
SELECT 
    CONCAT(username_adjective, ' ', username_noun) AS username,
    COALESCE(total_points, 0) AS total_points,
    placement
FROM 
    users
LEFT JOIN (
    SELECT
        user_id,
        COALESCE(SUM(points_awarded), 0) AS total_points,
        RANK() OVER (ORDER BY COALESCE(SUM(points_awarded), 0) DESC) AS placement
    FROM 
        user_answers
    GROUP BY 
        user_id
) AS ua ON ua.user_id = users.id
WHERE 
    opt_in_ranking = true;

    `)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var rankings []UserRanking
	for rows.Next() {
		var ranking UserRanking
		if err := rows.Scan(&ranking.Username, &ranking.Points, &ranking.Placement); err != nil {
			return nil, err
		}
		rankings = append(rankings, ranking)

	}
	return rankings, nil
}

// Returns the ranking of the specified user.
func GetUserRanking(db *sql.DB, userID uuid.UUID) (UserRanking, error) {
	row := db.QueryRow(`
SELECT 
    CONCAT(username_adjective, ' ', username_noun) AS username,
    COALESCE(total_points, 0) AS total_points,
    placement
FROM 
    users
LEFT JOIN (
    SELECT
        user_id,
        SUM(points_awarded) AS total_points,
        RANK() OVER (ORDER BY SUM(points_awarded) DESC) AS placement
    FROM 
        user_answers
    GROUP BY 
        user_id
) AS ua ON ua.user_id = users.id
WHERE 
    opt_in_ranking = true AND users.id = $1;

    `, userID)
	ranking := UserRanking{}
	err := row.Scan(&ranking.Username, &ranking.Points, &ranking.Placement)
	if err != nil {
		return UserRanking{}, err
	}
	return ranking, nil
}
