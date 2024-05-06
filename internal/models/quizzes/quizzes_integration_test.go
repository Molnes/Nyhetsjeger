//go:build integration

package quizzes

import (
	"testing"

	db_integration_test_suite "github.com/Molnes/Nyhetsjeger/db"
	"github.com/stretchr/testify/suite"
)

type UsersIntegrationTestSuit struct {
	db_integration_test_suite.DbIntegrationTestBaseSuite
}

func TestQuizzesIntegrationSuite(t *testing.T) {
	suite.Run(t, new(UsersIntegrationTestSuit))
}

func (s *UsersIntegrationTestSuit) TestCreateQuiz() {
	quiz := CreateDefaultQuiz()

	id, err := CreateQuiz(s.DB, quiz)
	s.Require().NoError(err)

	s.Require().Equal(quiz.ID, *id)
}

func (s *UsersIntegrationTestSuit) TestGetQuizByIdQuiz() {
	quiz := CreateDefaultQuiz()

	_, err := CreateQuiz(s.DB, quiz)
	s.Require().NoError(err)

	createdQuiz, err := GetQuizByID(s.DB, quiz.ID)
	s.Require().NoError(err)

	s.Require().Equal(quiz.ID, createdQuiz.ID)
	s.Require().Equal(quiz.ImageURL, createdQuiz.ImageURL)
	s.Require().Equal(quiz.IsDeleted, createdQuiz.IsDeleted)
	s.Require().Equal(quiz.Published, createdQuiz.Published)
	s.Require().Equal(quiz.Title, createdQuiz.Title)
}
