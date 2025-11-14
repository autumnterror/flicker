package jwt

import (
	"errors"
	"fmt"
	"github.com/autumnterror/breezynotes/pkg/utils/format"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

// GenerateToken generation JWT by TYPE values: "ACCESS" or "REFRESH"
func (w *WithConfig) GenerateToken(id, _type string) (string, error) {
	const op = "jwt.WithConfig.GenerateToken"
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = id
	claims["type"] = _type
	//claims["role"] = role
	switch _type {
	case TokenTypeAccess:
		claims["exp"] = time.Now().Add(w.cfg.AccessTokenLifeTime).Unix()
	case TokenTypeRefresh:
		claims["exp"] = time.Now().Add(w.cfg.RefreshTokenLifeTime).Unix()
	default:
		return "", format.Error(op, ErrWrongType)
	}

	ts, err := token.SignedString([]byte(w.cfg.TokenKey))
	if err != nil {
		return "", format.Error(op, err)
	}
	return ts, nil
}

// VerifyToken return the raw token
func (w *WithConfig) VerifyToken(tokenString string) (*jwt.Token, error) {
	const op = "jwt.WithConfig.VerifyToken"
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, format.Error(op, fmt.Errorf("unexpected signing method: %v", token.Header["alg"]))
		}
		return []byte(w.cfg.TokenKey), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, format.Error(op, err)
	}
	return token, nil
}

// GetIdFromToken return id from token
func (w *WithConfig) GetIdFromToken(token *jwt.Token) (string, error) {
	const op = "jwt.WithConfig.GetLoginFromToken"

	c, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", format.Error(op, fmt.Errorf("invalid token claims"))
	}

	id := c["id"]
	if id == "" {
		return "", format.Error(op, fmt.Errorf("id not detected"))
	}
	return id.(string), nil
}

//// GetRoleFromToken return role from token
//func (w *WithConfig) GetRoleFromToken(token *jwt.Token) (string, error) {
//	const op = "jwt.WithConfig.GetLoginFromToken"
//
//	c, ok := token.Claims.(jwt.MapClaims)
//	if !ok || !token.Valid {
//		return "", format.Error(op, fmt.Errorf("invalid token claims"))
//	}
//
//	id := c["role"]
//	if id == "" {
//		return "", format.Error(op, fmt.Errorf("role not detected"))
//	}
//	return id.(string), nil
//}

// GetTypeFromToken return type from token
func (w *WithConfig) GetTypeFromToken(token *jwt.Token) (string, error) {
	const op = "jwt.WithConfig.GetLoginFromToken"

	c, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", format.Error(op, fmt.Errorf("invalid token claims"))
	}

	tp := c["type"]
	if tp == "" {
		return "", format.Error(op, fmt.Errorf("type not detected"))
	}
	return tp.(string), nil
}

// Refresh check refresh token and if all ok return new access token
func (w *WithConfig) Refresh(refreshToken string) (string, error) {
	const op = "jwt.WithConfig.Refresh"

	rawRefToken, err := w.VerifyToken(refreshToken)
	if err != nil {
		if errors.Is(err, ErrTokenExpired) {
			return "", err
		}
		return "", format.Error(op, err)
	}
	tp, err := w.GetTypeFromToken(rawRefToken)
	if err != nil {
		return "", format.Error(op, err)
	}

	if tp != TokenTypeRefresh {
		return "", ErrWrongType
	}

	id, err := w.GetIdFromToken(rawRefToken)
	if err != nil {
		return "", format.Error(op, err)
	}
	//role, err := w.GetRoleFromToken(rawRefToken)
	//if err != nil {
	//	return "", format.Error(op, err)
	//}

	token, err := w.GenerateToken(id, TokenTypeAccess)
	if err != nil {
		return "", format.Error(op, err)
	}
	return token, nil
}

type WithConfigRepo interface {
	GenerateToken(id, _type string) (string, error)
	VerifyToken(tokenString string) (*jwt.Token, error)
	GetIdFromToken(token *jwt.Token) (string, error)
	GetTypeFromToken(token *jwt.Token) (string, error)
	Refresh(refreshToken string) (string, error)
}
