package core

import (
	"io/ioutil"

	"encoding/json"

	"github.com/pkg/errors"
)

const (
	_user = "$user"
	_pass = "$pass"
)

type (
	character struct {
		Name      string  `json:"name"`
		Address   string  `json:"address"`
		Password  []byte  `json:"password"`
		AutoLogin chain   `json:"auto_login"`
		Aliases   aliases `json:"aliases"`
	}
)

func (e *entropy) loadCharacter(file string) (*character, error) {
	byt, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Wrap(err, "ioutil.ReadFile")
	}

	c := &character{}
	if err = json.Unmarshal(byt, c); err != nil {
		return nil, errors.Wrap(err, "json.Unmarshal")
	}

	return c, nil
}
