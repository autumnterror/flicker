package jwt

import (
	"flicker/internal/config"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateAndVerifyToken(t *testing.T) {
	t.Parallel()
	j := NewWithConfig(config.Test())

	t.Run("generate and verify access token", func(t *testing.T) {
		t.Parallel()
		tokenStr, err := j.GenerateToken("user123", TokenTypeAccess)
		assert.NoError(t, err)

		token, err := j.VerifyToken(tokenStr)
		assert.NoError(t, err)
		assert.True(t, token.Valid)

		id, err := j.GetIdFromToken(token)
		assert.NoError(t, err)
		assert.Equal(t, "user123", id)

		//role, err := j.GetRoleFromToken(token)
		//assert.NoError(t, err)
		//assert.Equal(t, "editor", role)

		tp, err := j.GetTypeFromToken(token)
		assert.NoError(t, err)
		assert.Equal(t, "ACCESS", tp)
	})

	t.Run("generate and verify refresh token", func(t *testing.T) {
		t.Parallel()
		tokenStr, err := j.GenerateToken("user456", TokenTypeRefresh)
		assert.NoError(t, err)

		token, err := j.VerifyToken(tokenStr)
		assert.NoError(t, err)
		assert.True(t, token.Valid)

		id, err := j.GetIdFromToken(token)
		assert.NoError(t, err)
		assert.Equal(t, "user456", id)

		//role, err := j.GetRoleFromToken(token)
		//assert.NoError(t, err)
		//assert.Equal(t, "reader", role)

		tp, err := j.GetTypeFromToken(token)
		assert.NoError(t, err)
		assert.Equal(t, "REFRESH", tp)
	})
}

func TestRefreshToken(t *testing.T) {
	t.Parallel()
	j := NewWithConfig(config.Test())

	refreshToken, err := j.GenerateToken("refresh_id", TokenTypeRefresh)
	assert.NoError(t, err)

	accessToken, err := j.Refresh(refreshToken)
	assert.NoError(t, err)

	parsed, err := j.VerifyToken(accessToken)
	assert.NoError(t, err)

	tp, err := j.GetTypeFromToken(parsed)
	assert.NoError(t, err)
	assert.Equal(t, TokenTypeAccess, tp)
}

func TestExpiredToken(t *testing.T) {
	t.Parallel()
	cfg := config.Test()
	cfg.AccessTokenLifeTime = -1 * time.Minute
	cfg.RefreshTokenLifeTime = -1 * time.Minute
	j := NewWithConfig(cfg)

	tokenStr, err := j.GenerateToken("expired_user", TokenTypeAccess)
	assert.NoError(t, err)

	_, err = j.VerifyToken(tokenStr)
	assert.ErrorIs(t, err, ErrTokenExpired)

	tokenStr, err = j.GenerateToken("expired_user", TokenTypeRefresh)
	assert.NoError(t, err)

	_, err = j.VerifyToken(tokenStr)
	assert.ErrorIs(t, err, ErrTokenExpired)
}

func TestWrongTypeRefresh(t *testing.T) {
	t.Parallel()
	j := NewWithConfig(config.Test())

	accessToken, err := j.GenerateToken("userX", TokenTypeAccess)
	assert.NoError(t, err)

	_, err = j.Refresh(accessToken)
	assert.ErrorIs(t, err, ErrWrongType)
}

func TestInvalidTokenStructure(t *testing.T) {
	t.Parallel()
	j := NewWithConfig(config.Test())

	_, err := j.VerifyToken("not.a.token")
	assert.Error(t, err)
}
