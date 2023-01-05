package storage

import (
	"database/sql"
	"os"

	"github.com/alrusov/initializer"
	"github.com/alrusov/jsonw"
	"github.com/alrusov/log"
	"github.com/alrusov/misc"

	"github.com/alrusov/balance-collector/internal/config"
	"github.com/alrusov/balance-collector/internal/operator"
)

//----------------------------------------------------------------------------------------------------------------------------//

var (
	// Log --
	Log = log.NewFacility("storage")

	fileName string
)

//----------------------------------------------------------------------------------------------------------------------------//

func init() {
	// Регистрируем инициализатор
	initializer.RegisterModuleInitializer(initModule)
}

// Инициализация
func initModule(appCfg any, h any) (err error) {
	cfg := appCfg.(*config.Config)

	fileName = cfg.Processor.DB

	fd, fileErr := os.Open(fileName)
	if fd != nil {
		fd.Close()
	}

	db, err := sql.Open("sqlite3", fileName)
	if err != nil {
		return
	}
	defer db.Close()

	if fileErr != nil {
		Log.Message(log.NOTICE, `Database "%s" is not found, try to create new...`, fileName)

		query := `
			create table config(version int);
			insert into config(version) values(1);

			create table history(
				id integer primary key autoincrement,
				entity_id	int,
				ts			int,
				data		text
			);
			create index i_history_complex1 on history(entity_id, ts);

			create table errors(
				id integer primary key autoincrement,
				entity_id	int,
				ts		int,
				data	text
			);
			create index i_errors_complex1 on errors(entity_id, ts);
		`
		_, err = db.Exec(query)
		if err != nil {
			return
		}

	}

	Log.Message(log.INFO, "Initialized")
	return
}

//----------------------------------------------------------------------------------------------------------------------------//

// SaveToHistory --
func SaveToHistory(entityID uint, data *operator.Data) (err error) {
	j, err := jsonw.Marshal(data)
	if err != nil {
		return
	}

	js := string(j)
	ts := misc.NowUTC().Unix()

	Log.Message(log.DEBUG, `SaveToHistory(%d, %d, %q)`, entityID, ts, js)

	query := `insert into history(entity_id, ts, data) values(?, ?, ?)`

	return Save(query, entityID, ts, js)
}

//----------------------------------------------------------------------------------------------------------------------------//

// SaveToErrors --
func SaveToErrors(entityID uint, data any) (err error) {
	j, err := jsonw.Marshal(data)
	if err != nil {
		return
	}

	js := string(j)
	ts := misc.NowUTC().Unix()

	Log.Message(log.DEBUG, `SaveErrors(%d, %d, %q)`, entityID, ts, js)

	query := `insert into errors(entity_id, ts, data) values(?, ?, ?)`

	return Save(query, entityID, ts, js)
}

//----------------------------------------------------------------------------------------------------------------------------//

// Save --
func Save(query string, args ...any) (err error) {
	db, err := sql.Open("sqlite3", fileName)
	if err != nil {
		return
	}
	defer db.Close()

	stmt, err := db.Prepare(query)
	if err != nil {
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(args...)
	if err != nil {
		return
	}

	return
}

//----------------------------------------------------------------------------------------------------------------------------//
