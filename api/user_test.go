package api

import (
	db "practice-docker/db/sqlc"
	"practice-docker/util"
)

func randomUser() db.Users {
	return db.Users{
		Username:       util.RandomOwner(),
		HashedPassword: util.RandomHashedPassword(),
	}
}
