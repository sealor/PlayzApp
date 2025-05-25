package player

import (
	"bufio"
	"io"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"
)

type MpvIpcReadWriter interface {
	io.Reader
	io.Writer
	io.Closer

	ReadResponse() ([]byte, error)
	WriteRequest([]byte) error
}

type MpvIpc struct {
	cmd          *exec.Cmd
	cmdIO        *os.File
	cmdBufReader *bufio.Reader
}

var _ MpvIpcReadWriter = &MpvIpc{}

func Open() (MpvIpcReadWriter, error) {
	fds, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil {
		return nil, err
	}

	parentConn := os.NewFile(uintptr(fds[0]), "parentConn")
	childConn := os.NewFile(uintptr(fds[1]), "childConn")
	defer childConn.Close()

	cmd := exec.Command("mpv", "--input-ipc-client=fd://3", "--idle", "--no-terminal")
	cmd.ExtraFiles = []*os.File{childConn}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &MpvIpc{cmd, parentConn, bufio.NewReader(parentConn)}, nil
}

func (m *MpvIpc) Close() error {
	if m.cmd == nil {
		return ErrNotStarted
	}

	if _, err := m.cmdIO.Write([]byte("quit\n")); err != nil {
		return err
	}
	if err := m.cmdIO.Close(); err != nil {
		return nil
	}

	quitted := make(chan error)
	go func() { quitted <- m.cmd.Wait() }()

	select {
	case err := <-quitted:
		return err
	case <-time.After(time.Second):
		log.Println("quitting mpv failed")
		log.Println("kill mpv")
		if err := m.cmd.Process.Kill(); err != nil {
			return err
		}
	}

	m.cmd = nil
	return nil
}

func (m *MpvIpc) Read(b []byte) (int, error) {
	return m.cmdBufReader.Read(b)
}

func (m *MpvIpc) ReadResponse() ([]byte, error) {
	return m.cmdBufReader.ReadBytes('\n')
}

func (m *MpvIpc) Write(b []byte) (int, error) {
	return m.cmdIO.Write(b)
}

func (m *MpvIpc) WriteRequest(b []byte) error {
	_, err := m.Write(append(b, '\n'))
	return err
}
