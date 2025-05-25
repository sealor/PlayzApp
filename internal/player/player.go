package player

import (
	"encoding/json"
	"errors"
	"log"
	"sync"
)

type Player struct {
	ipc         MpvIpcReadWriter
	mutex       sync.Mutex
	responseMap map[int]ResponseResult
	requestId   int
	eventCh     chan []byte
}

type Command struct {
	Command   []any `json:"command"`
	RequestId int   `json:"request_id"`
}

type ResponseBase struct {
	RequestId int `json:"request_id"`
}

type ResponseResult struct {
	Value any
	ErrCh chan error
}

var (
	ErrAlreadyStarted = errors.New("mpv was already started")
	ErrNotStarted     = errors.New("mpv was not started")
)

func (p *Player) Start() error {
	if p.ipc != nil {
		return ErrAlreadyStarted
	}
	ipc, err := Open()
	if err != nil {
		return err
	}

	p.ipc = ipc
	p.responseMap = make(map[int]ResponseResult)

	go func() {
		for {
			var base ResponseBase
			response, err := p.ipc.ReadResponse()
			if err != nil {
				log.Println(err)
				break
			}
			err = json.Unmarshal(response, &base)
			if err != nil {
				log.Println(err)
				break
			}

			// TODO: handle events with map too
			// TODO: return registered event type or map
			p.mutex.Lock()
			result, ok := p.responseMap[base.RequestId]

			if ok {
				result.ErrCh <- json.Unmarshal(response, result.Value)
				delete(p.responseMap, base.RequestId)
			} else {
				select {
				case p.eventCh <- response:
				default:
				}
			}
			p.mutex.Unlock()
		}
	}()

	return nil
}

func (p *Player) SetEventChannel(eventCh chan []byte) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.eventCh = eventCh
}

func (p *Player) Exec(out any, cmd ...any) (chan error, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.requestId++
	requestId := p.requestId
	command := Command{cmd, requestId}
	encoder := json.NewEncoder(p.ipc)
	if err := encoder.Encode(command); err != nil {
		return nil, err
	}

	errCh := make(chan error)
	p.responseMap[requestId] = ResponseResult{out, errCh}
	return errCh, nil
}

func (p *Player) Stop() error {
	if p.ipc == nil {
		return ErrNotStarted
	}

	if err := p.ipc.WriteRequest([]byte("quit")); err != nil {
		return err
	}

	if err := p.ipc.Close(); err != nil {
		return err
	}

	p.ipc = nil
	return nil
}
