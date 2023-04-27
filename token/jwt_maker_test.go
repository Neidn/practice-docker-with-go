package token

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
	"practice-docker/util"
	"testing"
	"time"
)

func TestJWTMaker_CreateToken(t *testing.T) {
	maker, err := NewJWTMaker(util.RandomString(32))
	require.NoErrorf(t, err, "cannot create jwt maker")

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

func TestJWTMaker_VerifyToken_ExpiredJWTToken(t *testing.T) {
	maker, err := NewJWTMaker(util.RandomString(32))
	require.NoErrorf(t, err, "cannot create jwt maker")

	// create token with -1 minute duration
	token, err := maker.CreateToken(util.RandomOwner(), -time.Minute)
	require.NoErrorf(t, err, "cannot create token")
	require.NotEmptyf(t, token, "token should not be empty")

	payload, err := maker.VerifyToken(token)
	require.Errorf(t, err, "")
	require.EqualErrorf(t, err, fmt.Sprintf("%s: %s", ErrTokenInvalidClaims, ErrExpiredToken), "token should be expired")
	require.Nilf(t, payload, "payload should be nil")

}

func TestJWTMaker_VerifyToken_InvalidJWTTokenAlgNone(t *testing.T) {
	payload, err := NewPayload(util.RandomOwner(), time.Minute)
	require.NoErrorf(t, err, "cannot create payload")

	// create token with none algorithm
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodNone, payload)
	token, err := jwtToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoErrorf(t, err, "cannot sign token")

	maker, err := NewJWTMaker(util.RandomString(32))
	require.NoErrorf(t, err, "cannot create jwt maker")

	payload, err = maker.VerifyToken(token)
	require.Errorf(t, err, "")
	require.EqualErrorf(t, err, fmt.Sprintf("%s: error while executing keyfunc: %s", ErrTokenUnverifiable, ErrInvalidToken), "token should be invalid")
	require.Nilf(t, payload, "payload should be nil")
}
