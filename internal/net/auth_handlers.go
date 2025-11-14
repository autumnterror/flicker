package net

import (
	"context"
	"errors"
	"flicker/internal/auth/jwt"
	"flicker/internal/auth/psql"
	"flicker/internal/views"
	"net/http"
	"time"

	"github.com/autumnterror/breezynotes/pkg/log"
	"github.com/autumnterror/breezynotes/pkg/utils/id"
	"github.com/autumnterror/breezynotes/pkg/utils/validate"
	"github.com/labstack/echo/v4"
)

// Auth godoc
// @Summary Authorize user
// @Description Authenticates user and returns access/refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param User body views.AuthRequest true "Login or Email and Password"
// @Success 200 {object} views.Tokens
// @Failure 400 {object} views.SWGError
// @Failure 401 {object} views.SWGError
// @Failure 502 {object} views.SWGError
// @Router /api/auth [post]
func (e *Echo) Auth(c echo.Context) error {
	const op = "net.Auth"
	log.Info(op, "")

	var r views.AuthRequest
	if err := c.Bind(&r); err != nil {
		log.Error(op, "auth bind", err)
		return c.JSON(http.StatusBadRequest, views.SWGError{Error: "bad JSON"})
	}

	ctx, done := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer done()

	id, err := e.authAPI.Authentication(ctx, r.Email, r.Login, r.Password)
	if err != nil {
		switch {
		case errors.Is(err, psql.ErrNoUser):
			log.Warn(op, "", err)
			return c.JSON(http.StatusUnauthorized, views.SWGError{Error: "wrong login or password"})
		case errors.Is(err, psql.ErrPasswordIncorrect):
			log.Warn(op, "", err)
			return c.JSON(http.StatusUnauthorized, views.SWGError{Error: "wrong login or password"})
		case errors.Is(err, psql.ErrWrongInput):
			log.Warn(op, "", err)
			return c.JSON(http.StatusBadRequest, views.SWGError{Error: "bad argument"})
		default:
			log.Error(op, "", err)
			return c.JSON(http.StatusInternalServerError, views.SWGError{Error: "authentication error"})
		}
	}

	at, err := e.jwtAPI.GenerateToken(id, jwt.TokenTypeAccess)
	if err != nil {
		log.Error(op, "token generation error", err)
		return c.JSON(http.StatusBadGateway, views.SWGError{Error: "token generation error"})
	}

	rt, err := e.jwtAPI.GenerateToken(id, jwt.TokenTypeRefresh)
	if err != nil {
		log.Error(op, "token generation error", err)
		return c.JSON(http.StatusBadGateway, views.SWGError{Error: "token generation error"})
	}

	c.SetCookie(&http.Cookie{
		Name:     "access_token",
		Value:    at,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(e.cfg.AccessTokenLifeTime),
	})
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    rt,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(e.cfg.RefreshTokenLifeTime),
	})

	log.Success(op, "")

	return c.JSON(http.StatusOK, views.Tokens{
		AccessToken:  at,
		RefreshToken: rt,
	})
}

