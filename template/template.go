package template

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"encoding/base64"
	"net/url"
	"strconv"
	texttemplate "text/template"
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

	return New().WithName(filepath).WithBody(string(contents)), nil
}

// Template is a wrapper for html.Template.
type Template struct {
	name    string
	body    string
	vars    map[string]interface{}
	env     map[string]string
	helpers Helpers
}

// WithName sets the template name.
func (t *Template) WithName(name string) *Template {
	t.name = name
	return t
}

// Name returns the template name if set, or if not set, just "template" as a constant.
func (t *Template) Name() string {
	if len(t.name) > 0 {
		return t.name
	}
	return "template"
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
func (t *Template) Var(key string, defaults ...interface{}) (interface{}, error) {
	if value, hasVar := t.vars[key]; hasVar {
		return value, nil
	}

	if len(defaults) > 0 {
		return defaults[0], nil
	}

	return nil, fmt.Errorf("template variable `%s` is unset and no default is provided", key)
}

// Env returns an environment variable.
func (t *Template) Env(key string, defaults ...string) (string, error) {
	if value, hasVar := t.env[key]; hasVar {
		return value, nil
	}

	if len(defaults) > 0 {
		return defaults[0], nil
	}

	return "", fmt.Errorf("template env variable `%s` is unset and no default is provided", key)
}

// File returns the contents of a file.
func (t *Template) File(path string) (string, error) {
	contents, err := ioutil.ReadFile(path)
	return string(contents), err
}

// Helpers returns the helpers object.
func (t *Template) Helpers() *Helpers {
	return &t.helpers
}

// Process processes the template.
func (t *Template) Process(dst io.Writer) error {
	temp, err := texttemplate.New(t.Name()).Funcs(t.funcMap()).Parse(t.body)
	if err != nil {
		return err
	}
	return temp.Execute(dst, t)
}

func (t *Template) funcMap() texttemplate.FuncMap {
	return texttemplate.FuncMap{
		"unix": func(t time.Time) string {
			return fmt.Sprintf("%d", t.Unix())
		},
		"rfc3339": func(t time.Time) string {
			return t.Format(time.RFC3339)
		},
		"short": func(t time.Time) string {
			return t.Format("1/02/2006 3:04:05 PM")
		},
		"short_date": func(t time.Time) string {
			return t.Format("1/02/2006")
		},
		"medium": func(t time.Time) string {
			return t.Format("Jan 02, 2006 3:04:05 PM")
		},
		"kitchen": func(t time.Time) string {
			return t.Format(time.Kitchen)
		},
		"month_day": func(t time.Time) string {
			return t.Format("1/2")
		},
		"in": func(loc string, t time.Time) time.Time {
			location, err := time.LoadLocation(loc)
			if err != nil {
				panic(err)
			}
			return t.In(location)
		},
		"time": func(format, v string) (time.Time, error) {
			return time.Parse(format, v)
		},
		"time_unix": func(v string) (t time.Time, err error) {
			var value int64
			value, err = strconv.ParseInt(v, 10, 64)
			if err != nil {
				return
			}

			t = time.Unix(value, 0)
			return
		},
		"year": func(t time.Time) int {
			return t.Year()
		},
		"month": func(t time.Time) int {
			return int(t.Month())
		},
		"day": func(t time.Time) int {
			return t.Day()
		},
		"hour": func(t time.Time) int {
			return t.Hour()
		},
		"minute": func(t time.Time) int {
			return t.Minute()
		},
		"second": func(t time.Time) int {
			return t.Second()
		},
		"millisecond": func(t time.Time) int {
			return int(time.Duration(t.Nanosecond()) / time.Millisecond)
		},

		"float64": func(v string) (float64, error) {
			return strconv.ParseFloat(v, 64)
		},
		"money": func(d float64) string {
			return fmt.Sprintf("$%0.2f", d)
		},

		"base64": func(v string) string {
			return base64.StdEncoding.EncodeToString([]byte(v))
		},
		"base64decode": func(v string) (string, error) {
			result, err := base64.StdEncoding.DecodeString(v)
			if err != nil {
				return "", err
			}
			return string(result), nil
		},

		// string transforms
		"upper": func(v string) string {
			return strings.ToUpper(v)
		},
		"lower": func(v string) string {
			return strings.ToLower(v)
		},
		"title": func(v string) string {
			return strings.ToTitle(v)
		},
		"trim": func(v string) string {
			return strings.TrimSpace(v)
		},

		// string tests
		"has_suffix": func(suffix, v string) bool {
			return strings.HasSuffix(v, suffix)
		},
		"has_prefix": func(prefix, v string) bool {
			return strings.HasPrefix(v, prefix)
		},
		"contains": func(v, substr string) bool {
			return strings.Contains(v, substr)
		},

		// url transforms and helpers
		"url": func(v string) (*url.URL, error) {
			return url.Parse(v)
		},
		"proto": func(v *url.URL) string {
			return v.Scheme
		},
		"host": func(v *url.URL) string {
			return v.Host
		},
		"port": func(v *url.URL) string {
			portValue := v.Port()
			if len(portValue) > 0 {
				return portValue
			}
			switch strings.ToLower(v.Scheme) {
			case "http":
				return "80"
			case "https":
				return "443"
			case "ssh":
				return "22"
			case "ftp":
				return "21"
			case "sftp":
				return "22"
			}
			return ""
		},
		"path": func(v *url.URL) string {
			return v.Path
		},
		"rawquery": func(v *url.URL) string {
			return v.RawQuery
		},
		"query": func(name string, v *url.URL) string {
			return v.Query().Get(name)
		},

		"sha1": func(v string) string {
			h := sha1.New()
			io.WriteString(h, v)
			return fmt.Sprintf("%x", h.Sum(nil))
		},
		"sha256": func(v string) string {
			h := sha256.New()
			io.WriteString(h, v)
			return fmt.Sprintf("%x", h.Sum(nil))
		},
		"sha512": func(v string) string {
			h := sha512.New()
			io.WriteString(h, v)
			return fmt.Sprintf("%x", h.Sum(nil))
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
