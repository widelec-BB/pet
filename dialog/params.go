package dialog

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jroimartin/gocui"
	"github.com/knqyf263/pet/config"
)

var (
	views      = []string{}
	layoutStep = 3
	curView    = -1
	idxView    = 0

	//CurrentCommand is the command before assigning to variables
	CurrentCommand string
	//FinalCommand is the command after assigning to variables
	FinalCommand string
)

func insertParams(command string, params map[string]string) string {
	resultCommand := command
	for k, v := range params {
		resultCommand = strings.Replace(resultCommand, k, v, -1)
	}
	return resultCommand
}

// SearchForParams returns variables from a command
func SearchForParams(lines []string) (map[string]string, map[string]string) {
	if len(lines) == 1 {
		normal := searchForParams(`<([\S].+?[\S])>`, lines[0], func(string) string {
			return ""
		})
		paths := searchForParams(`{{([\S].+?[\S])}}`, lines[0], func(name string) string {
			stdin := bytes.NewBuffer(nil)
			err := filepath.Walk("./", func(path string, info os.FileInfo, err error) error {
				stdin.WriteString(path)
				stdin.WriteRune('\n')
				return nil
			})
			if err != nil {
				return ""
			}
			var w bytes.Buffer
			cmd := exec.Command(config.Conf.General.SelectCmd, "--prompt="+name+" >")
			cmd.Stderr = os.Stderr
			cmd.Stdout = &w
			cmd.Stdin = stdin
			if err := cmd.Run(); err != nil {
				return ""
			}
			return "\"" + strings.TrimRight(w.String(), "\n\r ") + "\""
		})
		return normal, paths
	}
	return nil, nil
}

func searchForParams(regex, line string, def func(string) string) map[string]string {
	r, _ := regexp.Compile(regex)

	matches := r.FindAllStringSubmatch(line, -1)
	if len(matches) == 0 {
		return nil
	}

	extracted := make(map[string]string)
	for _, p := range matches {
		splitted := strings.Split(p[1], "=")
		if _, exists := extracted[p[0]]; exists == true {
			continue
		}
		if len(splitted) == 1 {
			extracted[p[0]] = def(p[0])
		} else {
			extracted[p[0]] = splitted[1]
		}
	}

	return extracted
}

func evaluateParams(g *gocui.Gui, _ *gocui.View) error {
	paramsFilled := map[string]string{}
	for _, v := range views {
		view, _ := g.View(v)
		res := view.Buffer()
		res = strings.Replace(res, "\n", "", -1)
		paramsFilled[v] = strings.TrimSpace(res)
	}
	FinalCommand = insertParams(CurrentCommand, paramsFilled)
	return gocui.ErrQuit
}
