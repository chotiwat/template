package template

import (
	"bytes"
	"testing"
	"time"

	assert "github.com/blendlabs/go-assert"
)

func TestTemplate(t *testing.T) {
	assert := assert.New(t)

	test := `{{ .Var "foo" }}`
	temp := New().WithBody(test).WithVar("foo", "bar")

	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.Equal("bar", buffer.String())
}

func TestTemplateHelpers(t *testing.T) {
	assert := assert.New(t)

	test := `{{ .Var "now" | unix }}`
	temp := New().WithBody(test).WithVar("now", time.Date(2017, 05, 20, 21, 00, 00, 00, time.UTC))

	buffer := bytes.NewBuffer(nil)
	err := temp.Process(buffer)
	assert.Nil(err)
	assert.Equal("1495314000", buffer.String())
}
