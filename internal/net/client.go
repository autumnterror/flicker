package net

import (
	"errors"
	"flicker/internal/auth/jwt"
	"flicker/internal/auth/psql"
	"flicker/internal/config"
	"fmt"

	"net/http"

	"github.com/autumnterror/breezynotes/pkg/utils/format"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

type Echo struct {
	echo    *echo.Echo
	cfg     *config.Config
	authAPI psql.AuthRepo
	jwtAPI  *jwt.WithConfig
}

func New(
	cfg *config.Config,
	authAPI psql.AuthRepo,
	jwtAPI *jwt.WithConfig,
) *Echo {
	e := &Echo{
		echo:    echo.New(),
		cfg:     cfg,
		authAPI: authAPI,
		jwtAPI:  jwtAPI,
	}

	e.echo.GET("/swagger/*", echoSwagger.WrapHandler)
	e.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:5500", "http://127.0.0.1:5500", "http://localhost:8080"},
		AllowMethods:     []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.PATCH, echo.OPTIONS},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true,
	}))
	//e.echo.Use(middleware.Logger(), middleware.Recover())

	api := e.echo.Group("/api")
	{
		api.GET("/health", e.Healthz)
		auth := api.Group("/auth")
		{
			auth.GET("/token", e.ValidateToken)

			auth.POST("", e.Auth)
			auth.POST("/reg", e.Reg)
		}
		ai := api.Group("/ai")
		{
			ai.POST("/generatemd", e.GenerateMarkdown)
		}
	}

	return e
}

func (e *Echo) MustRun() {
	const op = "net.Run"

	if err := e.echo.Start(fmt.Sprintf(":%d", e.cfg.Port)); err != nil && !errors.Is(err, http.ErrServerClosed) {
		e.echo.Logger.Fatal(format.Error(op, err))
	}
}

func (e *Echo) Stop() error {
	const op = "net.Stop"

	if err := e.echo.Close(); err != nil {
		return format.Error(op, err)
	}
	return nil
}
