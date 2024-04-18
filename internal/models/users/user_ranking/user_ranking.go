package user_ranking

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type UserRanking struct {
	User_id   uuid.UUID
	Username  string
	Points    int
	Placement int
}

type DateRange int

const (
	All   DateRange = 0
	Month DateRange = 1
	Year  DateRange = 2
)

// Returns the ranking of all users who have opted in to the ranking.
func GetRanking(db *sql.DB, month time.Month, year int, timeZone *time.Location, dateRange DateRange) ([]UserRanking, error) {

	firstMoment, lastMoment := GetDateRange(dateRange, timeZone, year, month)

	rows, err := db.Query(`
        SELECT user_id, SUM(total_points_awarded) AS total_points, 
CONCAT(u.username_adjective, ' ', u.username_noun) AS username,

 RANK() OVER (ORDER BY SUM(total_points_awarded) DESC) as ranking
FROM "user_quizzes"
JOIN (
SELECT * FROM quizzes
) as q ON q.id = quiz_id

JOIN (
SELECT * FROM users
) as u ON u.id = user_id

WHERE
is_completed = true
AND answered_within_active_time = true

AND q.published = true
AND q.is_deleted = false
AND q.active_from >$1
AND q.active_to < $2
AND u.opt_in_ranking = true

GROUP BY user_id, username
ORDER BY total_points DESC;
    `, firstMoment, lastMoment)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var rankings []UserRanking
	for rows.Next() {
		var ranking UserRanking
		if err := rows.Scan(
			&ranking.User_id,
			&ranking.Points,
			&ranking.Username,
			&ranking.Placement); err != nil {
			return nil, err
		}
		rankings = append(rankings, ranking)

	}
	return rankings, nil
}

// Returns the first and last moment of the specified date range.
func GetDateRange(dateRange DateRange, timeZone *time.Location, year int, month time.Month) (time.Time, time.Time) {
	var firstMoment time.Time
	var lastMoment time.Time

	switch dateRange {
	case All:
		firstMoment = time.Date(0, 1, 1, 0, 0, 0, 0, timeZone)
		lastMoment = time.Date(9999, 12, 31, 23, 59, 59, 0, timeZone)
	case Month:
		firstMoment = time.Date(year, month, 1, 0, 0, 0, 0, timeZone)
		lastMoment = firstMoment.AddDate(0, 1, 0).Add(-time.Nanosecond * 1)
	case Year:
		firstMoment = time.Date(year, 1, 1, 0, 0, 0, 0, timeZone)
		lastMoment = time.Date(year, 12, 31, 23, 59, 59, 0, timeZone)
	}
	return firstMoment, lastMoment
}

// Returns the ranking of the specified user.
func GetUserRanking(db *sql.DB, userID uuid.UUID, month time.Month, year int, timeZone *time.Location, dateRange DateRange) (UserRanking, error) {

	firstMoment, lastMoment := GetDateRange(dateRange, timeZone, year, month)

	row := db.QueryRow(`
SELECT user_id, SUM(total_points_awarded) AS total_points, 
CONCAT(u.username_adjective, ' ', u.username_noun) AS username,

 RANK() OVER (ORDER BY SUM(total_points_awarded) DESC) as ranking
FROM "user_quizzes"
JOIN (
SELECT * FROM quizzes
) as q ON q.id = quiz_id

JOIN (
SELECT * FROM users
) as u ON u.id = user_id

WHERE
is_completed = true
AND answered_within_active_time = true

AND q.published = true
AND q.is_deleted = false
AND q.active_from >$1
AND q.active_to < $2
AND u.opt_in_ranking = true AND u.id = $3

GROUP BY user_id, username
ORDER BY total_points DESC;

    `, firstMoment, lastMoment, userID)

	ranking := UserRanking{}
	err := row.Scan(
		&ranking.User_id,
		&ranking.Points,
		&ranking.Username,
		&ranking.Placement)
	if err != nil {
		return UserRanking{}, err
	}
	return ranking, nil
}
