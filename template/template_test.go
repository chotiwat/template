package template

import (
	"bytes"
	"testing"
	"time"

	"strconv"

	"fmt"
	"os"

	"strings"

	assert "github.com/blendlabs/go-assert"
)

func TestTemplateFromFile(t *testing.T) {
	assert := assert.New(t)

	temp, err := NewFromFile("testdata/test.template.yml")
	assert.Nil(err)

	temp = temp.
		WithVar("name", "test-service").
		WithVar("container-name", "nginx").
		WithVar("container-image", "nginx:1.7.9").
		WithVar("container-port", "disabled")

	buffer := bytes.NewBuffer(nil)
	err = temp.Process(buffer)
	assert.Nil(err)

	result := buffer.String()
	assert.True(strings.Contains(result, "name: test-service"))
	assert.True(strings.Contains(result, "replicas: 2"))
	assert.False(strings.Contains(result, "containerPort:"))

	temp = temp.WithVar("container-port", 80)
	err = temp.Process(buffer)
	assert.Nil(err)
	result = buffer.String()
	assert.True(strings.Contains(result, "port: 80"))
}

func TestTemplateInclude(t *testing.T) {
	assert := assert.New(t)

	main := `{{ template "test" . }}`
	test := `{{ define "test" }}{{ .Var "foo" }}{{end}}`

	buffer := bytes.NewBuffer(nil)
	err := New().WithBody(main).WithInclude(test).WithVar("foo", "bar").Process(buffer)
	assert.Nil(err)
	assert.Equal("bar", buffer.String())

	buffer = bytes.NewBuffer(nil)
	err = New().WithBody(main).WithVar("foo", "bar").Process(buffer)
	assert.NotNil(err)
}

func TestTemplateVar(t *testing.T) {
	assert := assert.New(t)

	test := `{{ .Var "foo" }}`
	temp := New().WithBody(test).WithVar("foo", "bar")

	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.Equal("bar", buffer.String())
}

func TestTemplateVarMissing(t *testing.T) {
	assert := assert.New(t)

	test := `{{ .Var "baz" }}`
	temp := New().WithBody(test).WithVar("foo", "bar")

	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.NotNil(err)
}

func TestTemplateEnv(t *testing.T) {
	assert := assert.New(t)

	varName := UUIDv4().String()
	os.Setenv(varName, "bar")
	defer os.Unsetenv(varName)

	test := fmt.Sprintf(`{{ .Env "%s" }}`, varName)
	temp := New().WithBody(test)

	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.Equal("bar", buffer.String())
}

func TestTemplateHasEnv(t *testing.T) {
	assert := assert.New(t)

	varName := UUIDv4().String()
	os.Setenv(varName, "bar")
	defer os.Unsetenv(varName)

	test := fmt.Sprintf(`{{ if .HasEnv "%s" }}yep{{end}}`, varName)
	temp := New().WithBody(test)

	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.Equal("yep", buffer.String())
}

func TestTemplateHasEnvFalsey(t *testing.T) {
	assert := assert.New(t)

	varName := UUIDv4().String()

	test := fmt.Sprintf(`{{ if .HasEnv "%s" }}yep{{else}}nope{{end}}`, varName)
	temp := New().WithBody(test)

	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.Equal("nope", buffer.String())
}

func TestTemplateEnvMissing(t *testing.T) {
	assert := assert.New(t)

	varName := UUIDv4().String()

	test := fmt.Sprintf(`{{ .Env "%s" }}`, varName)
	temp := New().WithBody(test)

	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.NotNil(err)
}

func TestTemplateFile(t *testing.T) {
	assert := assert.New(t)

	test := `{{ .File "testdata/inline_file" }}`
	temp := New().WithBody(test)

	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.Equal("this is a test", buffer.String())
}

func TestTemplateHasFile(t *testing.T) {
	assert := assert.New(t)

	test := `{{ if .HasFile "testdata/inline_file" }}yep{{end}}`
	temp := New().WithBody(test)

	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.Equal("yep", buffer.String())
}

func TestTemplateHasFileFalsey(t *testing.T) {
	assert := assert.New(t)

	fileName := UUIDv4().String()

	test := fmt.Sprintf(`{{ if .HasFile "testdata/%s" }}yep{{else}}nope{{end}}`, fileName)
	temp := New().WithBody(test)

	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.Equal("nope", buffer.String())
}

func TestTemplateViewFuncTimeUnix(t *testing.T) {
	assert := assert.New(t)

	test := `{{ .Var "now" | unix }}`
	temp := New().WithBody(test).WithVar("now", time.Date(2017, 05, 20, 21, 00, 00, 00, time.UTC))

	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.Equal("1495314000", buffer.String())
}

func TestTemplateHelpersUTCNow(t *testing.T) {
	assert := assert.New(t)

	test := `{{ .Helpers.UTCNow | unix }}`
	temp := New().WithBody(test)

	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)

	parsed, err := strconv.ParseInt(buffer.String(), 10, 64)
	assert.Nil(err)
	assert.NotZero(parsed)
}

