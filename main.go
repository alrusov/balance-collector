package main

import (
	"github.com/alrusov/launcher"
	"github.com/alrusov/panic"
	"github.com/alrusov/stdhttp"

	"github.com/alrusov/balance-collector/internal/config"
	"github.com/alrusov/balance-collector/internal/http"

	// sqlite driver
	_ "github.com/mattn/go-sqlite3"
)

//----------------------------------------------------------------------------------------------------------------------------//

type app struct {
	cfg *config.Config
}

//----------------------------------------------------------------------------------------------------------------------------//

func (a *app) CheckConfig() error {
	return a.cfg.Check()
}

func (a *app) NewListener() (*stdhttp.HTTP, error) {
	h, err := http.NewHTTP(a.cfg)
	if err != nil {
		return nil, err
	}

	return h, nil
}

//----------------------------------------------------------------------------------------------------------------------------//

func main() {
	panicID := panic.ID()
	defer panic.SaveStackToLogEx(panicID)

	app := &app{
		cfg: &config.Config{},
	}

	launcher.Go(app, app.cfg)
}

//----------------------------------------------------------------------------------------------------------------------------//
