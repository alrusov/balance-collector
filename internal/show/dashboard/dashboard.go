package dashboard

import (
	"database/sql"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/alrusov/jsonw"
	"github.com/alrusov/misc"

	"github.com/alrusov/balance-collector/internal/config"
	"github.com/alrusov/balance-collector/internal/entity"
	"github.com/alrusov/balance-collector/internal/htmlpage"
	"github.com/alrusov/balance-collector/internal/operator"
)

//----------------------------------------------------------------------------------------------------------------------------//

type (
	dataRow struct {
		Idx        int            `json:"-"`
		ID         uint           `json:"id"`
		Name       string         `json:"name"`
		Login      string         `json:"login"`
		TS         time.Time      `json:"ts"`
		Info       operator.Data  `json:"info"`
		LastChange operator.FVals `json:"lastChange"`
		Error      string         `json:"error"`
		Ferror     []string       `json:"fErrors"`
		FLegend    []string       `json:"fLegend"`
		SLegend    []string       `json:"sLegend"`
		Ftail      int            `json:"-"`
		Stail      int            `json:"-"`
	}

	dataArr []*dataRow
	dataMap map[uint]*dataRow
)

//----------------------------------------------------------------------------------------------------------------------------//

// Do --
func Do(id uint64, prefix string, w http.ResponseWriter, r *http.Request) (err error) {
	cfg := config.Get()

	var mList dataMap
	var aList dataArr

	func() {
		var db *sql.DB

		db, err = sql.Open("sqlite3", cfg.Processor.DB)
		if err != nil {
			return
		}
		defer db.Close()

		// Готовим данные для вывода

		mList, aList, err = loadLast(cfg, db)
		if err != nil {
			return
		}

		err = loadPrev(cfg, db, mList)
		if err != nil {
			return
		}

		err = setErrors(cfg, db, mList)
		if err != nil {
			return
		}
	}()

	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	title := "Последние значения"
	// Выводим
	switch r.Form.Get("tp") {
	case "json":
		j, _ := jsonw.Marshal(aList)
		err = htmlpage.Raw(cfg, prefix, w, r, errMsg, title, string(j))
		return

	default:
		err = showPage(cfg, prefix, w, r, errMsg, title, aList)
		return
	}
}

//----------------------------------------------------------------------------------------------------------------------------//

// loadLast -- Загружаем последние
func loadLast(cfg *config.Config, db *sql.DB) (mList dataMap, aList dataArr, err error) {
	mList = dataMap{}

	for _, e := range entity.GetAll() {
		cfg := e.Config()
		mList[cfg.ID] = &dataRow{
			ID:    cfg.ID,
			Name:  cfg.Name,
			Login: cfg.Login,
		}
	}

	rows, err := db.Query(`
	select h.entity_id, h.ts, h.data
		from history as h
			join (select max(id) as id from history group by entity_id) as h1 on h.id=h1.id;
	`)
	if err != nil {
		return
	}

	// Загружаем из базы

	for rows.Next() {
		var id uint
		var ts int64
		var s string

		err = rows.Scan(&id, &ts, &s)
		if err != nil {
			return
		}

		d, exists := mList[id]
		if !exists {
			continue
		}

		d.TS = time.Unix(ts, 0) // локальное

		err = jsonw.Unmarshal([]byte(s), &d.Info)
		if err != nil {
			return
		}

		d.LastChange = make(operator.FVals, len(d.Info.FVals))
		d.Ferror = make([]string, len(d.Info.FVals))
	}

	aList = make(dataArr, len(mList))

	nF := 0 // количество активных float колонок
	nS := 0 // количество активных string колонок

	i := 0
	for id, df := range mList {
		e := entity.GetByID(id)
		if e == nil || !e.Enabled() {
			// если неизвесное или не разрешено - удаляем
			delete(mList, id)
			continue
		}

		cfg := e.Config()
		df.Idx = cfg.Idx

		df.FLegend, df.SLegend = e.Legend()

		ln := len(df.Info.FVals)
		if ln > nF {
			// акивных float колонок стало больше
			nF = ln
		}

		msg := []string{}
		if ln > 0 {
			v := df.Info.FVals[0]
			if cfg.AlertLevelLow > 0 && v < float64(cfg.AlertLevelLow) {
				msg = append(msg, fmt.Sprintf("Меньше %d", cfg.AlertLevelLow))
			}
			if cfg.AlertLevelHigh > 0 && v > float64(cfg.AlertLevelHigh) {
				msg = append(msg, fmt.Sprintf("Больше %d", cfg.AlertLevelHigh))
			}
		}
		if len(msg) > 0 {
			df.Ferror[0] = "!!! " + strings.Join(msg, "; ") + " !!!"
		}

		ln = len(df.Info.SVals)
		if ln > nS {
			// активных string колонок стало больше
			nS = ln
		}

		aList[i] = df
		i++
	}

	// Ну хотя бы по одной колонке надо сделать

	if nF == 0 {
		nF = 1
	}

	if nS == 0 {
		nS = 1
	}

	for _, df := range mList {
		// сколько колонок надо будет дополнять при выводе до максимално отображаемых
		df.Ftail = nF - len(df.Info.FVals)
		df.Stail = nS - len(df.Info.SVals)

		// если легенды короче, чем данные, дополняем пробелами

		for {
			if len(df.FLegend) >= nF {
				break
			}
			df.FLegend = append(df.FLegend, "")
		}

		for {
			if len(df.SLegend) >= nS {
				break
			}
			df.SLegend = append(df.SLegend, "")
		}
	}

	// обрезаем отброшенные и сортируем
	aList = aList[:len(mList)]
	sort.Sort(aList)

	return
}

