package glas

import (
	"io/ioutil"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

const (
	_user = "$user"
	_pass = "$pass"
)

type (
	// Conf base toml conf
	conf struct {
		Character characterConf
		Connect   connectConf
		Aliases   aliases

		// Unexported, instance only settings.
		filePath string
	}
	// CharacterConf character related sub-sections from toml conf
	characterConf struct {
		Name string
		//Password []byte
		Password string // toml-compat... []byte was bad enough but this is terribad...
	}
	// ConnectConf connection related sub-sections from toml conf
	connectConf struct {
		Address   string
		AutoLogin chain `toml:"auto_login"`
	}
)

func (g *glas) loadConf(file string) (*conf, error) {
	byt, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Wrapf(err, "ioutil.ReadFile")
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

/*
func (g *glas) saveConf(c *conf) error {
	byt, err := json.Marshal(c)
	if err != nil {
		return errors.Wrap(err, "json.Marshal")
	}

	// Only owner can read/write conf filg.
	if err = ioutil.WriteFile(c.filePath, byt, 600); err != nil {
		return errors.Wrap(err, "iotuil.WriteFile")
	}

	return nil
}
*/
