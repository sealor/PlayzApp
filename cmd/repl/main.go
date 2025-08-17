package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/sealor/PlayzApp/internal/player"
	"golang.org/x/term"
)

func main() {
	mpv := player.Player{}
	if err := mpv.Start(); err != nil {
		log.Fatal(err)
	}

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatal(err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	t := term.NewTerminal(os.Stdin, "> ")

	go func() {
		for event := range mpv.GetEventChannel() {
			fmt.Fprintln(t, "event:", event)
		}
	}()

	for {
		cmd, err := t.ReadLine()
		if err != nil {
			if err != io.EOF {
				fmt.Fprintln(t, "Fatal:", err)
			}
			break
		}

		if cmd == "" {
			continue
		}

		var out any
		cmdFields := strings.Fields(cmd)
		errCh, err := mpv.Exec(&out, stringToAnySlice(cmdFields)...)
		if err != nil {
			fmt.Fprintln(t, "Fatal:", err)
			break
		}
		if err := <-errCh; err != nil {
			fmt.Fprintln(t, "Error:", err)
		} else {
			fmt.Fprintln(t, out)
		}
	}

	term.Restore(int(os.Stdin.Fd()), oldState)

	if err := mpv.Stop(); err != nil {
		log.Println(err)
	}
}

func stringToAnySlice(s []string) []any {
	r := make([]any, 0, 5)
	for _, i := range s {
		r = append(r, i)
	}
	return r
}
