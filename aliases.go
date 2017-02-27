package glas

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type (
	alias struct {
		Command string `toml:"cmd"`
		Action  string `toml:"action"`
	}

	aliases map[string]*alias
)

func (g *glas) maybeHandleAlias(input string) (bool, error) {
	g.aliasesMutex.Lock()
	defer g.aliasesMutex.Unlock()

	// TODO allow multi-field command matching.
	fields := strings.SplitN(input, " ", 2)
	// TODO allow naming to be separate from the match argument.
	match := fields[0]
	var args []string
	if len(fields) > 1 {
		// TODO replace fields with something that can account for quoted strings.
		args = strings.Fields(fields[1])
	}

	al, ok := g._aliases[match]
	if !ok {
		return false, nil
	}

	if len(args) != strings.Count(al.Command, "*") {
		return false, nil
	}

	actionLines := strings.Split(al.Action, "\n")
	var action chain
	for _, line := range actionLines {
		fields := strings.Fields(line)
		for i, f := range fields {
			quoted := false
			if strings.Contains(f, `"`) {
				f = strings.Trim(f, `"`)
				quoted = true
			}
			if strings.HasPrefix(f, "%") {
				f = strings.TrimSpace(strings.TrimPrefix(f, "%"))
				n, err := strconv.Atoi(f)
				if err != nil {
					return false, errors.Wrapf(err, "strconv.Atoi : %s", f)
				}

				n = n - 1
				if n > len(args) {
					return false, nil
				}
				fields[i] = args[n]
			}
			if quoted {
				fields[i] = fmt.Sprintf(`"%s"`, fields[i])
			}
		}

		line = strings.Join(fields, " ")
		action = append(action, line)
	}

	if err := g.handleChain(action); err != nil {
		return false, errors.Wrap(err, "handleChain")
	}

	return true, nil
}

// TODO make this support multi-line alias (may require some form of curses)
func (g *glas) newAlias(input string) {
	fields := strings.SplitN(input, " ", 2)
	if len(fields) != 2 {
		if al, ok := g._aliases[input]; ok {
			fmt.Println(al.Action)
		}
		return
	}

	match := fields[0]
	cmd := fields[1]

	g.aliasesMutex.Lock()
	defer g.aliasesMutex.Unlock()

	// Check if the alias exists and warn that it was overwritten if it did.
	warn := ""
	if al, ok := g._aliases[match]; ok {
		warn = fmt.Sprintf("%s", al.Action)
	}

	defer func(s string) {
		if s != "" {
			fmt.Printf("Warning: %s overwritten:%s\n", match, warn)
		}
	}(warn)

	g._aliases[match] = &alias{Action: cmd}
	fmt.Printf("%s set to %s\n", match, cmd)
}
