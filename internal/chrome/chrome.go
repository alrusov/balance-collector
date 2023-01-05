package chrome

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/oliveagle/jsonpath"

	"github.com/alrusov/initializer"
	"github.com/alrusov/jsonw"
	"github.com/alrusov/log"
	"github.com/alrusov/misc"

	"github.com/alrusov/balance-collector/internal/config"
)

//----------------------------------------------------------------------------------------------------------------------------//

type (
	// Chrome --
	Chrome struct {
		tasks          []*CallParams
		options        misc.StringMap
		fLegend        []string
		sLegend        []string
		skippedResults map[int]bool
	}

	// CallParams --
	CallParams struct {
		methodName string
		method     *Method

		node    string
		param   string
		options []chromedp.QueryOption

		resultsCount int
		re           *regexp.Regexp
		reIdx        int

		jsonPath string
	}

	// Method --
	Method struct {
		paramsCount int
	}

	// ExecData --
	ExecData struct {
		entityCfg      *config.Entity
		options        misc.StringMap
		tasks          []chromedp.Action
		results        [maxResultCount]result
		vars           misc.StringMap
		resultsCount   int
		skippedResults map[int]bool
		data           Data
	}

	result struct {
		cp       *CallParams
		v        []string
		tp       resultType
		nodeIdx  int
		jsonData any
	}

	resultType uint

	// Data --
	Data struct {
		FVals FVals `json:"fVals"`
		SVals SVals `json:"sVals"`
	}

	// FVals --
	FVals []float64

	// SVals --
	SVals []string
)

const (
	maxResultCount = 32
)

const (
	resultTypeUnknown resultType = iota
	resultTypeFloat
	resultTypeString
)

const (
	mOption         = "Option"
	mSleep          = "Sleep"
	mNavigate       = "Navigate"
	mWaitVisible    = "WaitVisible"
	mWaitNotVisible = "WaitNotVisible"
	mClear          = "Clear"
	mSetValue       = "SetValue"
	mSendKeys       = "SendKeys"
	mClick          = "Click"
	mSubmit         = "Submit"
	mFloat          = "Float"
	mString         = "String"
	mMultiFloat     = "MultiFloat"
	mMultiString    = "MultiString"
)

var (
	methods = map[string]*Method{
		mOption:         {paramsCount: 1}, // name, value
		mSleep:          {paramsCount: 1}, // duration
		mNavigate:       {paramsCount: 1}, // url
		mWaitVisible:    {paramsCount: 1}, // selector, [ options... ]
		mWaitNotVisible: {paramsCount: 1}, // selector, [ options... ]
		mClear:          {paramsCount: 1}, // selector, [ options... ]
		mSetValue:       {paramsCount: 2}, // selector, value, [ options... ]
		mSendKeys:       {paramsCount: 2}, // selector, value, [ options... ]
		mClick:          {paramsCount: 1}, // selector, [ options... ]
		mSubmit:         {paramsCount: 1}, // selector, [ options... ]
		mFloat:          {paramsCount: 2}, // selector, caption, [ re, reIdx, options... ]
		mString:         {paramsCount: 2}, // selector, caption, [ re, reIdx, options... ]
		mMultiFloat:     {paramsCount: 2}, // selector, count, [count]caption..., [ re, reIdx, options... ]
		mMultiString:    {paramsCount: 2}, // selector, count, [count]caption..., [ re, reIdx, options... ]
	}

	options = map[string]chromedp.QueryOption{
		"ByQuery":    chromedp.ByQuery,
		"ByQueryAll": chromedp.ByQueryAll,
		"ByID":       chromedp.ByID,
		"BySearch":   chromedp.BySearch,
		"ByJSPath":   chromedp.ByJSPath,
		"ByNodeID":   chromedp.ByNodeID,
	}
)

