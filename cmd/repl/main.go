package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/sealor/PlayzApp/internal/player"
)

func main() {
	mpv := player.Player{}
	if err := mpv.Start(); err != nil {
		log.Fatal(err)
	}

	input := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !input.Scan() {
			break
		}
		cmd := input.Text()
		out, err := mpv.Exec(cmd)
		fmt.Println(out)
		if err != nil {
			log.Println(err)
			break
		}
	}

	if input.Err() != nil {
		log.Println(input.Err())
	}

	if err := mpv.Stop(); err != nil {
		log.Fatal(err)
	}
}
