package history

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/alrusov/jsonw"

	"github.com/alrusov/balance-collector/internal/config"
	"github.com/alrusov/balance-collector/internal/entity"
	"github.com/alrusov/balance-collector/internal/htmlpage"
	"github.com/alrusov/balance-collector/internal/operator"
)

//----------------------------------------------------------------------------------------------------------------------------//

type (
	outData struct {
		ID      uint       `json:"id"`
		Name    string     `json:"name"`
		Login   string     `json:"login"`
		List    []*dataRow `json:"list"`
		FLegend []string   `json:"fLegend"`
		SLegend []string   `json:"sLegend"`
		Ftail   []int      `json:"-"`
		Stail   []int      `json:"-"`
	}

	dataRow struct {
		TS   time.Time     `json:"ts"`
		Info operator.Data `json:"data"`
	}
)

//----------------------------------------------------------------------------------------------------------------------------//

// Do --
func Do(id uint64, prefix string, w http.ResponseWriter, r *http.Request, entityID uint) (err error) {
	cfg := config.Get()

	var data *outData

	func() {
		// Готовим данные для вывода
		data, err = load(cfg, entityID)
		if err != nil {
			return
		}
	}()

	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	title := "История запроcов"

	// Выводим
	switch r.Form.Get("raw") {
	case "json":
		j, _ := jsonw.Marshal(data)
		err = htmlpage.Raw(cfg, prefix, w, r, errMsg, title, string(j))
		return

	default:
		err = showPage(cfg, prefix, w, r, errMsg, title, data)
		return
	}
}

//----------------------------------------------------------------------------------------------------------------------------//

// load -- Загружаем последние
func load(cfg *config.Config, entityID uint) (data *outData, err error) {
	data = &outData{}

	e := entity.GetByID(entityID)
	eCfg := e.Config()
	if e == nil || !eCfg.Enabled {
		// если неизвесное или не разрешено - удаляем
		err = fmt.Errorf(`unknown or disabled entity %d`, entityID)
		return
	}

	data.ID = eCfg.ID
	data.Name = eCfg.Name
	data.Login = eCfg.Login
	data.FLegend, data.SLegend = e.Legend()

	query := `
	select h.ts, h.data
		from history as h
		where h.entity_id=?
		order by h.ts desc;
	`

	db, err := sql.Open("sqlite3", cfg.Processor.DB)
	if err != nil {
		return
	}
	defer db.Close()

	stmt, err := db.Prepare(query)
	if err != nil {
		return
	}
	defer stmt.Close()

	rows, err := stmt.Query(entityID)
	if err != nil {
		return
	}

	nF := len(data.FLegend) // количество активных float колонок
	nS := len(data.SLegend) // количество активных string колонок

	// Загружаем из базы

	for rows.Next() {
		var d dataRow
		var ts int64
		var s string

		err = rows.Scan(&ts, &s)
		if err != nil {
			return
		}

		d.TS = time.Unix(ts, 0) // локальное

		err = jsonw.Unmarshal([]byte(s), &d.Info)
		if err != nil {
			return
		}

		data.List = append(data.List, &d)

		ln := len(d.Info.FVals)
		if ln > nF {
			// акивных float колонок стало больше
			nF = ln
		}

		ln = len(d.Info.SVals)
		if ln > nS {
			// активных string колонок стало больше
			nS = ln
		}
	}

	data.Ftail = make([]int, len(data.List))
	data.Stail = make([]int, len(data.List))

	for i, d := range data.List {
		// сколько колонок надо будет дополнять при выводе до максимално отображаемых
		data.Ftail[i] = nF - len(d.Info.FVals)
		data.Stail[i] = nS - len(d.Info.SVals)

		// если легенды короче, чем данные, дополняем пробелами

		for {
			if len(data.FLegend) >= nF {
				break
			}
			data.FLegend = append(data.FLegend, "")
		}

		for {
			if len(data.SLegend) >= nS {
				break
			}
			data.SLegend = append(data.SLegend, "")
		}
	}

	return
}

//----------------------------------------------------------------------------------------------------------------------------//

func showPage(cfg *config.Config, prefix string, w http.ResponseWriter, r *http.Request, errMsg string, title string, data *outData) (err error) {
	params := struct {
		Data      *outData
		Fcount    int
		Scount    int
		ColsCount int
	}{
		Data:      data,
		Fcount:    1,
		Scount:    1,
		ColsCount: 3,
	}

	if len(data.List) > 0 {
		// если есть хотя бы одна строка, вычисляем общее количество колонок по типам
		params.Fcount = data.Ftail[0] + len(data.List[0].Info.FVals)
		params.Scount = data.Stail[0] + len(data.List[0].Info.SVals)
		params.ColsCount = 1 + params.Fcount + params.Scount
	}

	return htmlpage.Do("history", prefix, w, r, errMsg, title, params)
}

//----------------------------------------------------------------------------------------------------------------------------//
