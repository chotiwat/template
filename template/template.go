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

	yaml "gopkg.in/yaml.v2"

	"encoding/base64"
	"encoding/json"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	texttemplate "text/template"
)

// New creates a new template.
func New() *Template {
	temp := &Template{
		vars: map[string]interface{}{},
		env:  parseEnvVars(os.Environ()),
	}
	temp.funcs = temp.baseFuncMap()
	return temp
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
	name     string
	body     string
	vars     map[string]interface{}
	env      map[string]string
	includes []string
	funcs    texttemplate.FuncMap
	helpers  Helpers
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

// WithInclude includes a (sub) template into the rendering assets.
func (t *Template) WithInclude(body string) *Template {
	t.includes = append(t.includes, body)
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

// HasEnv returns if an env var is set.
func (t *Template) HasEnv(key string) bool {
	_, hasKey := t.env[key]
	return hasKey
}

// File returns the contents of a file.
func (t *Template) File(path string) (string, error) {
	contents, err := ioutil.ReadFile(path)
	return string(contents), err
}

// HasFile returns if a file exists.
func (t *Template) HasFile(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Helpers returns the helpers object.
func (t *Template) Helpers() *Helpers {
	return &t.helpers
}

// Process processes the template.
func (t *Template) Process(dst io.Writer) error {
	base := texttemplate.New(t.Name()).Funcs(t.ViewFuncs())

	var err error
	for _, include := range t.includes {
		_, err = base.New(t.Name()).Parse(include)
		if err != nil {
			return err
		}
	}

	final, err := base.New(t.Name()).Parse(t.body)
	if err != nil {
		return err
	}
	return final.Execute(dst, t)
}

// ViewFuncs returns the view funcs.
func (t *Template) ViewFuncs() texttemplate.FuncMap {
	return t.funcs
}

func (t *Template) baseFuncMap() texttemplate.FuncMap {
	return texttemplate.FuncMap{
		"string": func(v interface{}) string {
			return fmt.Sprintf("%v", v)
		},

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
		"in": func(loc string, t time.Time) (time.Time, error) {
			location, err := time.LoadLocation(loc)
			if err != nil {
				return time.Time{}, err
			}
			return t.In(location), err
		},
		"time": func(format, v string) (time.Time, error) {
			return time.Parse(format, v)
		},
		"time_unix": func(v int64) time.Time {
			return time.Unix(v, 0)
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

		"bool": func(raw interface{}) (bool, error) {
			v := fmt.Sprintf("%v", raw)
			if len(v) == 0 {
				return false, nil
			}
			switch strings.ToLower(v) {
			case "true", "1", "yes":
				return true, nil
			case "false", "0", "no":
				return false, nil
			default:
				return false, fmt.Errorf("invalid boolean value `%s`", v)
			}
		},
		"int": func(v interface{}) (int, error) {
			return strconv.Atoi(fmt.Sprintf("%v", v))
		},
		"int64": func(v interface{}) (int64, error) {
			return strconv.ParseInt(fmt.Sprintf("%v", v), 10, 64)
		},
		"float": func(v string) (float64, error) {
			return strconv.ParseFloat(v, 64)
		},

		"money": func(d float64) string {
			return fmt.Sprintf("$%0.2f", d)
		},
		"pct": func(d float64) string {
			return fmt.Sprintf("%0.2f%%", d*100)
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
		"prefix": func(pref, v string) string {
			return pref + v
		},
		"suffix": func(suf, v string) string {
			return v + suf
		},

		"split": func(sep, v string) []string {
			return strings.Split(v, sep)
		},

		"slice": func(from, to int, collection interface{}) (interface{}, error) {
			value := reflect.ValueOf(collection)

			if value.Type().Kind() != reflect.Slice {
				return nil, fmt.Errorf("input must be a slice")
			}

			return value.Slice(from, to).Interface(), nil
		},
		"first": func(collection interface{}) (interface{}, error) {
			value := reflect.ValueOf(collection)
			if value.Type().Kind() != reflect.Slice {
				return nil, fmt.Errorf("input must be a slice")
			}
			if value.Len() == 0 {
				return nil, nil
			}
			return value.Index(0).Interface(), nil
		},
		"at": func(index int, collection interface{}) (interface{}, error) {
			value := reflect.ValueOf(collection)
			if value.Type().Kind() != reflect.Slice {
				return nil, fmt.Errorf("input must be a slice")
			}
			if value.Len() == 0 {
				return nil, nil
			}
			return value.Index(index).Interface(), nil
		},
		"last": func(collection interface{}) (interface{}, error) {
			value := reflect.ValueOf(collection)
			if value.Type().Kind() != reflect.Slice {
				return nil, fmt.Errorf("input must be a slice")
			}
			if value.Len() == 0 {
				return nil, nil
			}
			return value.Index(value.Len() - 1).Interface(), nil
		},
		"join": func(sep string, collection interface{}) (string, error) {
			value := reflect.ValueOf(collection)
			if value.Type().Kind() != reflect.Slice {
				return "", fmt.Errorf("input must be a slice")
			}
			if value.Len() == 0 {
				return "", nil
			}
			values := make([]string, value.Len())
			for i := 0; i < value.Len(); i++ {
				values[i] = fmt.Sprintf("%v", value.Index(i).Interface())
			}
			return strings.Join(values, sep), nil
		},

		// string tests
		"has_suffix": func(suffix, v string) bool {
			return strings.HasSuffix(v, suffix)
		},
		"has_prefix": func(prefix, v string) bool {
			return strings.HasPrefix(v, prefix)
		},
		"contains": func(substr, v string) bool {
			return strings.Contains(v, substr)
		},
		"matches": func(expr, v string) (bool, error) {
			return regexp.MatchString(expr, v)
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

		"semver": func(v string) (*Semver, error) {
			return NewSemver(v)
		},
		"major": func(v *Semver) int {
			return int(v.Major)
		},
		"minor": func(v *Semver) int {
			return int(v.Minor)
		},
		"patch": func(v *Semver) int {
			return int(v.Patch)
		},
		"prerelease": func(v *Semver) string {
			return string(v.PreRelease)
		},

		"yaml": func(v interface{}) (string, error) {
			data, err := yaml.Marshal(v)
			return string(data), err
		},

		"json": func(v interface{}) (string, error) {
			data, err := json.Marshal(v)
			return string(data), err
		},

		"indent": func(tabCount int, corpus string) string {
			lines := strings.Split(corpus, "\n")
			outputLines := make([]string, len(lines))

			var tabs string
			for i := 0; i < tabCount; i++ {
				tabs = tabs + "\t"
			}

			for i := 0; i < len(lines); i++ {
				outputLines[i] = tabs + lines[i]
			}
			return strings.Join(outputLines, "\n")
		},

		"indentSpaces": func(spaceCount int, corpus string) string {
			lines := strings.Split(corpus, "\n")
			outputLines := make([]string, len(lines))

			var spaces string
			for i := 0; i < spaceCount; i++ {
				spaces = spaces + " "
			}

			for i := 0; i < len(lines); i++ {
				outputLines[i] = spaces + lines[i]
			}
			return strings.Join(outputLines, "\n")
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
