package htmlpage

import (
	"bytes"
	"html/template"
	"io"
	"net/http"
	"sync"

	"github.com/alrusov/misc"

	"github.com/alrusov/balance-collector/internal/config"
)

//----------------------------------------------------------------------------------------------------------------------------//

var (
	mutex        = new(sync.Mutex)
	cache        = map[string]*template.Template{}
	cacheEnabled = false
)

//----------------------------------------------------------------------------------------------------------------------------//

// Do --
func Do(name string, prefix string, w http.ResponseWriter, r *http.Request, errMsg string, title string, data interface{}) (err error) {
	cfg := config.Get()

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
		Data      interface{}
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

	w.WriteHeader(http.StatusOK)
	_, err = io.Copy(w, buf)

	return
}

//----------------------------------------------------------------------------------------------------------------------------//

// JSON --
func JSON(cfg *config.Config, prefix string, w http.ResponseWriter, r *http.Request, errMsg string, title string, data string) (err error) {
	return Do("json", prefix, w, r, errMsg, title, data)
}

//----------------------------------------------------------------------------------------------------------------------------//
