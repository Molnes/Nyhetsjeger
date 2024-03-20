package userranking

import "database/sql"

type UserRanking struct {
	Email  string
	Points int
}

func GetRanking(db *sql.DB) ([]UserRanking, error) {
	rows, err := db.Query(`
        SELECT email, pt.sum_points
FROM users
JOIN (
SELECT user_id, SUM(points_awarded) AS sum_points FROM user_answers
GROUP BY user_id
) AS pt ON users.id = pt.user_id
WHERE opt_in_ranking = true
ORDER BY pt.sum_points DESC;
    `)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var rankings []UserRanking
	for rows.Next() {
		var ranking UserRanking
		if err := rows.Scan(&ranking.Email, &ranking.Points); err != nil {
			return nil, err
		}
		rankings = append(rankings, ranking)
	}

	return rankings, nil
}
