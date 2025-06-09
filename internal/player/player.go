package player

import (
	"encoding/json"
	"errors"
	"log"
)

type Player struct {
	ipc        MpvIpcReadWriter
	resHandler *ResponseHandler
}

type Command struct {
	Command   []any `json:"command"`
	RequestId int   `json:"request_id"`
}

var (
	ErrAlreadyStarted = errors.New("mpv was already started")
	ErrNotStarted     = errors.New("mpv was not started")
)

func (p *Player) NewPlayer(resHandler *ResponseHandler) *Player {
	return &Player{nil, resHandler}
}

func (p *Player) Start() error {
	if p.ipc != nil {
		return ErrAlreadyStarted
	}
	ipc, err := Open()
	if err != nil {
		return err
	}

	p.ipc = ipc

	if p.resHandler == nil {
		p.resHandler = NewResponseHandler(16)
	}

	go func() {
		for {
			response, err := p.ipc.ReadResponse()
			if err != nil {
				log.Println(err)
				log.Println(string(response))
				break
			}
			if err = p.resHandler.HandleResponse(response); err != nil {
				log.Println(err)
				log.Println(string(response))
				break
			}
		}
	}()

	return nil
}

func (p *Player) GetEventChannel() <-chan any {
	return p.resHandler.GetEventChannel()
}

func (p *Player) Exec(out any, cmd ...any) (<-chan error, error) {
	requestId, errCh := p.resHandler.AddRequest(out)
	command := Command{cmd, requestId}
	encoder := json.NewEncoder(p.ipc)
	if err := encoder.Encode(command); err != nil {
		return nil, err
	}

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
