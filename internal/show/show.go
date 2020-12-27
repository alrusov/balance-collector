package show

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/alrusov/stdhttp"

	"github.com/alrusov/balance-collector/internal/entity"
	"github.com/alrusov/balance-collector/internal/show/dashboard"
	"github.com/alrusov/balance-collector/internal/show/history"
)

//----------------------------------------------------------------------------------------------------------------------------//

// Show --
func Show(id uint64, prefix string, path string, w http.ResponseWriter, r *http.Request) (processed bool) {
	processed = true

	var err error
	defer func() {
		if err != nil {
			stdhttp.Error(id, false, w, http.StatusInternalServerError, "Server error", err)
		}
	}()

	err = r.ParseForm()
	if err != nil {
		return
	}

	returnToMain := true

	switch r.Form.Get("op") {
	case "update-all":
		entity.Update(fmt.Sprintf("%d", id), true)

	case "repeat":
		entity.Repeat(fmt.Sprintf("%d", id))

	case "history":
		entityID, e := strconv.ParseUint(r.Form.Get("id"), 10, 32)
		if e != nil {
			err = e
		} else {
			err = history.Do(id, w, r, uint(entityID))
			returnToMain = false
		}

	case "update":
		entityID, e := strconv.ParseUint(r.Form.Get("id"), 10, 32)
		if e != nil {
			err = e
		} else {
			entity.UpdateByID(fmt.Sprintf("%d", id), uint(entityID))
		}

	case "":
		fallthrough
	default:
		err = dashboard.Do(id, w, r)
		returnToMain = false
	}

	if returnToMain {
		stdhttp.ReturnRefresh(id, w, r, http.StatusOK, path, nil, nil)

	}

	return
}

//----------------------------------------------------------------------------------------------------------------------------//
