package config

import (
	"github.com/alrusov/config"
	"github.com/alrusov/misc"
)

//----------------------------------------------------------------------------------------------------------------------------//

type (
	// Config -- application config
	Config struct {
		Common    config.Common `toml:"common"`
		Listener  HTTP          `toml:"http"`
		Processor Processor     `toml:"processor"`
		Operators Operators     `toml:"operators"`
		Entities  Entities      `toml:"entities"`
	}

	// HTTP -- http listener config
	HTTP struct {
		Listener config.Listener `toml:"listener"`
	}

	// Processor -- processor options
	Processor struct {
		TimeZoneInfoFile string `toml:"tz-info-file"`
		CronLocation     string `toml:"cron-location"`
		Schedule         string `toml:"schedule"`
		ViewBrowser      bool   `toml:"view-browser"`
		StdTimeout       uint   `toml:"std-timeout"`

		DB string `toml:"db"`

		TemplatesDir string   `toml:"templates-dir"`
		Templates    []string `toml:"-"`
	}

	// Operators --
	Operators map[string]*Operator

	// Operator --
	Operator struct {
		Description string   `toml:"description"`
		Timeout     uint     `toml:"timeout"`
		Tasks       []string `toml:"tasks"`
	}

	// Entities -- objects list
	Entities []*Entity

	// Entity -- object
	Entity struct {
		Idx            int    `toml:"-"`
		ID             uint   `toml:"id"`
		Enabled        bool   `toml:"enabled"`
		Name           string `toml:"name"`
		Description    string `toml:"description"`
		Type           string `toml:"type"`
		Delay          uint   `toml:"delay"`
		AlertLevelHigh int    `toml:"alert-level-high"`
		AlertLevelLow  int    `toml:"alert-level-low"`
		Login          string `toml:"login"`
		Password       string `toml:"password"`
		Schedule       string `toml:"schedule"`
	}
)

//----------------------------------------------------------------------------------------------------------------------------//

// Check -- check http listener config
func (x *HTTP) Check(cfg *Config) error {
	msgs := misc.NewMessages()

	err := x.Listener.Check(cfg)
	if err != nil {
		msgs.Add("%s", err.Error())
	}

	return msgs.Error()
}

//----------------------------------------------------------------------------------------------------------------------------//

// Check -- check processor config
func (x *Processor) Check(cfg *Config) error {
	msgs := misc.NewMessages()
	var err error

	if x.DB == "" {
		msgs.Add("processor.db is empty")
	} else {
		x.DB, err = misc.AbsPath(x.DB)
		if err != nil {
			msgs.Add("processor.db - %s", err.Error())
		}
	}

	if x.TemplatesDir == "" {
		msgs.Add("processor.templates-dir is empty")
	} else {
		x.TemplatesDir, err = misc.AbsPath(x.TemplatesDir)
		if err != nil {
			msgs.Add("processor.templates-dir - %s", err.Error())
		}
	}

	x.Templates = []string{
		"header",
		"footer",
		"back",
	}

	for i, n := range x.Templates {
		x.Templates[i], err = misc.AbsPath(x.TemplatesDir + "/" + n + ".tpl")
		if err != nil {
			msgs.Add("processor.template %s - %s", n, err.Error())
		}
	}

	if x.StdTimeout == 0 {
		msgs.Add("processor.std-timeout is zero")
	}

	return msgs.Error()
}

//----------------------------------------------------------------------------------------------------------------------------//

// Check -- check operators config
func (x Operators) Check(cfg *Config) error {
	msgs := misc.NewMessages()

	for name, df := range x {
		err := df.Check(cfg)
		if err != nil {
			msgs.Add("Operator %s - %s", name, err.Error())
		}
	}

	return msgs.Error()
}

//----------------------------------------------------------------------------------------------------------------------------//

// Check -- check operators config
func (x *Operator) Check(cfg *Config) error {
	msgs := misc.NewMessages()

	if x.Timeout == 0 {
		x.Timeout = cfg.Processor.StdTimeout
	}

	return msgs.Error()
}

//----------------------------------------------------------------------------------------------------------------------------//

// Check -- check entities config
func (x Entities) Check(cfg *Config) error {
	msgs := misc.NewMessages()

	knownIDs := make(map[uint]*Entity, len(x))
	knownNames := make(map[string]*Entity, len(x))

	for i, df := range x {
		df.Idx = i

		if !df.Enabled {
			continue
		}

		err := df.Check(cfg)
		if err != nil {
			msgs.Add("Entity #%d (%s) - %s", i+1, df.Name, err.Error())
		}

		prev, exists := knownIDs[df.ID]
		if exists {
			msgs.Add(`Entity #%d - ID=%d already defined in block %d`, i+1, df.ID, prev.Idx)
		}

		prev, exists = knownNames[df.Name]
		if exists {
			msgs.Add(`Entity #%d - "%s" already defined in block %d`, i+1, df.Name, prev.Idx)
		}

		knownNames[df.Name] = df
	}

	return msgs.Error()
}

//----------------------------------------------------------------------------------------------------------------------------//

// Check -- check entity config
func (x *Entity) Check(cfg *Config) error {
	msgs := misc.NewMessages()

	if x.Name == "" {
		msgs.Add("name is empty")
	}

	if x.Type == "" {
		msgs.Add("type is empty")
	}

	if x.AlertLevelLow >= x.AlertLevelHigh && x.AlertLevelHigh != 0 {
		msgs.Add("alert-level-low >= alert-level-high")
	}

	return msgs.Error()
}

//----------------------------------------------------------------------------------------------------------------------------//

// Check -- check application config
func (cfg *Config) Check() (err error) {
	appcfg = cfg
	return config.Check(
		cfg,
		[]interface{}{
			&cfg.Common,
			&cfg.Listener,
			&cfg.Processor,
			&cfg.Entities,
			&cfg.Operators,
		},
	)
}

//----------------------------------------------------------------------------------------------------------------------------//

var appcfg *Config

// Get --
func Get() *Config {
	return appcfg
}

//----------------------------------------------------------------------------------------------------------------------------//
