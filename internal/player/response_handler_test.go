package player

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntDataRequestHandling(t *testing.T) {
	r := NewResponseHandler(16)
	var val int

	requestID, errCh := r.AddRequest(&val)
	err := r.HandleResponse(fmt.Appendf(nil, `{"request_id": %d, "data": 23}`, requestID))
	require.NoError(t, err)

	<-errCh
	assert.Equal(t, 23, val)
}

func TestErrorOnRequestHandling(t *testing.T) {
	r := NewResponseHandler(16)
	var val int

	requestID, errCh := r.AddRequest(&val)
	err := r.HandleResponse(fmt.Appendf(nil, `{"request_id": %d, "error": "invalid parameter"}`, requestID))
	require.NoError(t, err)

	assert.Equal(t, errors.New("invalid parameter"), <-errCh)
}

func TestDefaultEventHandling(t *testing.T) {
	r := NewResponseHandler(16)

	err := r.HandleResponse([]byte(`{"event": "idle"}`))
	require.NoError(t, err)

	select {
	case val := <-r.GetEventChannel():
		expectedVal := make(map[string]any)
		expectedVal["event"] = "idle"
		assert.Equal(t, expectedVal, val)
	default:
		t.Fatal("event expected")
	}
}

func TestNewFuncEventHandling(t *testing.T) {
	type HelloEvent struct {
		Event    string `json:"event"`
		Greeting string `json:"greeting"`
	}

	r := NewResponseHandler(16)

	r.RegisterNewEventFunc("hello", func() any { return &HelloEvent{} })
	err := r.HandleResponse([]byte(`{"event": "hello", "greeting": "Hello World!"}`))
	require.NoError(t, err)

	select {
	case val := <-r.GetEventChannel():
		expectedVal := &HelloEvent{"hello", "Hello World!"}
		assert.Equal(t, expectedVal, val)
	default:
		t.Fatal("hello event expected")
	}
}
