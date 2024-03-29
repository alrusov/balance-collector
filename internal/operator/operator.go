package operator

import (
	"fmt"
	"time"

	"github.com/alrusov/initializer"
	"github.com/alrusov/log"
	"github.com/alrusov/misc"

	"github.com/alrusov/balance-collector/internal/chrome"
	"github.com/alrusov/balance-collector/internal/config"
)

//----------------------------------------------------------------------------------------------------------------------------//

type (
	// Operator --
	Operator struct {
		cfg    *config.Operator
		Name   string
		chrome *chrome.Chrome
	}

	// Status --
	Status struct {
		LastResult Result
		LastUpdate time.Time
	}

	// Data --
	Data chrome.Data

	// FVals --
	FVals chrome.FVals

	// SVals --
	SVals chrome.SVals

	// Result --
	Result uint
)

const (
	// ResultUnknown --
	ResultUnknown Result = iota
	// ResultInProgress --
	ResultInProgress
	// ResultOK --
	ResultOK
	// ResultError --
	ResultError
)

var (
	// Log --
	Log = log.NewFacility("operator")

	operators = make(map[string]*Operator)
)

//----------------------------------------------------------------------------------------------------------------------------//

func init() {
	// Регистрируем инициализатор
	initializer.RegisterModuleInitializer(initModule)
}

// Инициализация
func initModule(appCfg any, h any) (err error) {
	cfg := appCfg.(*config.Config)

	msgs := misc.NewMessages()

	for name, cfg := range cfg.Operators {
		o := &Operator{
			cfg:  cfg,
			Name: name,
		}

		o.chrome, err = chrome.New(cfg.Tasks)
		if err != nil {
			msgs.Add("%s - %s", name, err.Error())
			continue
		}

		operators[name] = o
	}

	err = msgs.Error()
	if err != nil {
		return
	}

	Log.Message(log.INFO, "Initialized")
	return
}

//----------------------------------------------------------------------------------------------------------------------------//

// Get --
func Get(name string) *Operator {
	return operators[name]
}

//----------------------------------------------------------------------------------------------------------------------------//

// Legend --
func (o *Operator) Legend() ([]string, []string) {
	return o.chrome.Legend()
}

//----------------------------------------------------------------------------------------------------------------------------//

// UpdateInfo --
func (o *Operator) UpdateInfo(status *Status, entityCfg *config.Entity) (info *Data, err error) {
	fn := fmt.Sprintf("%s.Get(%s)", o.Name, entityCfg.Name)
	Log.MessageWithSource(log.DEBUG, fn, "Begin")

	status.LastResult = ResultInProgress

	defer func() {
		if err == nil {
			status.LastResult = ResultOK
		} else {
			status.LastResult = ResultError
		}
		Log.MessageWithSource(log.DEBUG, fn, "End")
	}()

	data, err := o.chrome.Prepare(entityCfg)
	if err != nil {
		return
	}

	err = data.Exec(time.Duration(o.cfg.Timeout))
	if err != nil {
		return
	}

	d := data.Data()
	info = &Data{
		FVals: d.FVals,
		SVals: d.SVals,
	}

	return
}

//----------------------------------------------------------------------------------------------------------------------------//
