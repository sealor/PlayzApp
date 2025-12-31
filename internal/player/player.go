// Package player provides an API for mpv.
package player

import (
	"encoding/json"
	"errors"
	"io"
	"log"
)

type Player struct {
	ipc        MpvIpcReadWriter
	resHandler *ResponseHandler
}

type Command struct {
	Command   []any `json:"command"`
	RequestID int   `json:"request_id"`
}

var (
	ErrAlreadyStarted = errors.New("mpv was already started")
	ErrNotStarted     = errors.New("mpv was not started")
)

func NewPlayer(resHandler *ResponseHandler) *Player {
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
				if err != io.EOF {
					log.Println(err)
					log.Println(string(response))
				}
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
	requestID, errCh := p.resHandler.AddRequest(out)
	command := Command{cmd, requestID}
	encoder := json.NewEncoder(p.ipc)
	if err := encoder.Encode(command); err != nil {
		return nil, err
	}

	return errCh, nil
}

func (p *Player) GetCommandNames() ([]string, error) {
	commandDefinitions := make([]map[string]any, 0)
	errCh, err := p.Exec(&commandDefinitions, "get_property", "command-list")
	if err != nil {
		return nil, err
	}
	err = <-errCh
	if err != nil {
		return nil, err
	}

	commandNames := make([]string, 0)
	for _, commandDefinition := range commandDefinitions {
		commandNames = append(commandNames, commandDefinition["name"].(string))
	}
	return commandNames, nil
}

func (p *Player) GetPropertyNames() ([]string, error) {
	propertyNames := []string{}
	errCh, err := p.Exec(&propertyNames, "get_property", "property-list")
	if err != nil {
		return nil, err
	}
	err = <-errCh
	if err != nil {
		return nil, err
	}
	return propertyNames, nil
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
