package entity

import (
	"fmt"
	"sync"

	"github.com/alrusov/log"
	"github.com/alrusov/misc"

	"github.com/alrusov/balance-collector/internal/config"
	"github.com/alrusov/balance-collector/internal/operator"
	"github.com/alrusov/balance-collector/internal/storage"
)

//----------------------------------------------------------------------------------------------------------------------------//

type (
	// Entity --
	Entity struct {
		cfg      *config.Entity
		operator *operator.Operator
		status   *operator.Status
	}
)

var (
	// Log --
	Log = log.NewFacility("entity")

	mutex      = new(sync.Mutex)
	inProgress = false

	entities = []*Entity{}
)

//----------------------------------------------------------------------------------------------------------------------------//

// Init --
func Init() (err error) {
	msgs := misc.NewMessages()

	for _, cfg := range config.Get().Entities {
		if !cfg.Enabled {
			continue
		}

		op := operator.Get(cfg.Type)
		if op == nil {
			cfg.Enabled = false
			msgs.Add(`%s: Unknown type "%s", entity will be disabled`, cfg.Name, cfg.Type)
			continue
		}

		entities = append(entities,
			&Entity{
				cfg:      cfg,
				operator: op,
				status: &operator.Status{
					LastResult: operator.ResultUnknown,
				},
			},
		)
	}

	return msgs.Error()

}

//----------------------------------------------------------------------------------------------------------------------------//

// GetByID --
func GetByID(id uint) (e *Entity) {
	for _, e = range entities {
		if e.cfg.ID == id {
			return
		}
	}

	return nil
}

//----------------------------------------------------------------------------------------------------------------------------//

// GetByName --
func GetByName(name string) (e *Entity) {
	for _, e = range entities {
		if e.cfg.Name == name {
			return
		}
	}

	return nil
}

//----------------------------------------------------------------------------------------------------------------------------//

// GetAll --
func GetAll() (e []*Entity) {
	return entities
}

//----------------------------------------------------------------------------------------------------------------------------//

// Enabled --
func (e *Entity) Enabled() bool {
	return e.cfg.Enabled
}

//----------------------------------------------------------------------------------------------------------------------------//

// Config --
func (e *Entity) Config() *config.Entity {
	return e.cfg
}

//----------------------------------------------------------------------------------------------------------------------------//

// Status --
func (e *Entity) Status() *operator.Status {
	return e.status
}

//----------------------------------------------------------------------------------------------------------------------------//

// Legend --
func (e *Entity) Legend() ([]string, []string) {
	return e.operator.Legend()
}

//----------------------------------------------------------------------------------------------------------------------------//

// List --
//func List() []*Entity {
//	return entities
//}

//----------------------------------------------------------------------------------------------------------------------------//

// Update --
func Update(fn string, isBatch bool) {
	t0 := misc.NowUnixNano()

	mutex.Lock()
	alreadyInProgress := inProgress
	if !alreadyInProgress {
		inProgress = true
	}
	mutex.Unlock()

	if alreadyInProgress {
		Log.MessageWithSource(log.INFO, fn, "Already in the process, skipping")
		return
	}

	Log.MessageWithSource(log.INFO, fn, "Started")

	for i, e := range entities {
		if isBatch && e.Config().Schedule != "" {
			continue
		}
		e.Update(fn, isBatch && i > 0)
	}

	inProgress = false
	misc.LogProcessingTime(Log.Name(), "INFO", 0, fn, "", t0)
}

//----------------------------------------------------------------------------------------------------------------------------//

// UpdateByID --
func UpdateByID(fn string, id uint) {
	e := GetByID(id)
	if e != nil {
		e.Update(fn, false)
	}
}

//----------------------------------------------------------------------------------------------------------------------------//

// Update --
func (e *Entity) Update(fn string, isBatch bool) {
	fn = fmt.Sprintf("%s: %s", fn, e.cfg.Name)

	if !e.cfg.Enabled {
		Log.MessageWithSource(log.DEBUG, fn, `Entity is disabled`)
		return
	}

	if isBatch && e.cfg.Delay > 0 {
		Log.MessageWithSource(log.DEBUG, fn, `Sleep %d seconds...`, e.cfg.Delay)
		misc.Sleep(e.cfg.Delay)
	}

	t0 := misc.NowUnixNano()
	Log.MessageWithSource(log.DEBUG, fn, `Processing`)

	data, err := e.operator.UpdateInfo(e.status, e.cfg)
	if err != nil {
		Log.MessageWithSource(log.ERR, fn, `Get: %s`, err.Error())
		m := struct {
			Error string `json:"error"`
		}{
			Error: err.Error(),
		}
		err = storage.SaveToErrors(e.cfg.ID, &m)
		if err != nil {
			Log.MessageWithSource(log.ERR, fn, `SaveToErrors: %s`, err.Error())
		}

	} else {
		err = storage.SaveToHistory(e.cfg.ID, data)
		if err != nil {
			Log.MessageWithSource(log.ERR, fn, `SaveToHistory: %s`, err.Error())
		}
	}

	misc.LogProcessingTime(Log.Name(), "", 0, fn, "", t0)
	return
}

//----------------------------------------------------------------------------------------------------------------------------//

// Repeat --
func Repeat(fn string) {
	first := true
	for _, e := range entities {
		if e.status.LastResult == operator.ResultError {
			e.Update(fn, !first)
			first = false
		}
	}
}

//----------------------------------------------------------------------------------------------------------------------------//
