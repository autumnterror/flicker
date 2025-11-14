package config

import "time"

func Test() *Config {
	return &Config{
		Uri:                  "postgres://postgres:postgrespw@localhost:5432/userdb?sslmode=disable",
		TokenKey:             "SANYAmt+i061Wfm6#(=m1?4G&nLKqq*P9HlQTyS@k#D%YOKSVSl@jEQVcb2cwRbPROVOROV",
		AccessTokenLifeTime:  5 * time.Second,
		RefreshTokenLifeTime: time.Minute,
		Port:                 8008,
	}
}
