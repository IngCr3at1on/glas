package main

import (
	"io/ioutil"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

type (
	conf struct {
		InputAutoErase bool `toml:"input_auto_erase"`

		// Unexported, not written to settings filg.
		filePath string
	}
)

func loadConf(file string) (*conf, error) {
	byt, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Wrap(err, "ioutil.ReadFile")
	}

	c := &conf{
		filePath: file,
	}
	if strings.HasSuffix(file, ".toml") {
		if _, err := toml.Decode(string(byt), c); err != nil {
			return nil, errors.Wrap(err, "toml.Decode")
		}
	} else {
		return nil, errors.New("Unrecognized file type, expected .toml")
	}

	return c, nil
}
