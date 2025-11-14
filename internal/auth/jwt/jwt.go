package jwt

import (
	"errors"
	"flicker/internal/config"
)

type WithConfig struct {
	cfg *config.Config
}

func NewWithConfig(cfg *config.Config) *WithConfig {
	return &WithConfig{cfg: cfg}
}

var (
	ErrTokenExpired = errors.New("token is expired")
	ErrWrongType    = errors.New("wrong type of token")
)

const (
	TokenTypeAccess  = "ACCESS"
	TokenTypeRefresh = "REFRESH"
)
