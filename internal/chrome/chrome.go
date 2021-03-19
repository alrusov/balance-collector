package chrome

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"

	"github.com/alrusov/log"
	"github.com/alrusov/misc"

	"github.com/alrusov/balance-collector/internal/config"
)

//----------------------------------------------------------------------------------------------------------------------------//

type (
	// Chrome --
	Chrome struct {
		tasks          []*CallParams
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
	}

	// Method --
	Method struct {
		paramsCount int
	}

	// ExecData --
	ExecData struct {
		entityCfg      *config.Entity
		tasks          []chromedp.Action
		results        [maxResultCount]result
		resultsCount   int
		skippedResults map[int]bool
		data           Data
	}

	result struct {
		v       string
		tp      resultType
		nodeIds *[]cdp.NodeID
		nodeIdx int
		re      *regexp.Regexp
		reIdx   int
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
)

//----------------------------------------------------------------------------------------------------------------------------//

// New --
func New(src []string) (c *Chrome, err error) {
	c = &Chrome{}

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

var (
	reTask   = regexp.MustCompile(`^\s*(?U:(\S+)\s*\((.*)\))\s*$`)
	reParams = regexp.MustCompile(`(\\,|[^,])*`)
)

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
			msgs.Add(`Found %d params, but method "%s" expect %d at least`, nParams, cp.methodName, cp.method.paramsCount)
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
				if !cp.parseRE(msgs, params[2:]) {
					continue
				}
				optsStart += 2
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
		resultsCount:   0,
		skippedResults: c.skippedResults,
	}

	for _, df := range c.tasks {
		var task chromedp.Action
		tp := resultTypeUnknown

		switch df.methodName {
		case mSleep:
			v, _ := strconv.ParseInt(df.param, 10, 64)
			task = chromedp.Sleep(time.Duration(v))

		case mNavigate:
			q := df.node
			q = strings.Replace(q, "{Login}", entityCfg.Login, -1)
			q = strings.Replace(q, "{Password}", entityCfg.Password, -1)
			task = chromedp.Navigate(q)

		case mWaitVisible:
			task = chromedp.WaitVisible(df.node, df.options...)

		case mWaitNotVisible:
			task = chromedp.WaitNotVisible(df.node, df.options...)

		case mClear:
			task = chromedp.Clear(df.node, df.options...)

		case mSetValue, mSendKeys:
			v := ""
			switch df.param {
			case "$Login":
				v = entityCfg.Login
			case "$Password":
				v = entityCfg.Password
			default:
				v = df.param
			}

			switch df.methodName {
			case mSetValue:
				task = chromedp.SetValue(df.node, v, df.options...)
			case mSendKeys:
				task = chromedp.SendKeys(df.node, v, df.options...)
			}

		case mClick:
			task = chromedp.Click(df.node, df.options...)

		case mSubmit:
			task = chromedp.Submit(df.node, df.options...)

		case mFloat, mMultiFloat:
			tp = resultTypeFloat
			fallthrough
		case mString, mMultiString:
			if tp == resultTypeUnknown {
				tp = resultTypeString
			}

			nodeIds := []cdp.NodeID{}

			for i := 0; i < df.resultsCount; i++ {
				if r.resultsCount == len(r.results) {
					err = fmt.Errorf(`Too many results`)
					return
				}

				res := &r.results[r.resultsCount]
				res.nodeIdx = i

				if df.node == "" {
					task = nil
				} else {
					switch df.methodName {
					case mFloat, mString:
						task = chromedp.Text(df.node, &res.v, df.options...)
						break

					case mMultiFloat, mMultiString:
						if i == 0 {
							r.tasks = append(r.tasks, chromedp.NodeIDs(df.node, &nodeIds))
						}

						res.nodeIds = &nodeIds

						task = chromedp.ActionFunc(
							func(ctx context.Context) (err error) {
								if res.nodeIds != nil && res.nodeIdx < len(*res.nodeIds) {
									res.v, err = dom.GetOuterHTML().WithNodeID((*res.nodeIds)[res.nodeIdx]).Do(ctx)
								}
								return
							},
						)
					}
				}

				res.tp = tp
				res.re = df.re
				res.reIdx = df.reIdx

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
func (r *ExecData) Exec(timeout uint) (err error) {
	headless := !config.Get().Processor.ViewBrowser

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", headless),
		chromedp.Flag("disable-gpu", headless),
		chromedp.Flag("enable-automation", true),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-default-browser-check", true),
		chromedp.Flag("hide-scrollbars", true),
		chromedp.Flag("mute-audio", true),
		chromedp.Flag("disable-background-networking", true),
		chromedp.Flag("enable-features", "NetworkService,NetworkServiceInProcess"),
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-backgrounding-occluded-windows", true),
		chromedp.Flag("disable-breakpad", true),
		chromedp.Flag("disable-client-side-phishing-detection", true),
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-features", "site-per-process,TranslateUI,BlinkGenPropertyTrees"),
		chromedp.Flag("disable-hang-monitor", true),
		chromedp.Flag("disable-ipc-flooding-protection", true),
		chromedp.Flag("disable-popup-blocking", true),
		chromedp.Flag("disable-prompt-on-repost", true),
		chromedp.Flag("disable-renderer-backgrounding", true),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("force-color-profile", "srgb"),
		chromedp.Flag("metrics-recording-only", true),
		chromedp.Flag("safebrowsing-disable-auto-update", true),
		chromedp.Flag("password-store", "basic"),
		chromedp.Flag("use-mock-keychain", true),
	)

	userAgent := config.Get().Processor.UserAgent
	if userAgent != "" {
		opts = append(opts, chromedp.Flag("user-agent", userAgent))
	}

	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	logOptions := []chromedp.ContextOption{
		chromedp.WithErrorf(
			func(fmt string, i ...interface{}) {
				Log.MessageWithSource(log.ERR, "Exec", fmt, i...)
			},
		),
		chromedp.WithLogf(
			func(fmt string, i ...interface{}) {
				Log.MessageWithSource(log.INFO, "Exec", fmt, i...)
			},
		),
	}

	if Log.CurrentLogLevel() >= log.TRACE4 {
		logOptions = append(logOptions,
			chromedp.WithDebugf(
				func(fmt string, i ...interface{}) {
					Log.MessageWithSource(log.TRACE4, "Exec", fmt, i...)
				},
			),
		)
	}

	ctx, cancel = chromedp.NewContext(ctx, logOptions...)

	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
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
		if r.skippedResults[i] {
			continue
		}

		if res.re != nil {
			x := res.re.FindAllStringSubmatch(res.v, -1)
			if len(x) > 0 && len(x[0]) > res.reIdx {
				res.v = x[0][res.reIdx]
			}
		}

		switch res.tp {
		case resultTypeFloat:
			v, e := Float(res.v)
			r.data.FVals = append(r.data.FVals, v)
			if e != nil {
				Log.MessageWithSource(log.ERR, r.entityCfg.Name, "%s: %s", res.v, e.Error())
				continue
			}

		case resultTypeString:
			r.data.SVals = append(r.data.SVals, res.v)

		default:
			break
		}

		if dataOK || res.v != "" {
			dataFound = true
		}
	}

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
