package user_ranking

import (
	"database/sql"

	"github.com/Molnes/Nyhetsjeger/internal/models/labels"
	"github.com/google/uuid"
)

// UserRanking represents a user's ranking in the scoreboard.
type UserRanking struct {
	UserID    uuid.UUID
	Username  string
	Points    int
	Placement int
}

// UserRankingWithLabel represents a user's ranking in the scoreboard with a label.
type UserRankingWithLabel struct {
	UserID    uuid.UUID
	Username  string
	Points    int
	Placement int
	Label     labels.Label
}

// RankingByLabel represents a ranking of users by label.
type RankingByLabel struct {
	Label   labels.Label
	Ranking []UserRanking
}

type DateRange int

const (
	All   DateRange = 0
	Month DateRange = 1
	Year  DateRange = 2
)

// Returns an empty UserRankingWithLabel struct.
func EmptyRanking() UserRankingWithLabel {
	return UserRankingWithLabel{
		UserID:    uuid.Nil,
		Username:  "",
		Points:    0,
		Placement: 0,
		Label:     labels.Label{},
	}
}

// Returns the ranking of all users who have opted in to the ranking.
func GetRanking(db *sql.DB, labelID uuid.UUID) ([]UserRanking, error) {

	rows, err := db.Query(`
        SELECT user_id, SUM(total_points_awarded) AS total_points, 
CONCAT(u.username_adjective, ' ', u.username_noun) AS username,

 RANK() OVER (ORDER BY SUM(total_points_awarded) DESC) as ranking
FROM "user_quizzes"
JOIN (
SELECT * FROM quizzes
JOIN (
SELECT * FROM quiz_labels
WHERE label_id = $1
) as ql ON ql.quiz_id = quizzes.id
) as q ON q.id = user_quizzes.quiz_id

JOIN (
SELECT * FROM users
) as u ON u.id = user_id

WHERE
is_completed = true
AND answered_within_active_time = true

AND q.published = true
AND q.is_deleted = false
AND u.opt_in_ranking = true

GROUP BY user_id, username
ORDER BY total_points DESC;
    `, labelID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var rankings []UserRanking
	for rows.Next() {
		var ranking UserRanking
		if err := rows.Scan(
			&ranking.UserID,
			&ranking.Points,
			&ranking.Username,
			&ranking.Placement); err != nil {
			return nil, err
		}
		rankings = append(rankings, ranking)

	}
	return rankings, nil
}

// Returns the ranking of the specified user.
func GetUserRanking(db *sql.DB, userID uuid.UUID, label labels.Label) (UserRanking, error) {

	row := db.QueryRow(`
    SELECT * FROM (
SELECT user_id, SUM(total_points_awarded) AS total_points, 
CONCAT(u.username_adjective, ' ', u.username_noun) AS username,

RANK() OVER (ORDER BY SUM(total_points_awarded) DESC) as ranking
FROM "user_quizzes"
JOIN (
SELECT * FROM quizzes
JOIN (
SELECT * FROM quiz_labels
WHERE label_id = $1
) as ql ON ql.quiz_id = quizzes.id
) as q ON q.id = user_quizzes.quiz_id


JOIN (
SELECT * FROM users
) as u ON u.id = user_id

WHERE
is_completed = true
AND answered_within_active_time = true

AND q.published = true
AND q.is_deleted = false
AND u.opt_in_ranking = true 

GROUP BY user_id, username
ORDER BY total_points DESC) AS ranking
WHERE user_id = $2;

    `, label.ID, userID)

	ranking := UserRanking{}
	err := row.Scan(
		&ranking.UserID,
		&ranking.Points,
		&ranking.Username,
		&ranking.Placement)
	if err != nil {
		return UserRanking{}, err
	}
	return ranking, nil
}

// Returns the ranking of the specified user regardless of label.
func GetUserRankingAllLabels(db *sql.DB, userID uuid.UUID) (UserRanking, error) {

	row := db.QueryRow(`
    SELECT * FROM (
SELECT user_id, SUM(total_points_awarded) AS total_points, 
CONCAT(u.username_adjective, ' ', u.username_noun) AS username,

RANK() OVER (ORDER BY SUM(total_points_awarded) DESC) as ranking
FROM "user_quizzes"
JOIN (
SELECT * FROM quizzes
) as q ON q.id = user_quizzes.quiz_id


JOIN (
SELECT * FROM users
) as u ON u.id = user_id

WHERE
is_completed = true
AND answered_within_active_time = true

AND q.published = true
AND q.is_deleted = false
AND u.opt_in_ranking = true 

GROUP BY user_id, username
ORDER BY total_points DESC) AS ranking
WHERE user_id = $1;

    `, userID)

	ranking := UserRanking{}
	err := row.Scan(
		&ranking.UserID,
		&ranking.Points,
		&ranking.Username,
		&ranking.Placement)
	if err != nil {
		return UserRanking{}, err
	}
	return ranking, nil
}

// Data transfer object wrapping 3 user rankings in the three DateRanges
type RankingCollection struct {
	ByLabel UserRankingWithLabel
	AllTime UserRanking
}

// Gets a colleciton of user rankings, Monthly, Yearly and AllTime for the given period.
func GetUserRankingsInAllRanges(db *sql.DB, userId uuid.UUID, label labels.Label) (*RankingCollection, error) {
	labelRank, err := GetUserRanking(db, userId, label)
	if err != nil {
		if err == sql.ErrNoRows {
			labelRank = createEmptyRanking(userId, "")
		} else {
			return nil, err
		}
	}

	allTimeRank, err := GetUserRankingAllLabels(db, userId)
	if err != nil {
		if err == sql.ErrNoRows {
			allTimeRank = createEmptyRanking(userId, "")
		} else {
			return nil, err
		}
	}

	return &RankingCollection{
		ByLabel: UserRankingWithLabel{
			UserID:    labelRank.UserID,
			Username:  labelRank.Username,
			Points:    labelRank.Points,
			Placement: labelRank.Placement,
			Label:     label,
		},
		AllTime: allTimeRank,
	}, nil
}

// Creates a UserRanking struct with correct user data but points and placement set to 0.
func createEmptyRanking(userID uuid.UUID, username string) UserRanking {
	return UserRanking{
		userID,
		username,
		0,
		0,
	}
}
