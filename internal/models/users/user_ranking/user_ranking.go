package user_ranking

import "database/sql"

type UserRanking struct {
	Username string
	Points int
}

// Returns the ranking of all users who have opted in to the ranking.
func GetRanking(db *sql.DB) ([]UserRanking, error) {
	rows, err := db.Query(`
        SELECT CONCAT(username_adjective, ' ', username_noun) AS Username, pt.sum_points
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
		if err := rows.Scan(&ranking.Username, &ranking.Points); err != nil {
                        return nil, err
                }
                rankings = append(rankings, ranking)

        }
        return rankings, nil
}
