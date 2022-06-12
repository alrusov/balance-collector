package http

import (
	"fmt"
	"net/http"

	"github.com/alrusov/initializer"
	"github.com/alrusov/misc"
	"github.com/alrusov/stdhttp"

	"github.com/alrusov/balance-collector/internal/config"
	"github.com/alrusov/balance-collector/internal/show"
)

//----------------------------------------------------------------------------------------------------------------------------//

type (
	// HTTP -- listener struct
	HTTP struct {
		h *stdhttp.HTTP
	}
)

var (
	extraInfo struct {
	}
)

//----------------------------------------------------------------------------------------------------------------------------//

// NewHTTP -- listener initializtion
func NewHTTP(cfg *config.Config) (*stdhttp.HTTP, error) {
	var err error

	h := &HTTP{}

	h.h, err = stdhttp.NewListener(&cfg.Listener.Listener, h)
	if err != nil {
		return nil, err
	}

	h.h.SetExtraInfoFunc(
		func() interface{} {
			return &extraInfo
		},
	)

	h.h.AddEndpointsInfo(
		misc.StringMap{
			"/": "Show data. Parameters: POST[op=( history, name=<name> | update, name=<name> | update-all | repeat )]",
		},
	)

	h.h.RemoveStdPath("/")

	h.h.SetRootItemsFunc(
		func(prefix string) []string {
			return []string{
				fmt.Sprintf(`<a href="%s/"><strong>Return to the main page</strong></a>`, prefix),
			}
		},
	)

	// Инициализируем модули
	err = initializer.Do(cfg, h.h)
	if err != nil {
		return nil, err
	}

	return h.h, nil
}

//----------------------------------------------------------------------------------------------------------------------------//

// Handler -- custom http endpoints handler
func (h *HTTP) Handler(id uint64, prefix string, path string, w http.ResponseWriter, r *http.Request) (processed bool) {
	processed = true

	switch path {
	case "/":
		processed = show.Show(id, prefix, path, w, r)
		return

	default:
		processed = false
		return
	}
}

//----------------------------------------------------------------------------------------------------------------------------//
