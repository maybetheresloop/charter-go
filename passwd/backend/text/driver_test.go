package text

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpenReader(t *testing.T) {
	text := "user1,passwd1\nuser2,passwd2\nuser3,passwd3\n"

	c, err := openReader(strings.NewReader(text))
	assert.Nil(t, err)

	tests := []struct {
		user string
		pass string
	}{
		{user: "user1", pass: "passwd1"},
		{user: "user2", pass: "passwd2"},
		{user: "user3", pass: "passwd3"},
	}

	for _, tt := range tests {
		assert.Contains(t, c, tt.user)
		pass, err := c.GetPassword(tt.user)
		assert.Nil(t, err)
		assert.Equal(t, tt.pass, pass)
	}
}

func TestOpenReaderMalformedRecord(t *testing.T) {
	text := "user1\n"

	c, err := openReader(strings.NewReader(text))
	assert.NotNil(t, err)
	assert.Nil(t, c)
}

func TestOpenReaderEmpty(t *testing.T) {
	text := ""

	c, err := openReader(strings.NewReader(text))
	assert.Nil(t, err)
	v := c.(connector)

	assert.Len(t, v, 0)
}
