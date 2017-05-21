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

	temp, err := NewFromFile("testdata/test.template")
	assert.Nil(err)

	temp = temp.
		WithVar("service-name", "test-service").
		WithVar("app-name", "test-service-app").
		WithVar("container-name", "nginx").
		WithVar("container-image", "nginx:1.7.9")

	buffer := bytes.NewBuffer(nil)
	err = temp.Process(buffer)
	assert.Nil(err)

	result := buffer.String()
	assert.True(strings.Contains(result, "name: test-service"))
	assert.True(strings.Contains(result, "replicas: 2"))
	assert.True(strings.Contains(result, "app: test-service-app"))
	assert.False(strings.Contains(result, "ports:"))

	temp = temp.WithVar("container-port", 80)
	err = temp.Process(buffer)
	assert.Nil(err)
	result = buffer.String()
	assert.True(strings.Contains(result, "containerPort: 80"))
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

func TestTemplateFile(t *testing.T) {
	assert := assert.New(t)

	test := `{{ .File "testdata/inline_file" }}`
	temp := New().WithBody(test)

	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.Equal("this is a test", buffer.String())
}

func TestTemplateViewFuncs(t *testing.T) {
	assert := assert.New(t)

	test := `{{ .Var "now" | time_unix }}`
	temp := New().WithBody(test).WithVar("now", time.Date(2017, 05, 20, 21, 00, 00, 00, time.UTC))

	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.Equal("1495314000", buffer.String())
}

func TestTemplateHelpersUTCNow(t *testing.T) {
	assert := assert.New(t)

	test := `{{ .Helpers.UTCNow | time_unix }}`
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
