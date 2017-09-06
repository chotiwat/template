package template

import (
	"testing"

	assert "github.com/blendlabs/go-assert"
)

func TestSemver(t *testing.T) {
	assert := assert.New(t)

	sv, err := NewSemver("1.2.3-beta1")
	assert.Nil(err)
	assert.Equal(1, sv.Major)
	assert.Equal(2, sv.Minor)
	assert.Equal(3, sv.Patch)
	assert.Equal("beta1", sv.PreRelease)
}