var (
	// Log --
	Log = log.NewFacility("chrome")

	cfg *config.Config

	reTask   = regexp.MustCompile(`^\s*(?U:(\S+)\s*\((.*)\))\s*$`)
	reParams = regexp.MustCompile(`(\\,|[^,])*`)
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

// New --
func New(src []string) (c *Chrome, err error) {
	c = &Chrome{
		options: make(misc.StringMap),
	}

	err = c.parseTaskDef(src)
	if err != nil {
		c = nil
		return
	}

	return
}

//----------------------------------------------------------------------------------------------------------------------------//

// Legend --
func (c *Chrome) Legend() ([]string, []string) {
	return c.fLegend, c.sLegend
}

//----------------------------------------------------------------------------------------------------------------------------//

func (c *Chrome) parseTaskDef(src []string) (err error) {
	msgs := misc.NewMessages()
	list := []*CallParams{}

	resultIdx := 0

	if c.skippedResults == nil {
		c.skippedResults = map[int]bool{}
	}

	for _, s := range src {
		df := reTask.FindAllStringSubmatch(s, -1)

		if len(df) != 1 || len(df[0]) < 2 {
			msgs.Add(`Illegal task "%s"`, s)
			continue
		}

		params := reParams.FindAllString(df[0][2], -1)

		if params[0] == "" {
			msgs.Add(`Empty first parameter in "%s"`, s)
			continue
		}

		cp := &CallParams{
			methodName: strings.TrimSpace(df[0][1]),
		}

		exists := false
		cp.method, exists = methods[cp.methodName]
		if !exists {
			msgs.Add(`Unknown method "%s"`, cp.methodName)
			continue
		}

		nParams := len(params)

		if cp.method.paramsCount > nParams {
			msgs.Add(`Found %d params, but %d at least expected for the method "%s" `, nParams, cp.method.paramsCount, cp.methodName)
			continue
		}

		for i, p := range params {
			params[i] = strings.TrimSpace(p)
		}

		var legend *[]string
		optsStart := cp.method.paramsCount

		cp.node = params[0]

		switch cp.methodName {
		//-----//
		case mOption:
			c.options[params[0]] = params[1]
			continue

		//-----//
		case mSleep:
			v, err := strconv.ParseInt(params[0], 10, 64)
			if err != nil {
				msgs.Add(`%s`, err.Error())
				continue
			}
			cp.param = fmt.Sprintf("%d", v*int64(time.Millisecond))

		//-----//
		case mSetValue, mSendKeys:
			cp.param = params[1]

		//-----//
		case mFloat:
			legend = &c.fLegend
			fallthrough
		case mString:
			if legend == nil {
				legend = &c.sLegend
			}

			*legend = append(*legend, params[1])

			if nParams > 2 {
				p2 := params[2]
				if strings.HasPrefix(p2, "json(") {
					if p2[len(p2)-1] != ')' {
						msgs.Add("illegal json() format: %s", p2)
						continue
					}

					cp.jsonPath = strings.TrimSpace(p2[len("json(") : len(p2)-1])
					optsStart++

				} else if !cp.parseRE(msgs, params[2:]) {
					continue
				} else {
					optsStart += 2
				}
			}

			cp.resultsCount = 1
			resultIdx++

		//-----//
		case mMultiFloat:
			legend = &c.fLegend
			fallthrough
		case mMultiString:
			if legend == nil {
				legend = &c.sLegend
			}

			cp.resultsCount, err = strconv.Atoi(params[1])
			if err != nil {
				msgs.Add(`count: %s`, err.Error())
				continue
			}

			if nParams < 2+cp.resultsCount {
				msgs.Add(`found %d captions, expected %d`, nParams-2, cp.resultsCount)
				continue
			}

			for i := 0; i < cp.resultsCount; i++ {
				caption := params[2+i]
				if caption == "" {
					c.skippedResults[resultIdx] = true
				} else {
					*legend = append(*legend, caption)
				}
				resultIdx++
			}

			optsStart += cp.resultsCount

			if nParams > optsStart {
				if !cp.parseRE(msgs, params[optsStart:]) {
					continue
				}
				optsStart += 2
			}
		}

		if optsStart < nParams {
			for _, name := range params[optsStart:] {
				o, exists := options[name]
				if !exists {
					msgs.Add(`Unknown option "%s"`, name)
					continue
				}
				cp.options = append(cp.options, o)
			}
		}

		list = append(list, cp)
	}

	err = msgs.Error()

	if err == nil {
		c.tasks = list
	}

	return
}

func (cp *CallParams) parseRE(msgs *misc.Messages, params []string) bool {
	var err error

	if len(params) < 2 {
		msgs.Add(`Found regexp without followed index param`)
		return false
	}

	if params[0] != "" {
		cp.re, err = regexp.Compile(params[0])
		if err != nil {
			msgs.Add(`regexp: %s`, err.Error())
			return false
		}

		cp.reIdx, err = strconv.Atoi(params[1])
		if err != nil {
			msgs.Add(`regexp index: %s`, err.Error())
			return false
		}
	}

	return true
}

//----------------------------------------------------------------------------------------------------------------------------//

// Prepare --
func (c *Chrome) Prepare(entityCfg *config.Entity) (r *ExecData, err error) {
	r = &ExecData{
		entityCfg:      entityCfg,
		vars:           make(misc.StringMap, len(entityCfg.Vars)),
		resultsCount:   0,
		skippedResults: c.skippedResults,
	}

	// Копируем, так как в перспективе переменные могут создаваться в процессе выполнения
	for n, v := range entityCfg.Vars {
		r.vars[n] = v
	}

	r.tasks = append(r.tasks,
		chromedp.ActionFunc(
			func(cxt context.Context) error {
				_, err := page.AddScriptToEvaluateOnNewDocument("Object.defineProperty(navigator, 'webdriver', { get: () => false, });").Do(cxt)
				if err != nil {
					return err
				}
				return nil
			},
		),
	)

	r.options = c.options

	for _, cp := range c.tasks {
		var task chromedp.Action
		tp := resultTypeUnknown

		switch cp.methodName {
		case mSleep:
			v, _ := strconv.ParseInt(cp.param, 10, 64)
			task = chromedp.Sleep(time.Duration(v))

		case mNavigate:
			q := cp.node

			// Заменяем только имеющиеся переменные, отсутствующие не трогаем, они могут быть заполнены позже в процессе исполнения
			for n, v := range r.vars {
				var re *regexp.Regexp
				re, err = regexp.Compile(`(?i)\{` + n + `\}`)
				if err != nil {
					return
				}

				q = re.ReplaceAllString(q, v)
			}

			task = chromedp.Navigate(q)

		case mWaitVisible:
			task = chromedp.WaitVisible(cp.node, cp.options...)

		case mWaitNotVisible:
			task = chromedp.WaitNotVisible(cp.node, cp.options...)

		case mClear:
			task = chromedp.Clear(cp.node, cp.options...)

		case mSetValue, mSendKeys:
			v := cp.param
			if v[0] == '$' {
				if vv, exists := r.vars[strings.ToLower(v[1:])]; exists {
					// Заменяем только имеющиеся переменные, отсутствующие не трогаем, они могут быть заполнены позже в просессе исполнения
					v = vv
				}
			}

			switch cp.methodName {
			case mSetValue:
				task = chromedp.SetValue(cp.node, v, cp.options...)
			case mSendKeys:
				task = chromedp.SendKeys(cp.node, v, cp.options...)
			}

		case mClick:
			task = chromedp.Click(cp.node, cp.options...)

		case mSubmit:
			task = chromedp.Submit(cp.node, cp.options...)

		case mFloat, mMultiFloat:
			tp = resultTypeFloat
			fallthrough
		case mString, mMultiString:
			if tp == resultTypeUnknown {
				tp = resultTypeString
			}

			for i := 0; i < cp.resultsCount; i++ {
				if r.resultsCount == len(r.results) {
					err = fmt.Errorf(`too many results`)
					return
				}

				res := &r.results[r.resultsCount]
				res.nodeIdx = i

				js := fmt.Sprintf(`[document.querySelectorAll('%s')[%d]].map((e) => e.innerText)`, strings.Replace(cp.node, `'`, `\'`, -1), res.nodeIdx)
				task = chromedp.Evaluate(js, &res.v)

				res.cp = cp
				res.tp = tp

				r.resultsCount++

				if task != nil {
					r.tasks = append(r.tasks, task)
					task = nil
				}
			}
		}

		if task != nil {
			r.tasks = append(r.tasks, task)
		}
	}

	return
}

//----------------------------------------------------------------------------------------------------------------------------//

// Exec --
func (r *ExecData) Exec(timeout time.Duration) (err error) {
	headless := !cfg.Processor.ViewBrowser

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", headless),
		chromedp.Flag("disable-gpu", headless),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-extensions", false),
		chromedp.WindowSize(1280, 800),
		chromedp.Flag("hide-scrollbars", false),
	//	chromedp.Flag("remote-debugging-port", "9222"),
	//	chromedp.Flag("user-data-dir", "remote-profile"),
	)

	for n, v := range r.options {
		opts = append(opts, chromedp.Flag(n, v))
	}

	userAgent := cfg.Processor.UserAgent
	if userAgent != "" {
		opts = append(opts, chromedp.Flag("user-agent", userAgent))
	}

	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	logOptions := []chromedp.ContextOption{
		chromedp.WithErrorf(
			func(fmt string, i ...any) {
				Log.MessageWithSource(log.ERR, "Exec", fmt, i...)
			},
		),
		chromedp.WithLogf(
			func(fmt string, i ...any) {
				Log.MessageWithSource(log.INFO, "Exec", fmt, i...)
			},
		),
	}

	if Log.CurrentLogLevel() >= log.TRACE4 {
		logOptions = append(logOptions,
			chromedp.WithDebugf(
				func(fmt string, i ...any) {
					Log.MessageWithSource(log.TRACE4, "Exec", fmt, i...)
				},
			),
		)
	}

	ctx, cancel = chromedp.NewContext(ctx, logOptions...)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, timeout)
	defer cancel()

	defer func() {
		chromedp.Stop()
		// Usually a "defer cancel()" will be enough for most use cases. However, Cancel
		// is the better option if one wants to gracefully close a browser, or catch
		// underlying errors happening during cancellation.
		chromedp.Cancel(ctx)
	}()

	msgs := misc.NewMessages()

	err = chromedp.Run(ctx, r.tasks...)
	if err != nil {
		msgs.AddError(err)
	}

	dataFound, convErr := r.convResults(err == nil)

	if convErr != nil {
		msgs.AddError(convErr)
	}

	if dataFound {
		err = msgs.Error()
		if err != nil {
			fn := fmt.Sprintf("Get(%s)", r.entityCfg.Name)
			Log.MessageWithSource(log.ERR, fn, "%s", err)
		}
		return nil
	}

	msgs.Add("No data found")

	return msgs.Error()

}

