package api

import (
	"github.com/stretchr/testify/require"
	db "practice-docker/db/sqlc"
	"practice-docker/util"
	"testing"
)

func randomUser(t *testing.T) db.Users {
	password := util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	require.NoErrorf(t, err, "cannot hash password: %v", err)

	return db.Users{
		Username:       util.RandomOwner(),
		HashedPassword: hashedPassword,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}
}
