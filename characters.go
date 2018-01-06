package glas

import (
	"fmt"
	"io/ioutil"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

type (
	// CharacterConfig is a character configuration.
	CharacterConfig struct {
		// Name is the character name.
		Name string
		// Password is the character password. TODO: encrypt this? bare minimum, hash it.
		Password string
		// Address is the mud address
		Address string
		// AutoLogin is an optional auto-login sequence.
		AutoLogin []string `toml:"auto_login"`

		aliases *aliases
		// Aliases are a characters aliases.
		Aliases map[string]*Alias `toml:"aliases"`
	}

	characters struct {
		sync.RWMutex
		m map[string]*CharacterConfig
	}
)

func (g *Glas) loadCharacterConfigs() error {
	if g.config.CharacterPath == "" {
		return nil
	}

	files, err := ioutil.ReadDir(g.config.CharacterPath)
	if err != nil {
		return err
	}

	for _, f := range files {
		if name := f.Name(); strings.HasSuffix(name, ".toml") && strings.Compare(name, "config.toml") != 0 {
			if _, err := g.loadCharacterConfig(fmt.Sprintf("%s/%s", g.config.CharacterPath, name)); err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *Glas) loadCharacterConfig(file string) (string, error) {
	byt, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}

	var config CharacterConfig
	if _, err := toml.Decode(string(byt), &config); err != nil {
		return "", errors.Wrap(err, "toml.Decode")
	}

	if err := config.Validate(); err != nil {
		return "", errors.Wrap(err, "Validate")
	}

	g.characters.addCharacter(&config)
	return config.Name, nil
}

// Validate a configuration, error if a required value is missing and set
// defaults (if a value is not provided) when not required.
func (c *CharacterConfig) Validate() error {
	if c.Address == "" {
		return errors.New("address cannot be empty")
	}

	c.aliases = &aliases{}
	c.aliases.Lock()
	defer c.aliases.Unlock()

	for k, v := range c.Aliases {
		if c.aliases.m == nil {
			c.aliases.m = make(map[string]*Alias)
		}
		c.aliases.m[k] = v
	}

	return nil
}

func (c *characters) addCharacter(cc *CharacterConfig) {
	c.Lock()
	defer c.Unlock()

	_, ok := c.m[cc.Name]
	if !ok {
		c.m[cc.Name] = cc
	}
}

func (c *characters) getCharacter(name string) *CharacterConfig {
	c.RLock()
	defer c.RUnlock()

	cc, ok := c.m[name]
	if ok {
		return cc
	}

	return nil
}

func (c *characters) getCharacters() []*CharacterConfig {
	c.RLock()
	defer c.RUnlock()

	ccs := []*CharacterConfig{}
	for _, cc := range c.m {
		ccs = append(ccs, cc)
	}

	return ccs
}
