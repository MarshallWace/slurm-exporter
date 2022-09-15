package ldapsearch

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseLDIfFile(t *testing.T) {
	s, err := Init("./testdata/data.ldif")
	assert.NoError(t, err)
	username := s.GetUsername("1932616621")
	assert.Equal(t, "aorsaria-adm", username)
}

func TestRunCmd(t *testing.T) {
	s, err := Init("")
	assert.NoError(t, err)
	username := s.GetUsername("1932616621")
	assert.Equal(t, "aorsaria-adm", username)
}
