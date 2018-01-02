package glas

import "errors"

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
)

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
