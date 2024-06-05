package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMarshal(t *testing.T) {
	type User struct {
		ID   uint64
		Name string
		Age  uint8
		Tag  string
	}
	u := User{
		ID:   1,
		Name: "test",
		Age:  18,
		Tag:  "test",
	}
	u1 := new(User)
	b, err := Marshal(&u)
	assert.NoError(t, err)
	t.Log(b)
	err = Unmarshal(b, &u1)
	assert.NoError(t, err)
	t.Log(u1)
	assert.Equal(t, u, *u1)
}
