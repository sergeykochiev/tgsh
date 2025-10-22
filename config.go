package main

import(
	"os"
	"fmt"
	"encoding/json"
	"errors"
)

type Config struct {
	UserId int
	SetName string
}

func (c* Config) GetOrCreate() error {
	ok, err := c.isFileExists()
	if err != nil {
		return err
	}
	if !ok {
		return c.WriteFile()
	}
	return c.FromFile()
}

func (c Config) IsConfigured() bool {
	return c.UserId != 0 && c.SetName != ""
}

func (c Config) isFileExists() (bool, error) {
	_, err := os.Stat(ConfigPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return false, fmt.Errorf("check if file exists: %s", err)
	}
	return !errors.Is(err, os.ErrNotExist), nil
}

func (c Config) WriteFile() error {
	file, err := os.OpenFile(ConfigPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open file: %s", err)
	}
	defer file.Close()
	err = json.NewEncoder(file).Encode(c)
	return err
}

func (c* Config) FromFile() error {
	file, err := os.OpenFile(ConfigPath, os.O_RDONLY, 0644)
	if err != nil {
		return fmt.Errorf("open file: %s", err)
	}
	defer file.Close()
	err = json.NewDecoder(file).Decode(c)
	if err != nil {
		return fmt.Errorf("json decode Config: %s", err)
	}
	return nil
}