func TestTemplateHelpersCreateKey(t *testing.T) {
	assert := assert.New(t)

	test := `{{ .Helpers.CreateKey 64 }}`
	temp := New().WithBody(test)

	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)

	assert.True(len(buffer.String()) > 64)
}

func TestTemplateViewFuncString(t *testing.T) {
	assert := assert.New(t)

	test := `{{ .Var "foo" | string }}`
	temp := New().WithBody(test).WithVar("foo", 123)

	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.Equal("123", buffer.String())
}

func TestTemplateViewFuncTime(t *testing.T) {
	assert := assert.New(t)

	test := `{{ .Var "foo" | time "2006/01/02" | day }}`
	temp := New().WithBody(test).WithVar("foo", "2017/05/30")

	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.Equal("30", buffer.String())
}
func TestTemplateViewFuncTimeFromUnix(t *testing.T) {
	assert := assert.New(t)

	test := `{{ .Var "foo" | time_unix | year }}`
	temp := New().WithBody(test).WithVar("foo", time.Date(2017, 05, 20, 21, 00, 00, 00, time.UTC).Unix())

	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.Equal("2017", buffer.String())
}

func TestTemplateViewFuncTimeFromUnixString(t *testing.T) {
	assert := assert.New(t)

	test := `{{ .Var "foo" | string | int64 | time_unix | year }}`
	temp := New().WithBody(test).WithVar("foo", time.Date(2017, 05, 20, 21, 00, 00, 00, time.UTC).Unix())

	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.Equal("2017", buffer.String())
}

func TestTemplateViewFuncBool(t *testing.T) {
	assert := assert.New(t)

	test := `{{ if .Var "foo" | bool }}yep{{end}}`
	temp := New().WithBody(test).WithVar("foo", "true")

	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.Equal("yep", buffer.String())
}

func TestTemplateViewFuncInt(t *testing.T) {
	assert := assert.New(t)

	test := `{{ if .Var "foo" | int | eq 123 }}yep{{end}}`
	temp := New().WithBody(test).WithVar("foo", "123")

	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.Equal("yep", buffer.String())
}

func TestTemplateViewFuncInt64(t *testing.T) {
	assert := assert.New(t)

	test := `{{ .Var "foo" | int64 }}`
	temp := New().WithBody(test).WithVar("foo", fmt.Sprintf("%d", (1<<33)))

	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.Equal("8589934592", buffer.String())
}

func TestTemplateViewFuncFloat(t *testing.T) {
	assert := assert.New(t)

	test := `{{ .Var "foo" | float }}`
	temp := New().WithBody(test).WithVar("foo", "3.14")

	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.Equal("3.14", buffer.String())
}

func TestTemplateViewFuncMoney(t *testing.T) {
	assert := assert.New(t)

	test := `{{ .Var "foo" | float | money }}`
	temp := New().WithBody(test).WithVar("foo", "3.00")

	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.Equal("$3.00", buffer.String())
}

func TestTemplateViewFuncPct(t *testing.T) {
	assert := assert.New(t)

	test := `{{ .Var "foo" | float | pct }}`
	temp := New().WithBody(test).WithVar("foo", "0.24")

	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.Equal("24.00%", buffer.String())
}

func TestTemplateViewFuncBase64(t *testing.T) {
	assert := assert.New(t)

	test := `{{ .Var "foo" | base64 }}`
	temp := New().WithBody(test).WithVar("foo", "bar")

	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.Equal("YmFy", buffer.String())
}

func TestTemplateViewFuncBase64Decode(t *testing.T) {
	assert := assert.New(t)

	test := `{{ .Var "foo" | base64 | base64decode }}`
	temp := New().WithBody(test).WithVar("foo", "bar")

	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.Equal("bar", buffer.String())
}

func TestTemplateViewFuncSplit(t *testing.T) {
	assert := assert.New(t)

	test := `{{ .Var "foo" | split "," }}`
	temp := New().WithBody(test).WithVar("foo", "bar,baz,biz")
	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.Equal("[bar baz biz]", buffer.String())
}

func TestTemplateViewFuncFirst(t *testing.T) {
	assert := assert.New(t)

	test := `{{ .Var "foo" | split "," | first }}`
	temp := New().WithBody(test).WithVar("foo", "bar,baz,biz")
	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.Equal("bar", buffer.String())
}
func TestTemplateViewFuncAt(t *testing.T) {
	assert := assert.New(t)

	test := `{{ .Var "foo" | split "," | at 1 }}`
	temp := New().WithBody(test).WithVar("foo", "bar,baz,biz")
	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.Equal("baz", buffer.String())
}

func TestTemplateViewFuncLast(t *testing.T) {
	assert := assert.New(t)

	test := `{{ .Var "foo" | split "," | last }}`
	temp := New().WithBody(test).WithVar("foo", "bar,baz,biz")
	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.Equal("biz", buffer.String())
}

