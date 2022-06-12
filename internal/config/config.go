package config

import (
	"fmt"
	"strings"

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
		CronLocation string          `toml:"cron-location"`
		Schedule     string          `toml:"schedule"`
		ViewBrowser  bool            `toml:"view-browser"`
		UserAgent    string          `toml:"user-agent"`
		StdTimeout   config.Duration `toml:"std-timeout"`

		DB string `toml:"db"`

		TemplatesDir string   `toml:"templates-dir"`
		Templates    []string `toml:"-"`
	}

	// Operators --
	Operators map[string]*Operator

	// Operator --
	Operator struct {
		Description string          `toml:"description"`
		Timeout     config.Duration `toml:"timeout"`
		Tasks       []string        `toml:"tasks"`
	}

	// Entities -- objects list
	Entities []*Entity

	// Entity -- object
	Entity struct {
		Idx            int             `toml:"-"`
		ID             uint            `toml:"id"`
		Enabled        bool            `toml:"enabled"`
		Name           string          `toml:"name"`
		Description    string          `toml:"description"`
		Type           string          `toml:"type"`
		Delay          config.Duration `toml:"delay"`
		AlertLevelHigh int             `toml:"alert-level-high"`
		AlertLevelLow  int             `toml:"alert-level-low"`
		Schedule       string          `toml:"schedule"`
		Vars           misc.StringMap  `toml:"vars"`
		// Устаревшие параметры, перенесены в Vars. Оставлено для совместимости
		DeprLogin    string `toml:"login"`
		DeprPassword string `toml:"password"`
	}
)

const (
	VarID = "id"
)

//----------------------------------------------------------------------------------------------------------------------------//

// Check -- check http listener config
func (x *HTTP) Check(cfg *Config) (err error) {
	msgs := misc.NewMessages()

	err = x.Listener.Check(cfg)
	if err != nil {
		msgs.Add("%s", err.Error())
	}

	return msgs.Error()
}

//----------------------------------------------------------------------------------------------------------------------------//

// Check -- check processor config
func (x *Processor) Check(cfg *Config) (err error) {
	msgs := misc.NewMessages()

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
		x.Templates[i], err = misc.AbsPath(fmt.Sprintf("%s/%s.tpl", x.TemplatesDir, n))
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
func (x Operators) Check(cfg *Config) (err error) {
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
func (x *Operator) Check(cfg *Config) (err error) {
	msgs := misc.NewMessages()

	if x.Timeout == 0 {
		x.Timeout = cfg.Processor.StdTimeout
	}

	return msgs.Error()
}

//----------------------------------------------------------------------------------------------------------------------------//

// Check -- check entities config
func (x Entities) Check(cfg *Config) (err error) {
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

		for n, v := range df.Vars {
			delete(df.Vars, n)
			df.Vars[strings.ToLower(n)] = v
		}

		// Для совместимости
		if df.DeprLogin != "" {
			df.Vars["login"] = df.DeprLogin
		}
		if df.DeprPassword != "" {
			df.Vars["password"] = df.DeprPassword
		}

		if df.Vars[VarID] == "" {
			df.Vars[VarID] = df.Vars["login"] // Если оно есть
		}

		knownNames[df.Name] = df
	}

	return msgs.Error()
}

//----------------------------------------------------------------------------------------------------------------------------//

// Check -- check entity config
func (x *Entity) Check(cfg *Config) (err error) {
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
