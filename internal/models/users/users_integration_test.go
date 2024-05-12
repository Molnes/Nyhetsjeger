//go:build integration

package users

import (
	"context"
	"testing"
	"time"

	db_integration_test_suite "github.com/Molnes/Nyhetsjeger/db"
	"github.com/stretchr/testify/suite"
)

type UsersIntegrationTestSuite struct {
	db_integration_test_suite.DbIntegrationTestBaseSuite
}

func TestUsersIntegrationSuite(t *testing.T) {
	suite.Run(t, new(UsersIntegrationTestSuite))
}

func (s *UsersIntegrationTestSuite) TestCreateUser() {
	const (
		ssoId = "1000000test"
		email = "integration_test_user@email.com"
	)

	user, err := CreateUser(s.DB, context.Background(), &PartialUser{
		SsoID:        ssoId,
		Email:        email,
		AccessToken:  "nothing",
		RefreshToken: "nothing2",
		TokenExpire:  time.Now().Add(time.Hour),
	})
	s.Require().NoError(err)

	s.Require().Equal(ssoId, user.SsoID)
	s.Require().Equal(email, user.Email)
}

func (s *UsersIntegrationTestSuite) TestGetUserById() {
	user, err := GetUserByID(s.DB, s.InsertedValues.UserId)
	s.Require().NoError(err)

	s.Require().Equal(s.InsertedValues.UserId, user.ID)
	s.Require().Equal(s.InsertedValues.UserSsoId, user.SsoID)
	s.Require().Equal(s.InsertedValues.UserEmail, user.Email)
}

func (s *UsersIntegrationTestSuite) TestGetUserBySsoId() {
	userById, err := GetUserByID(s.DB, s.InsertedValues.UserId)
	s.Require().NoError(err)

	user, err := GetUserBySsoID(s.DB, s.InsertedValues.UserSsoId)
	s.Require().NoError(err)

	s.Require().Equal(userById, user)

}

func (s *UsersIntegrationTestSuite) TestGetUserByEmail() {
	userById, err := GetUserByID(s.DB, s.InsertedValues.UserId)
	s.Require().NoError(err)

	user, err := GetUserByEmail(s.DB, s.InsertedValues.UserEmail)
	s.Require().NoError(err)

	s.Require().Equal(userById, user)
}
