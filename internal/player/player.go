package player

import (
	"encoding/json"
	"errors"
	"net"
	"os"
	"os/exec"
	"time"
)

type Player struct {
	cmd       *exec.Cmd
	conn      net.Conn
	requestId int
}

var (
	ErrAlreadyStarted = errors.New("mpv was already started")
	ErrNotStarted     = errors.New("mpv was not started")
)

type Command struct {
	Command   []any `json:"command"`
	RequestId int   `json:"request_id"`
}

func (p *Player) Start() error {
	if p.cmd != nil {
		return ErrAlreadyStarted
	}

	os.Remove("/tmp/mpvsocket")
	p.cmd = exec.Command("mpv", "--input-ipc-server=/tmp/mpvsocket", "--idle", "--no-terminal")

	p.cmd.Stdout = os.Stdout
	p.cmd.Stderr = os.Stderr

	if err := p.cmd.Start(); err != nil {
		return err
	}

	// TODO: replace sleep with polling
	time.Sleep(400 * time.Millisecond)

	conn, err := net.Dial("unix", "/tmp/mpvsocket")
	if err != nil {
		return err
	}
	p.conn = conn

	return nil
}

func (p *Player) Exec(cmd ...any) (map[string]any, error) {
	command := Command{Command: cmd, RequestId: p.requestId}
	p.requestId++

	input, err := json.Marshal(command)
	if err != nil {
		return nil, err
	}

	if _, err = p.conn.Write(append(input, '\n')); err != nil {
		return nil, nil
	}

	output := make(map[string]any)
	decoder := json.NewDecoder(p.conn)
	if err := decoder.Decode(&output); err != nil {
		return nil, err
	}

	return output, nil
}

func (p *Player) Stop() error {
	if p.cmd == nil {
		return ErrNotStarted
	}

	p.conn.Close()

	if err := p.cmd.Process.Kill(); err != nil {
		return err
	}

	if err := p.cmd.Wait(); err != nil {
		return err
	}

	return nil
}
