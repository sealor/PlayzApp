package test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalPartOfJsonOnly(t *testing.T) {
	type Pair struct {
		Key string `json:"key"`
	}

	var pair Pair
	text := []byte(`{"key": "value", "asd": "asd"}`)
	err := json.Unmarshal(text, &pair)
	require.NoError(t, err)

	assert.Equal(t, "value", pair.Key)
}

func TestAnyChannel(t *testing.T) {
	ch := make(chan any, 3)
	ch <- 7
	ch <- "nine"
	ch <- struct{}{}
	assert.Equal(t, 7, <-ch)
	assert.Equal(t, "nine", <-ch)
	assert.Equal(t, struct{}{}, <-ch)
}

func TestNilChannel(t *testing.T) {
	var ch chan any
	select {
	case ch <- 7:
		assert.FailNow(t, "insert into nil channel blocks forever")
	default:
	}
}

func TestBufferedChannel(t *testing.T) {
	ch := make(chan struct{}, 1)
	select {
	case ch <- struct{}{}:
	default:
		assert.FailNow(t, "buffered channel allows insert")
	}
	select {
	case ch <- struct{}{}:
		assert.FailNow(t, "full buffered channel does not allow insert")
	default:
	}
}

func TestReflectChannel(t *testing.T) {
	ch := make(chan float64)
	chType := reflect.TypeOf(ch)
	assert.Equal(t, reflect.Float64, chType.Elem().Kind())
}

func TestChannelSync(t *testing.T) {
	ch := make(chan struct{})
	var out int
	go func(out *int) {
		*out = 7
		ch <- struct{}{}
	}(&out)
	<-ch
	assert.Equal(t, 7, out)
}

func TestStringsSplit(t *testing.T) {
	splits := strings.Split("hello   world", " ")
	assert.Equal(t, []string{"hello", "", "", "world"}, splits)
}

func TestStringsFields(t *testing.T) {
	splits := strings.Fields("hello   world")
	assert.Equal(t, []string{"hello", "world"}, splits)
}