//----------------------------------------------------------------------------------------------------------------------------//

func (r *ExecData) convResults(dataOK bool) (dataFound bool, err error) {
	for i, res := range r.results {
		if r.skippedResults[i] || res.cp == nil {
			continue
		}

		if len(res.v) == 0 {
			switch res.tp {
			case resultTypeFloat:
				r.data.FVals = append(r.data.FVals, 0)
			case resultTypeString:
				r.data.SVals = append(r.data.SVals, "")
			}
			continue
		}

		if res.cp.re != nil {
			x := res.cp.re.FindAllStringSubmatch(res.v[0], -1)
			if len(x) > 0 && len(x[0]) > res.cp.reIdx {
				res.v[0] = x[0][res.cp.reIdx]
			}
		} else if res.cp.jsonPath != "" {
			e := res.extractJson(r.entityCfg.Name)
			if e != nil {
				Log.MessageWithSource(log.ERR, r.entityCfg.Name, "%s: %s", res.v, e.Error())
				continue
			}
		}

		switch res.tp {
		case resultTypeFloat:
			v, e := Float(res.v[0])
			r.data.FVals = append(r.data.FVals, v)
			if e != nil {
				Log.MessageWithSource(log.ERR, r.entityCfg.Name, "%s: %s", res.v, e.Error())
				continue
			}

		case resultTypeString:
			r.data.SVals = append(r.data.SVals, res.v[0])

		default:
		}

		if dataOK || res.v[0] != "" {
			dataFound = true
		}
	}

	return
}

