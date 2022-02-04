package handler

import (
	"encoding/base64"
	"fmt"
	"html"
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/prometheus/common/version"
)

type data struct {
	Birth            time.Time
	BuildInfo, Flags map[string]string
	PathPrefix       string
	counter          int
}

// func (d *data) Count() int {
// 	d.counter++
// 	return d.counter
// }

// func (data) FormatTimestamp(ts int64) string {
// 	return time.Unix(ts/1000, ts%1000*1000000).String()
// }

// Status serves the status page.
func Status(
	root http.FileSystem,
	flags map[string]string,
	pathPrefix string,
	logger log.Logger,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {

		birth := time.Now()

		t := template.New("status")
		t.Funcs(template.FuncMap{
			"value": func(f float64) string {
				return strconv.FormatFloat(f, 'f', -1, 64)
			},
			"timeFormat": func(t time.Time) string {
				return t.Format(time.RFC3339)
			},
			"base64": func(s string) string {
				return base64.RawURLEncoding.EncodeToString([]byte(s))
			},
		})

		f, err := root.Open("template.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			level.Error(logger).Log("msg", "error loading template.html", "err", err.Error())
			return
		}
		defer f.Close()

		tpl, err := ioutil.ReadAll(f)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			level.Error(logger).Log("msg", "error reading template.html", "err", err.Error())
			return
		}

		_, err = t.Parse(string(tpl))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			level.Error(logger).Log("msg", "error parsing template", "err", err.Error())
			return
		}

		buildInfo := map[string]string{
			"version":   version.Version,
			"revision":  version.Revision,
			"branch":    version.Branch,
			"buildUser": version.BuildUser,
			"buildDate": version.BuildDate,
			"goVersion": version.GoVersion,
		}

		d := &data{
			Birth:      birth,
			BuildInfo:  buildInfo,
			PathPrefix: pathPrefix,
			Flags:      flags,
		}

		err = t.Execute(w, d)
		if err != nil {
			// Hack to get a visible error message right at the top.
			fmt.Fprintf(w, `<div id="template-error" class="alert alert-danger">Error executing template: %s</div>`, html.EscapeString(err.Error()))
			fmt.Fprintln(w, `<script>$("#template-error").prependTo("body")</script>`)
		}

	})
}
