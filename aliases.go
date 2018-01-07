package glas

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type (
	// Alias is a command alias
	Alias struct {
		// Match is the alias match.
		Match string
		// Action is the action performed on a match.
		Action []string
	}

	aliases struct {
		sync.RWMutex

		m map[string]*Alias
	}
)

func (a *aliases) maybeHandleAlias(g *Glas, input string) (bool, error) {
	if input == "" {
		return false, nil
	}

	a.RLock()
	defer a.RUnlock()

	// TODO: allow multi-field command matching.
	fields := strings.SplitN(input, " ", 2)
	// TODO: allow naming to be separate from the match argument.
	match := fields[0]
	var args []string
	if len(fields) > 1 {
		// TODO: replace fields with something that can account for quoted strings.
		args = strings.Fields(fields[1])
	}

	var al *Alias
	for _, _al := range a.m {
		if strings.Contains(_al.Match, "*") && strings.HasPrefix(_al.Match, fmt.Sprintf("%s ", match)) {
			al = _al
		}
		if _al.Match == match {
			al = _al
		}
	}

	if al == nil {
		return false, nil
	}

	g.log.WithFields(logrus.Fields{
		"match":  al.Match,
		"action": al.Action,
	}).Debug("Matched aliase")

	if len(args) != strings.Count(al.Match, "*") {
		return false, nil
	}

	var action []interface{}
	for _, line := range al.Action {
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

	if err := g.Send(action...); err != nil {
		return false, errors.Wrap(err, "Send")
	}

	return true, nil
}

// TODO: rewrite and utilize this...
// TODO: make this support multi-line alias (may require some form of curses)
// func (g *Glas) newAlias(input string) {
// 	fields := strings.SplitN(input, " ", 2)
// 	if len(fields) != 2 {
// 		if al, ok := g._aliases[input]; ok {
// 			fmt.Println(al.Action)
// 		}
// 		return
// 	}

// 	match := fields[0]
// 	cmd := fields[1]

// 	g.aliasesMutex.Lock()
// 	defer g.aliasesMutex.Unlock()

// 	// Check if the alias exists and warn that it was overwritten if it did.
// 	warn := ""
// 	if al, ok := g._aliases[match]; ok {
// 		warn = fmt.Sprintf("%s", al.Action)
// 	}

// 	defer func(s string) {
// 		if s != "" {
// 			fmt.Printf("Warning: %s overwritten:%s\n", match, warn)
// 		}
// 	}(warn)

// 	g._aliases[match] = &alias{Action: cmd}
// 	fmt.Printf("%s set to %s\n", match, cmd)
// }
