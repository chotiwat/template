package template

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"encoding/base64"
	htmltemplate "html/template"
)

// New creates a new template.
func New() *Template {
	return &Template{
		vars: map[string]interface{}{},
		env:  parseEnvVars(os.Environ()),
	}
}

// NewFromFile creates a new template from a file.
func NewFromFile(filepath string) (*Template, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	contents, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return New().WithBody(string(contents)), nil
}

// Template is a wrapper for html.Template.
type Template struct {
	body    string
	vars    map[string]interface{}
	env     map[string]string
	helpers Helpers
}

// WithBody sets the template body and returns a reference to the template object.
func (t *Template) WithBody(body string) *Template {
	t.body = body
	return t
}

// Body returns the template body.
func (t *Template) Body() string {
	return t.body
}

// WithVar sets a variable and returns a reference to the template object.
func (t *Template) WithVar(key string, value interface{}) *Template {
	t.SetVar(key, value)
	return t
}

// SetVar sets a var in the template.
func (t *Template) SetVar(key string, value interface{}) {
	t.vars[key] = value
}

// HasVar returns if a variable is set.
func (t *Template) HasVar(key string) bool {
	_, hasKey := t.vars[key]
	return hasKey
}

// Var returns the value of a variable, or panics if the variable is not set.
func (t *Template) Var(key string, defaults ...interface{}) interface{} {
	if value, hasVar := t.vars[key]; hasVar {
		return value
	}

	if len(defaults) > 0 {
		return defaults[0]
	}

	panic(fmt.Sprintf("template variable `%s` is unset, cannot continue", key))
}

// Env returns an environment variable.
func (t *Template) Env(key string, defaults ...string) string {
	if value, hasVar := t.env[key]; hasVar {
		return value
	}

	if len(defaults) > 0 {
		return defaults[0]
	}

	panic(fmt.Sprintf("template env variable `%s` is unset, cannot continue", key))
}

// File returns the contents of a file.
func (t *Template) File(path string) string {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return string(contents)
}

// Helpers returns the helpers object.
func (t *Template) Helpers() *Helpers {
	return &t.helpers
}

// Process processes the template.
func (t *Template) Process(dst io.Writer) error {
	temp, err := htmltemplate.New("").Funcs(t.funcMap()).Parse(t.body)
	if err != nil {
		return err
	}
	return temp.Execute(dst, t)
}

func (t *Template) funcMap() htmltemplate.FuncMap {
	return htmltemplate.FuncMap{

		"time_unix": func(t time.Time) string {
			return fmt.Sprintf("%d", t.Unix())
		},
		"time_rfc3339": func(t time.Time) string {
			return t.Format(time.RFC3339)
		},
		"time_short": func(t time.Time) string {
			return t.Format("1/02/2006 3:04:05 PM")
		},
		"time_short_date": func(t time.Time) string {
			return t.Format("1/02/2006")
		},
		"time_medium": func(t time.Time) string {
			return t.Format("Jan 02, 2006 3:04:05 PM")
		},
		"time_kitchen": func(t time.Time) string {
			return t.Format(time.Kitchen)
		},
		"time_month_day": func(t time.Time) string {
			return t.Format("1/2")
		},
		"time_in": func(t time.Time, loc string) time.Time {
			location, err := time.LoadLocation(loc)
			if err != nil {
				panic(err)
			}
			return t.In(location)
		},

		"float64_money": func(d float64) string {
			return fmt.Sprintf("$%0.2f", d)
		},

		"base64": func(v string) string {
			src := []byte(v)
			var dst []byte
			base64.StdEncoding.Encode(dst, src)
			return string(dst)
		},
		"base64decode": func(v string) string {
			src := []byte(v)
			var dst []byte
			_, err := base64.StdEncoding.Decode(dst, src)
			if err != nil {
				panic(err)
			}
			return string(dst)
		},
	}
}

func parseEnvVars(envVars []string) map[string]string {
	vars := map[string]string{}
	for _, str := range envVars {
		parts := strings.Split(str, "=")
		if len(parts) > 1 {
			vars[parts[0]] = parts[1]
		}
	}
	return vars
}
