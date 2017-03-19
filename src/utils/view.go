package utils

import (
	"errors"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/lfkeitel/verbose"
)

type Views struct {
	source string
	t      *template.Template
	e      *Environment
}

func NewViews(e *Environment, basepath string) (v *Views, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		}
	}()

	tmpl := template.New("").Funcs(template.FuncMap{
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, errors.New("invalid dict call")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, errors.New("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},

		"list": func(values ...interface{}) ([]interface{}, error) {
			return values, nil
		},

		"plus1": func(a int) int {
			return a + 1
		},
	})

	filepath.Walk(basepath, func(path string, info os.FileInfo, err1 error) error {
		if strings.HasSuffix(path, ".tmpl") {
			if _, err := tmpl.ParseFiles(path); err != nil {
				panic(err)
			}
		}
		return nil
	})

	v = &Views{
		source: basepath,
		t:      tmpl,
		e:      e,
	}

	return v, nil
}

func (v *Views) NewView(view string, r *http.Request) *View {
	return &View{
		name: view,
		t:    v.t,
		e:    v.e,
		r:    r,
	}
}

func (v *Views) Reload() error {
	views, err := NewViews(v.e, v.source)
	if err != nil {
		return err
	}
	v.t = views.t
	return nil
}

func (v *Views) RenderError(w http.ResponseWriter, r *http.Request, data map[string]interface{}) {
	if data == nil {
		v.NewView("error", r).Render(w, nil)
		return
	}
	v.NewView("custom-error", r).Render(w, data)
}

type View struct {
	name string
	t    *template.Template
	e    *Environment
	r    *http.Request
}

func (v *View) Render(w http.ResponseWriter, data map[string]interface{}) {
	if data == nil {
		data = make(map[string]interface{})
	}
	session := GetSessionFromContext(v.r)
	flashes := session.Flashes()
	flash := ""
	if len(flashes) > 0 {
		flash = flashes[0].(string)
	}
	if session.GetString("username") != "" {
		session.Save(v.r, w)
	}
	data["config"] = v.e.Config
	data["flashMessage"] = flash
	if err := v.t.ExecuteTemplate(w, v.name, data); err != nil {
		v.e.Log.WithFields(verbose.Fields{
			"template": v.name,
			"error":    err,
			"package":  "common:views",
		}).Error("Error rendering template")
	}
}
