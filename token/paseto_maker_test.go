package token

import (
	"github.com/stretchr/testify/require"
	"practice-docker/util"
	"testing"
	"time"
)

func TestPasetoMaker_CreateToken(t *testing.T) {
	maker, err := NewPasetoMaker(util.RandomString(32))
	require.NoErrorf(t, err, "cannot create paseto maker")

	username := util.RandomOwner()
	duration := time.Minute

	issueAt := time.Now()
	expireAt := issueAt.Add(duration)

	token, err := maker.CreateToken(username, duration)
	require.NoErrorf(t, err, "cannot create token")
	require.NotEmptyf(t, token, "token should not be empty")

	payload, err := maker.VerifyToken(token)
	require.NoErrorf(t, err, "cannot verify token")
	require.NotEmptyf(t, payload, "payload should not be empty")

	require.NotZerof(t, payload.ID, "id should not be zero")
	require.Equalf(t, username, payload.Username, "username should be the same")
	require.WithinDurationf(t, issueAt, payload.IssuedAt, time.Second, "issuedAt should be the same")
	require.WithinDurationf(t, expireAt, payload.ExpiredAt, time.Second, "expiredAt should be the same")
}
