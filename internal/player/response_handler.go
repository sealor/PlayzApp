package player

import (
	"encoding/json"
	"errors"
	"sync"
)

type ResponseHandler struct {
	mutex           sync.Mutex
	lastRequestID   int
	requestMap      map[int]ResponseResult
	newEventFuncMap map[string]func() any
	eventCh         chan any
}

type ResponseBase struct {
	RequestID int             `json:"request_id"`
	Event     string          `json:"event"`
	Data      json.RawMessage `json:"data"`
	Error     string          `json:"error"`
}

type ResponseResult struct {
	ValueRef any
	ErrCh    chan error
}

var (
	ErrUnknownResponse = errors.New("unknown response")
	ErrUnknownRequest  = errors.New("unknown request")
)

func NewResponseHandler(eventCacheSize int) *ResponseHandler {
	return &ResponseHandler{
		sync.Mutex{}, 0, make(map[int]ResponseResult), make(map[string]func() any), make(chan any, eventCacheSize),
	}
}

func (r *ResponseHandler) AddRequest(outData any) (int, <-chan error) {
	errCh := make(chan error, 1)
	r.mutex.Lock()
	r.lastRequestID++
	r.requestMap[r.lastRequestID] = ResponseResult{outData, errCh}
	r.mutex.Unlock()
	return r.lastRequestID, errCh
}

func (r *ResponseHandler) HandleResponse(response []byte) error {
	var base ResponseBase
	if err := json.Unmarshal(response, &base); err != nil {
		return err
	}

	if base.RequestID == 0 && base.Event == "" {
		return ErrUnknownResponse
	}

	if base.RequestID > 0 {
		if base.Error == "success" {
			return r.handleData(base.RequestID, base.Data)
		} else {
			return r.handleError(base.RequestID, base.Error)
		}
	} else {
		return r.handleEvent(base.Event, response)
	}
}

func (r *ResponseHandler) handleData(requestID int, data []byte) error {
	r.mutex.Lock()
	resResult, ok := r.requestMap[requestID]
	delete(r.requestMap, requestID)
	r.mutex.Unlock()

	if !ok {
		return ErrUnknownRequest
	}

	if len(data) > 0 {
		err := json.Unmarshal(data, resResult.ValueRef)
		resResult.ErrCh <- err
		return err
	} else {
		resResult.ErrCh <- nil
		return nil
	}
}

func (r *ResponseHandler) handleError(requestID int, error string) error {
	r.mutex.Lock()
	resResult, ok := r.requestMap[requestID]
	delete(r.requestMap, requestID)
	r.mutex.Unlock()

	if !ok {
		return ErrUnknownRequest
	}

	resResult.ErrCh <- errors.New(error)
	return nil
}

func (r *ResponseHandler) handleEvent(eventName string, response []byte) error {
	r.mutex.Lock()
	newFunc, ok := r.newEventFuncMap[eventName]
	delete(r.newEventFuncMap, eventName)
	r.mutex.Unlock()

	var event any
	if ok {
		event = newFunc()
	} else {
		event = make(map[string]any)
	}

	if err := json.Unmarshal(response, &event); err != nil {
		return err
	}

	select {
	case r.eventCh <- event:
	default:
	}

	return nil
}

func (r *ResponseHandler) RegisterNewEventFunc(eventName string, newFunc func() any) {
	r.mutex.Lock()
	r.newEventFuncMap[eventName] = newFunc
	r.mutex.Unlock()
}

func (r *ResponseHandler) GetEventChannel() <-chan any {
	return r.eventCh
}