//----------------------------------------------------------------------------------------------------------------------------//

func (res *result) extractJson(entityName string) (err error) {
	if len(res.v) == 0 {
		err = fmt.Errorf("no data")
		return
	}

	if res.jsonData == nil {
		err = jsonw.Unmarshal(misc.UnsafeString2ByteSlice(res.v[0]), &res.jsonData)
		if err != nil {
			Log.MessageWithSource(log.ERR, entityName, "%s: %s", res.v, err.Error())
			return
		}
	}

	v, err := jsonpath.JsonPathLookup(res.jsonData, res.cp.jsonPath)
	if err != nil {
		Log.MessageWithSource(log.ERR, entityName, "%s: %s", res.v, err.Error())
		return
	}

	res.v[0] = fmt.Sprint(v)

	return
}

//----------------------------------------------------------------------------------------------------------------------------//

// Data --
func (r *ExecData) Data() Data {
	return r.data
}

//----------------------------------------------------------------------------------------------------------------------------//

var (
	reMinus       = regexp.MustCompile(`-|−|&minus;|&#8722;|&#x2212;`)
	reOnlyNumbers = regexp.MustCompile(`[^0-9\.-]`)
)

// Float --
func Float(s string) (v float64, err error) {
	s = strings.ReplaceAll(s, ",", ".")
	s = reMinus.ReplaceAllString(s, "-")
	s = reOnlyNumbers.ReplaceAllString(s, "")
	s = strings.TrimRight(s, ".")

	v, err = strconv.ParseFloat(s, 64)
	if err != nil {
		return
	}

	return
}

//----------------------------------------------------------------------------------------------------------------------------//
