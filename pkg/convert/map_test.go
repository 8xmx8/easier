package cvt

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGetKeys(t *testing.T) {
	type TCast struct {
		Name string
		Time int64
	}
	nano := time.Now().UnixNano()
	cast := TCast{
		Name: "test",
		Time: nano,
	}
	toMap, err := ToMap(cast)
	assert.NoError(t, err)
	keys := GetKeys(toMap)
	t.Log(toMap)
	assert.Equal(t, 2, len(keys))
	assert.Equal(t, "Name", keys[0])
	assert.Equal(t, "Time", keys[1])
}
