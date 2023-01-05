package htmlpage

import (
	"bytes"
	"html/template"
	"net/http"
	"sync"

	"github.com/alrusov/initializer"
	"github.com/alrusov/log"
	"github.com/alrusov/misc"
	"github.com/alrusov/stdhttp"

	"github.com/alrusov/balance-collector/internal/config"
)

//----------------------------------------------------------------------------------------------------------------------------//

var (
	// Log --
	Log = log.NewFacility("htmlpage")

	cfg *config.Config

	mutex        = new(sync.Mutex)
	cache        = map[string]*template.Template{}
	cacheEnabled = false
)

//----------------------------------------------------------------------------------------------------------------------------//

func init() {
	// Регистрируем инициализатор
	initializer.RegisterModuleInitializer(initModule)
}

// Инициализация
func initModule(appCfg any, h any) (err error) {
	cfg = appCfg.(*config.Config)

	Log.Message(log.INFO, "Initialized")
	return
}

//----------------------------------------------------------------------------------------------------------------------------//

// Do --
func Do(name string, prefix string, w http.ResponseWriter, r *http.Request, errMsg string, title string, data any) (err error) {
	mutex.Lock()
	locked := true

	defer func() {
		if locked {
			mutex.Unlock()
		}
	}()

	t, exists := cache[name]

	if !exists {
		fn := ""
		fn, err = misc.AbsPath(cfg.Processor.TemplatesDir + "/" + name + ".tpl")
		if err != nil {
			return
		}

		t, err = template.ParseFiles(fn)
		if err != nil {
			return
		}

		fn, err = misc.AbsPath(cfg.Processor.TemplatesDir + "/snippets/*.tpl")
		if err != nil {
			return
		}

		t, err = t.ParseGlob(fn)
		if err != nil {
			return
		}

		if cacheEnabled {
			cache[name] = t
		}
	}

	t, err = t.Clone()
	if err != nil {
		return
	}

	mutex.Unlock()
	locked = false

	params := struct {
		Prefix    string
		Base      string
		Error     string
		Title     string
		Copyright string
		Name      string
		App       string
		Version   string
		Tags      string
		Data      any
	}{
		Prefix:    prefix,
		Base:      r.URL.Path,
		Error:     errMsg,
		Title:     title,
		Copyright: misc.Copyright(),
		Name:      cfg.Common.Name,
		App:       misc.AppName(),
		Version:   misc.AppVersion(),
		Tags:      misc.AppTags(),
		Data:      data,
	}

	buf := new(bytes.Buffer)

	err = t.ExecuteTemplate(buf, name, params)
	if err != nil {
		return
	}

	stdhttp.WriteReply(w, r, http.StatusOK, stdhttp.ContentTypeHTML, nil, bytes.TrimSpace(buf.Bytes()))
	return
}

//----------------------------------------------------------------------------------------------------------------------------//

// Raw --
func Raw(cfg *config.Config, prefix string, w http.ResponseWriter, r *http.Request, errMsg string, title string, data string) (err error) {
	return Do("raw", prefix, w, r, errMsg, title, data)
}

//----------------------------------------------------------------------------------------------------------------------------//

// Graph --
func Graph(cfg *config.Config, prefix string, w http.ResponseWriter, r *http.Request, errMsg string, title string, data string) (err error) {
	return Do("graph", prefix, w, r, errMsg, title, data)
}

//----------------------------------------------------------------------------------------------------------------------------//
