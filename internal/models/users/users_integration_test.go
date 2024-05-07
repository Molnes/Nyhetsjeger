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
