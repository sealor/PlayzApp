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
	defer func() {
		_ = term.Restore(int(os.Stdin.Fd()), oldState)
	}()

	t := term.NewTerminal(os.Stdin, "> ")
	allPropertyNames, _ := mpv.GetPropertyNames()
	sort.Strings(allPropertyNames)
	allCommandNames, _ := mpv.GetCommandNames()
	allCommandNames = append(allCommandNames, "get_property")
	allCommandNames = append(allCommandNames, "set_property")
	sort.Strings(allCommandNames)

	t.AutoCompleteCallback = func(line string, pos int, key rune) (newLine string, newPos int, ok bool) {
		return autoCompleteCallback(t, allCommandNames, allPropertyNames, line, key)
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

	if err := mpv.Stop(); err != nil {
		log.Println(err)
	}
}

func autoCompleteCallback(t *term.Terminal, allCommandNames, allPropertyNames []string, line string, key rune) (newLine string, newPos int, ok bool) {
	if key != '\t' {
		return "", 0, false
	}

	words := strings.Fields(line)

	if len(words) == 0 {
		fmt.Fprintln(t, strings.Join(allCommandNames, ", "))
		return "", 0, false
	}

	if len(words) == 1 && (words[0] == "get_property" || words[0] == "set_property") {
		words = append(words, "")
	}

	if len(words) == 1 {
		prefixMatches := findAllPrefixMatches(words[0], allCommandNames)

		if len(prefixMatches) == 0 {
			return "", 0, false
		}

		longestCommonPrefix := findLongestCommonPrefix(prefixMatches)
		completion := longestCommonPrefix

		if len(prefixMatches) == 1 {
			completion = completion + " "
		} else {
			fmt.Fprintln(t, strings.Join(prefixMatches, ", "))
		}

		return completion, len(completion), true
	}

	if len(words) == 2 && (words[0] == "get_property" || words[0] == "set_property") {
		prefixMatches := findAllPrefixMatches(words[1], allPropertyNames)

		if len(prefixMatches) == 0 {
			return "", 0, false
		}

		longestCommonPrefix := findLongestCommonPrefix(prefixMatches)
		completion := words[0] + " " + longestCommonPrefix

		if len(prefixMatches) == 1 {
			completion = completion + " "
		} else {
			fmt.Fprintln(t, strings.Join(prefixMatches, ", "))
		}

		return completion, len(completion), true
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

func findAllPrefixMatches(prefix string, words []string) []string {
	result := []string{}
	for _, word := range words {
		if word == prefix {
			result = []string{word}
			break
		}
		if strings.HasPrefix(word, prefix) {
			result = append(result, word)
		}
	}
	return result
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
