package processor

import (
	"fmt"

	"github.com/alrusov/appcron"
	"github.com/alrusov/initializer"
	"github.com/alrusov/log"
	"github.com/alrusov/stdhttp"
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

var (
	// Log --
	Log = log.NewFacility("processor")
)

//----------------------------------------------------------------------------------------------------------------------------//

func init() {
	// Регистрируем инициализатор
	initializer.RegisterModuleInitializer(initModule)
}

// Инициализация
func initModule(appCfg interface{}, h *stdhttp.HTTP) (err error) {
	cfg := appCfg.(*config.Config)

	err = appcron.Init("", cfg.Processor.CronLocation)
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

	Log.Message(log.INFO, "Initialized")
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
