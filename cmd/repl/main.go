package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/sealor/PlayzApp/internal/player"
	"golang.org/x/term"
)

func main() {
	fmt.Println("Commands https://mpv.io/manual/stable/#list-of-input-commands")
	fmt.Println("- loadfile https://www.sample-videos.com/video321/mp4/240/big_buck_bunny_240p_1mb.mp4")
	fmt.Println("- seek +15")
	fmt.Println("- get_property volume")
	fmt.Println("- set_property volume 95")
	fmt.Println("- stop")

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
	allPropertyNames, _ := mpv.GetPropertyNames()
	t.AutoCompleteCallback = func(line string, pos int, key rune) (newLine string, newPos int, ok bool) {
		return autoCompleteCallback(t, allPropertyNames, line, key)
	}

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

func autoCompleteCallback(t *term.Terminal, allPropertyNames []string, line string, key rune) (newLine string, newPos int, ok bool) {
	if key != '\t' {
		return "", 0, false
	}

	matchingPropertyNames := []string{}
	for _, properyName := range allPropertyNames {
		if properyName == line {
			matchingPropertyNames = []string{properyName}
			break
		}
		if strings.HasPrefix(properyName, line) {
			matchingPropertyNames = append(matchingPropertyNames, properyName)
		}
	}

	if len(matchingPropertyNames) == 1 {
		newLine = "get_property " + matchingPropertyNames[0]
		return newLine, len(newLine), true
	}

	if len(matchingPropertyNames) > 1 {
		fmt.Fprintln(t, strings.Join(matchingPropertyNames, ", "))
		commonPrefix := findLongestCommonPrefix(matchingPropertyNames)
		return commonPrefix, len(commonPrefix), true
	}

	return "", 0, false
}

func stringToAnySlice(s []string) []any {
	r := make([]any, 0, 5)
	for _, i := range s {
		r = append(r, i)
	}
	return r
}

func findLongestCommonPrefix(strs []string) string {
	if len(strs) == 0 {
		return ""
	}

	longestPrefix := ""

	sort.Strings(strs)
	first := string(strs[0])
	last := string(strs[len(strs)-1])

	for i := 0; i < len(first); i++ {
		if last[i] == first[i] {
			longestPrefix += string(last[i])
		} else {
			return longestPrefix
		}
	}

	return longestPrefix
}