// Reg godoc
// @Summary Register new user
// @Description Validates registration data, creates user and returns tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param User body views.UserRegister true "Reg data"
// @Success 200 {object} views.SWGError
// @Failure 400 {object} views.SWGError
// @Failure 302 {object} views.SWGError
// @Failure 502 {object} views.SWGError
// @Router /api/auth/reg [post]
func (e *Echo) Reg(c echo.Context) error {
	const op = "net.Reg"
	log.Info(op, "")

	var u views.UserRegister
	if err := c.Bind(&u); err != nil {
		log.Warn(op, "bad JSON", err)
		return c.JSON(http.StatusBadRequest, views.SWGError{Error: "bad JSON"})
	}
	if u.Pw1 != u.Pw2 {
		log.Warn(op, "password not same", nil)
		return c.JSON(http.StatusBadRequest, views.SWGError{Error: "password not same"})
	}

	if !validate.Password(u.Pw1) {
		log.Warn(op, "password not in policy", nil)
		return c.JSON(http.StatusBadRequest, views.SWGError{Error: "password not in policy"})
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 3*time.Second)
	defer cancel()

	id := id.New()
	err := e.authAPI.Create(ctx, &views.User{
		Id:       id,
		Login:    u.Login,
		Email:    u.Email,
		About:    "Write me!",
		Photo:    "images/default.png",
		Password: u.Pw1,
	})
	if err != nil {
		switch {
		case errors.Is(err, psql.ErrAlreadyExist):
			log.Warn(op, "", err)
			return c.JSON(http.StatusFound, views.SWGError{Error: "user already exist"})
		default:
			log.Error(op, "", err)
			return c.JSON(http.StatusBadGateway, views.SWGError{Error: "user creation failed"})
		}
	}

	at, err := e.jwtAPI.GenerateToken(id, jwt.TokenTypeAccess)
	if err != nil {
		log.Error(op, "token generation error", err)
		return c.JSON(http.StatusBadGateway, views.SWGError{Error: "token generation error"})
	}

	rt, err := e.jwtAPI.GenerateToken(id, jwt.TokenTypeRefresh)
	if err != nil {
		log.Error(op, "token generation error", err)
		return c.JSON(http.StatusBadGateway, views.SWGError{Error: "token generation error"})
	}

	c.SetCookie(&http.Cookie{
		Name:     "access_token",
		Value:    at,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(e.cfg.AccessTokenLifeTime),
	})
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    rt,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(e.cfg.RefreshTokenLifeTime),
	})

	log.Success(op, "")

	return c.JSON(http.StatusOK, views.Tokens{
		AccessToken:  at,
		RefreshToken: rt,
	})
}

// ValidateToken godoc
// @Summary Validate token (uses cookies)
// @Description Checks access token from cookie, tries to refresh if expired. If 410 (GONE) need to re-auth
// @Tags auth
// @Produce json
// @Success 200 {object} views.SWGMessage
// @Success 201 {object} string
// @Failure 400 {object} views.SWGError
// @Failure 410 {object} views.SWGError
// @Failure 502 {object} views.SWGError
// @Router /api/auth/token [get]
func (e *Echo) ValidateToken(c echo.Context) error {
	const op = "gateway.net.TokenValidate"
	log.Info(op, "")

	at, err := c.Cookie("access_token")
	if err != nil {
		log.Warn(op, "access_token cookie missing", err)
		return c.JSON(http.StatusBadRequest, views.SWGError{Error: "access_token cookie missing"})
	}
	rt, err := c.Cookie("refresh_token")
	if err != nil {
		log.Warn(op, "refresh_token cookie missing", err)
		return c.JSON(http.StatusBadRequest, views.SWGError{Error: "refresh_token cookie missing"})
	}

	if _, err := e.jwtAPI.VerifyToken(at.Value); err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			log.Warn(op, "", err)
			newAt, err := e.jwtAPI.Refresh(rt.Value)
			if err != nil {
				switch {
				case errors.Is(err, jwt.ErrTokenExpired):
					log.Warn(op, "", err)
					return c.JSON(http.StatusGone, views.SWGError{Error: "refresh token expired. Terminate authorization"})
				case errors.Is(err, jwt.ErrWrongType):
					log.Warn(op, "", err)
					return c.JSON(http.StatusBadRequest, views.SWGError{Error: "invalid refresh token"})
				default:
					log.Error(op, "", err)
					return c.JSON(http.StatusInternalServerError, views.SWGError{Error: "refresh failed"})
				}
			}
			c.SetCookie(&http.Cookie{
				Name:     "access_token",
				Value:    newAt,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				Expires:  time.Now().Add(e.cfg.AccessTokenLifeTime),
			})
			return c.JSON(http.StatusCreated, newAt)
		default:
			log.Warn(op, "", err)
			return c.JSON(http.StatusBadRequest, views.SWGError{Error: "invalid access token"})
		}
	}
	log.Success(op, "")
	return c.JSON(http.StatusOK, views.SWGMessage{Message: "tokens valid"})
}
