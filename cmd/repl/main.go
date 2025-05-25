package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/sealor/PlayzApp/internal/player"
)

func main() {
	mpv := player.Player{}
	if err := mpv.Start(); err != nil {
		log.Fatal(err)
	}

	eventCh := make(chan []byte, 16)
	mpv.SetEventChannel(eventCh)

	input := bufio.NewScanner(os.Stdin)
	for {
	loop:
		for {
			select {
			case event := <-eventCh:
				fmt.Print(string(event))
			default:
				break loop
			}
		}

		fmt.Print("> ")
		if !input.Scan() {
			break
		}

		cmd := input.Text()
		if cmd == "" {
			continue
		}

		var out map[string]any
		cmdFields := strings.Fields(cmd)
		errCh, err := mpv.Exec(&out, stringToAnySlice(cmdFields)...)
		if err != nil {
			log.Println(err)
			break
		}
		if err := <-errCh; err != nil {
			log.Println(err)
			break
		}

		fmt.Println(out)
	}

	if input.Err() != nil {
		log.Println(input.Err())
	}

	if err := mpv.Stop(); err != nil {
		log.Fatal(err)
	}
}

func stringToAnySlice(s []string) []any {
	r := make([]any, 0, 5)
	for _, i := range s {
		r = append(r, i)
	}
	return r
}