func TestTemplateViewFuncSlice(t *testing.T) {
	assert := assert.New(t)

	test := `{{ .Var "foo" | split "," | slice 1 3 }}`
	temp := New().WithBody(test).WithVar("foo", "bar,baz,biz,boof")
	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.Equal("[baz biz]", buffer.String())
}
func TestTemplateViewFuncJoin(t *testing.T) {
	assert := assert.New(t)

	test := `{{ .Var "foo" | split "," | slice 1 3 | join "/" }}`
	temp := New().WithBody(test).WithVar("foo", "bar,baz,biz,boof")
	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.Equal("baz/biz", buffer.String())
}

func TestTemplateViewFuncHasPrefix(t *testing.T) {
	assert := assert.New(t)

	test := `{{ if .Var "foo" | has_prefix "http" }}yep{{end}}`
	temp := New().WithBody(test).WithVar("foo", "http://foo.bar.com")
	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.Equal("yep", buffer.String())
}

func TestTemplateViewFuncHasSuffix(t *testing.T) {
	assert := assert.New(t)

	test := `{{ if .Var "foo" | has_suffix "com" }}yep{{end}}`
	temp := New().WithBody(test).WithVar("foo", "http://foo.bar.com")
	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.Equal("yep", buffer.String())
}

func TestTemplateViewFuncContains(t *testing.T) {
	assert := assert.New(t)

	test := `{{ if .Var "foo" | contains "bar" }}yep{{end}}`
	temp := New().WithBody(test).WithVar("foo", "http://foo.bar.com")
	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.Equal("yep", buffer.String())
}

func TestTemplateViewFuncMatches(t *testing.T) {
	assert := assert.New(t)

	test := `{{ if .Var "foo" | matches "^[a-z]+$" }}yep{{else}}nope{{end}}`
	temp := New().WithBody(test).WithVar("foo", "http://foo.bar.com")
	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.Equal("nope", buffer.String())
}

func TestTemplateViewFuncSha1(t *testing.T) {
	assert := assert.New(t)

	test := `{{ .Var "foo" | sha1 }}`
	temp := New().WithBody(test).WithVar("foo", "this is only a test")
	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.Equal("e7ee879d16c08f616c32e5bbe2253bdba18cf003", buffer.String())
}

func TestTemplateViewFuncSha256(t *testing.T) {
	assert := assert.New(t)

	test := `{{ .Var "foo" | sha256 }}`
	temp := New().WithBody(test).WithVar("foo", "this is only a test")
	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.Equal("1661186b8e38e79f434e4549a2d53f84716cfff7c45d334bbc67c9d41d1e3be6", buffer.String())
}

func TestTemplateViewFuncSha512(t *testing.T) {
	assert := assert.New(t)

	test := `{{ .Var "foo" | sha512 }}`
	temp := New().WithBody(test).WithVar("foo", "this is only a test")
	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.Equal("819bf8f4c3c5508a061d0637d09858cf098ef8ef7cafa312d07fca8480703eccf1a00b24b8915e24f926a8106331d7bc064e63c04262dbed65e05b28e208e53e", buffer.String())
}

func TestTemplateViewFuncSemverMajor(t *testing.T) {
	assert := assert.New(t)

	test := `{{ .Var "foo" | semver | major }}`
	temp := New().WithBody(test).WithVar("foo", "1.2.3-beta1")
	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.Equal("1", buffer.String())
}

func TestTemplateViewFuncSemverMinor(t *testing.T) {
	assert := assert.New(t)

	test := `{{ .Var "foo" | semver | minor }}`
	temp := New().WithBody(test).WithVar("foo", "1.2.3-beta1")
	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.Equal("2", buffer.String())
}

func TestTemplateViewFuncSemverPatch(t *testing.T) {
	assert := assert.New(t)

	test := `{{ .Var "foo" | semver | patch }}`
	temp := New().WithBody(test).WithVar("foo", "1.2.3-beta1")
	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.Equal("3", buffer.String())
}

func TestTemplateViewFuncSemverPreRelease(t *testing.T) {
	assert := assert.New(t)

	test := `{{ .Var "foo" | semver | prerelease }}`
	temp := New().WithBody(test).WithVar("foo", "1.2.3-beta1")
	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.Equal("beta1", buffer.String())
}

type label struct {
	Name string `yaml:"name"`
	Vaue string `yaml:"value"`
}

func TestTemplateViewFuncYAML(t *testing.T) {
	assert := assert.New(t)

	test := `
type: foo
meta: 
	name:
	labels:
{{ .Var "labels" | yaml | indent 1 }}
`
	temp := New().WithBody(test).WithVar("labels", []label{
		{"foo", "bar"},
		{"bar", "baz"},
		{"moobar", "zoobar"},
	})
	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.NotEmpty(buffer.String())
}

func TestTemplateWithVars(t *testing.T) {
	assert := assert.New(t)

	temp := New().WithVars(map[string]interface{}{
		"foo": "baz",
		"bar": "buz",
	})
	assert.True(temp.HasVar("foo"))
	assert.True(temp.HasVar("bar"))
	assert.False(temp.HasVar("baz"))
	assert.False(temp.HasVar("buz"))
	val, err := temp.Var("foo")
	assert.Nil(err)
	assert.Equal("baz", val)
	val, err = temp.Var("bar")
	assert.Nil(err)
	assert.Equal("buz", val)
}
