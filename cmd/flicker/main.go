package main

import (
	_ "flicker/docs"
	"flicker/internal/auth/jwt"
	"flicker/internal/auth/psql"
	"flicker/internal/config"
	"flicker/internal/net"
	"os"
	"os/signal"
	"syscall"

	"github.com/autumnterror/breezynotes/pkg/log"
)

// @title flicker rest api
// @version 0.1-.-infDev
// @description yoyo
// @termsOfService https://about.breezynotes.ru/

// @contact.name Alex "bustard" Provor
// @contact.url https://contacts.breezynotes.ru
// @contact.email help@breezynotes.ru

// @host localhost:8080
// @BasePath /
// @schemes http
func main() {
	const op = "cmd.auth"
	cfg := config.MustSetup()

	db := psql.MustConnect(cfg)

	e := net.New(cfg, psql.NewDriver(db.Driver), jwt.NewWithConfig(cfg))
	go e.MustRun()

	sign := wait()

	if err := e.Stop(); err != nil {
		log.Error(op, "stop echo", err)
	}

	if err := db.Disconnect(); err != nil {
		log.Error(op, "db disconnect", err)
	}

	log.Success(op, "stop signal "+sign)
}

func wait() string {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	sign := <-stop
	return sign.String()
}
