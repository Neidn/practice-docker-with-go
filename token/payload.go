package token

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"time"
)

var (
	ErrInvalidToken = errors.New("token is invalid")
	ErrExpiredToken = errors.New("token is expired")
)

// Payload is the payload of a token.
type Payload struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Issuer    string    `json:"issuer"`
	Subject   string    `json:"subject"`
	Audience  []string  `json:"audience"`
	NotBefore time.Time `json:"not_before"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
}

func (p Payload) GetExpirationTime() (*jwt.NumericDate, error) {
	if time.Now().After(p.ExpiredAt) {
		return nil, ErrExpiredToken
	}
	return jwt.NewNumericDate(p.ExpiredAt), nil
}

func (p Payload) GetIssuedAt() (*jwt.NumericDate, error) {
	if time.Now().Before(p.IssuedAt) {
		return nil, ErrInvalidToken
	}
	return jwt.NewNumericDate(p.IssuedAt), nil
}

func (p Payload) GetNotBefore() (*jwt.NumericDate, error) {
	if time.Now().Before(p.NotBefore) {
		return nil, ErrInvalidToken
	}
	return jwt.NewNumericDate(p.NotBefore), nil
}

func (p Payload) GetIssuer() (string, error) {
	return p.Issuer, nil
}

func (p Payload) GetSubject() (string, error) {
	return p.Subject, nil
}

func (p Payload) GetAudience() (jwt.ClaimStrings, error) {
	return p.Audience, nil
}

// NewPayload creates a new payload for a specific username and duration.
func NewPayload(username string, duration time.Duration) (payload *Payload, err error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return
	}

	payload = &Payload{
		ID:        tokenID,
		Username:  username,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}

	return
}
