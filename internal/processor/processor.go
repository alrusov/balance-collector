package processor

import (
	"fmt"

	"github.com/alrusov/appcron"
	cron "github.com/robfig/cron/v3"

	"github.com/alrusov/balance-collector/internal/config"
	"github.com/alrusov/balance-collector/internal/entity"
)

//----------------------------------------------------------------------------------------------------------------------------//

type (
	// Job --
	Job struct {
		id     cron.EntryID
		entity *entity.Entity
	}
)

//----------------------------------------------------------------------------------------------------------------------------//

// Init --
func Init() (err error) {
	cfg := config.Get()

	err = appcron.Init(cfg.Processor.TimeZoneInfoFile, cfg.Processor.CronLocation)
	if err != nil {
		return
	}

	j := &Job{
		entity: nil,
	}
	j.id, err = appcron.Add(cfg.Processor.Schedule, j)
	if err != nil {
		return
	}

	for _, e := range entity.GetAll() {
		cfg := e.Config()
		if cfg.Schedule != "" {
			j := &Job{
				entity: e,
			}
			j.id, err = appcron.Add(cfg.Schedule, j)
			if err != nil {
				return
			}
		}
	}

	return
}

//----------------------------------------------------------------------------------------------------------------------------//

// Run --
func (j *Job) Run() {
	fn := fmt.Sprintf("job %d", j.id)

	if j.entity == nil {
		entity.Update(fn, true)
		return
	}

	j.entity.Update(fn, true)
}

//----------------------------------------------------------------------------------------------------------------------------//
