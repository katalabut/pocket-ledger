package httpauth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Issuer struct {
	secret []byte
	issuer string
}

func NewIssuer(secret, issuer string) *Issuer {
	return &Issuer{secret: []byte(secret), issuer: issuer}
}

func (i *Issuer) Issue(userID, email string, now time.Time) (string, error) {
	claims := jwt.MapClaims{
		"iss": i.issuer,
		"sub": userID,
		"email": email,
		"iat": now.Unix(),
		"exp": now.Add(30 * 24 * time.Hour).Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(i.secret)
}

type Claims struct {
	UserID string
	Email  string
}

func Parse(tokenStr, secret, issuer string) (Claims, error) {
	tok, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	}, jwt.WithIssuer(issuer))
	if err != nil {
		return Claims{}, err
	}
	if !tok.Valid {
		return Claims{}, errors.New("invalid token")
	}
	mc, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return Claims{}, errors.New("invalid claims")
	}
	sub, _ := mc["sub"].(string)
	email, _ := mc["email"].(string)
	if sub == "" || email == "" {
		return Claims{}, errors.New("missing claims")
	}
	return Claims{UserID: sub, Email: email}, nil
}