//----------------------------------------------------------------------------------------------------------------------------//

// loadPrev -- Загружаем предыдущие
func loadPrev(cfg *config.Config, db *sql.DB, mList dataMap) (err error) {
	rows, err := db.Query(`
		select h.entity_id, h.data
			from history as h
		  		join (select max(h2.id) as id
						from history as h2
							join (
								select max(id) as id, entity_id from history group by entity_id
								) as h3 on h2.entity_id=h3.entity_id and h2.id<h3.id
						group by h2.entity_id
					) as h1 on h.id=h1.id
		`)
	if err != nil {
		return
	}

	for rows.Next() {
		var id uint
		var s string
		var data operator.Data

		err = rows.Scan(&id, &s)
		if err != nil {
			return
		}

		err = jsonw.Unmarshal([]byte(s), &data)
		if err != nil {
			return
		}

		// вычисляем изменение от предыдущего

		df, exists := mList[id]
		if exists {
			for i, v := range data.FVals {
				if i >= len(df.LastChange) {
					break
				}
				df.LastChange[i] = df.Info.FVals[i] - v
			}
		}
	}

	return
}

//----------------------------------------------------------------------------------------------------------------------------//

// setErrors -- Загружаем ошибки, которые были после последнего успеха
func setErrors(cfg *config.Config, db *sql.DB, mList dataMap) (err error) {
	rows, err := db.Query(`
		select e.entity_id, e.ts, e.data
			from errors as e
				join (select max(e2.id) as id
						from errors as e2
							join (
								select max(ts) as ts, entity_id from history group by entity_id
							) as h on e2.entity_id=h.entity_id and e2.ts>h.ts
						group by h.entity_id
					) as e1 on e.id=e1.id
		`)
	if err != nil {
		return
	}

	for rows.Next() {
		var id uint
		var ts int64
		var s string

		err = rows.Scan(&id, &ts, &s)
		if err != nil {
			return
		}

		df, exists := mList[id]
		if exists {
			df.Error = fmt.Sprintf(`[%s] %s`, time.Unix(ts, 0).Format(misc.DateTimeFormatRev+" "+misc.DateTimeFormatTZ), s)
			e := entity.GetByID(id)
			if e != nil {
				e.Status().LastResult = operator.ResultError
			}
		}
	}

	return
}

//----------------------------------------------------------------------------------------------------------------------------//

func showPage(cfg *config.Config, prefix string, w http.ResponseWriter, r *http.Request, errMsg string, title string, aList dataArr) (err error) {
	params := struct {
		List      dataArr
		Fcount    int
		Scount    int
		ColsCount int
	}{
		List:      aList,
		Fcount:    1,
		Scount:    1,
		ColsCount: 6,
	}

	if len(aList) > 0 {
		// если есть хотя бы одна строка, вычислем общее количество колонок по типам
		params.Fcount = aList[0].Ftail + len(aList[0].Info.FVals)
		params.Scount = aList[0].Stail + len(aList[0].Info.SVals)
		params.ColsCount = 2 + 2*params.Fcount + params.Scount + 1
	}

	return htmlpage.Do("dashboard", prefix, w, r, errMsg, title, params)
}

//----------------------------------------------------------------------------------------------------------------------------//

// Len implements sort.Interface.
func (d dataArr) Len() int {
	return len(d)
}

// Less implements sort.Interface.
func (d dataArr) Less(i, j int) bool {
	return d[i].Idx < d[j].Idx
}

// Swap implements sort.Interface.
func (d dataArr) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

//----------------------------------------------------------------------------------------------------------------------------//
