package main

// [27 91 51 49 109 206 187 27 91 48 109 32

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/tiago4orion/nash"
)

var completer = readline.NewPrefixCompleter(
	readline.PcItem("mode",
		readline.PcItem("vi"),
		readline.PcItem("emacs"),
	),
	readline.PcItem("rfork",
		readline.PcItem("c"),
		readline.PcItem("upmnis"),
		readline.PcItem("upmis"),
	),
	readline.PcItem("prompt="),
	readline.PcItem("path="),
)

func cli(sh *nash.Shell) error {
	var (
		err  error
		home string
	)

	home = os.Getenv("HOME")

	if home == "" {
		user := os.Getenv("USER")

		if user != "" {
			home = "/home/" + user
		} else {
			home = "/tmp"
		}
	}

	historyFile := home + "/.nash"

	os.Mkdir(historyFile, 0755)

	l, err := readline.NewEx(&readline.Config{
		Prompt:          sh.Prompt(),
		HistoryFile:     historyFile,
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})

	if err != nil {
		panic(err)
	}

	defer l.Close()

	log.SetOutput(l.Stderr())

	var content bytes.Buffer
	var lineidx int
	var line string

	for {
		line, err = l.Readline()

		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			err = nil
			break
		}

		lineidx++

		line = strings.TrimSpace(line)

		// handle special cli commands

		switch {
		case strings.HasPrefix(line, "set mode "):
			switch line[8:] {
			case "vi":
				l.SetVimMode(true)
			case "emacs":
				l.SetVimMode(false)
			default:
				fmt.Printf("invalid mode: %s\n", line[8:])
			}

			continue
		case line == "mode":
			if l.IsVimMode() {
				fmt.Printf("Current mode: vim\n")
			} else {
				fmt.Printf("Current mode: emacs\n")
			}

			continue

		case line == "exit":
			break
		}

		content.Write([]byte(line + "\n"))

		parser := nash.NewParser(fmt.Sprintf("line %d", lineidx), string(content.Bytes()))

		tr, err := parser.Parse()

		if err != nil {
			if err.Error() == "Open '{' not closed" {
				l.SetPrompt(">>> ")
				continue
			}

			fmt.Printf("ERROR: %s\n", err.Error())
			content.Reset()
			l.SetPrompt(sh.Prompt())
			continue
		}

		content.Reset()

		err = sh.ExecuteTree(tr)

		if err != nil {
			fmt.Printf("ERROR: %s\n", err.Error())
		}

		l.SetPrompt(sh.Prompt())
	}

	return err
}
