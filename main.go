package main

import (
	"fmt"
	"errors"
	"os"
	"encoding/json"
)

func isConfigured() (bool, error) {
	_, err := os.Stat(ConfigPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return false, fmt.Errorf("check if file exists: %s", err)
	}
	return !errors.Is(err, os.ErrNotExist), nil
}

func createConfigFile(config Config) error {
	file, err := os.OpenFile(ConfigPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open file: %s", err)
	}
	defer file.Close()
	err = json.NewEncoder(file).Encode(config)
	return err
}

func getConfigFromFile() (Config, error) {
	var config Config
	file, err := os.OpenFile(ConfigPath, os.O_RDONLY, 0644)
	if err != nil {
		return config, fmt.Errorf("open file: %s", err)
	}
	defer file.Close()
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		return config, fmt.Errorf("json decode Config: %s", err)
	}
	return config, nil
}

func configureInit() (Config, error) {
	var config Config
	var err error
	config.UserId, err = promptInt("Configure user id: ", 3)
	if err != nil {
		return config, err
	}
	err = createConfigFile(config)
	return config, err
}

func getConfig() (Config, error) {
	isConfigured, err := isConfigured()
	if err != nil {
		return Config{}, err
	}
	if !isConfigured {
		return configureInit()
	}
	return getConfigFromFile()
}

func main() {
	config, err := getConfig()
	if err != nil {
		fmt.Println("Failed to get config:", err)
		return
	}
	var sh StickerHub

	argc := len(os.Args)
	if argc < 2 {
		fmt.Println("Provide set name (\"new\" for new one)")
		return
	}
	if argc < 3 {
		fmt.Println("Provide action (actions include \"put\" and \"get\"")
		return
	}
	if os.Args[2] != "put" && os.Args[2] != "get" {
		fmt.Println("Unknown action", os.Args[2], "(actions include \"put\" and \"get\"")
		return
	}

	err = sh.GetUsername()
	if err != nil {
		fmt.Println("Failed to get bot username:", err)
		return
	}

	switch os.Args[1] {
	case "new":
		fmt.Println("Proceeding to create a new set for a stickerhub...")
		title := promptString("What should the title be? ")
		err = sh.FromNewSet(config.UserId, title)
		if err != nil {
			fmt.Println("Failed to create a new set:", err)
			return
		}
	default:
		fmt.Println("Proceeding to fetch set...")
		err = sh.FromExistingSet(config.UserId, os.Args[1])
		if err != nil {
			fmt.Println("Failed to fetch existing set:", err)
			return
		}
	}

	switch os.Args[2] {
	case "put":
		if argc < 4 {
			fmt.Println("Provide filename")
			break
		}
		err = sh.UploadFile(os.Args[3])
		if err != nil {
			fmt.Println("Failed to upload file:", err)
			break
		}
	case "get":
		sh.ListFiles()
		if sh.fileCount <= 0 {
			return
		}
		index, err := promptInt("Index: ", 3)
		if err != nil {
			fmt.Println("Failed to prompt for file index:", err)
			return
		}
		if index > sh.fileCount - 1 || index <= 0 {
			fmt.Printf("Invalid index. Only indexes in range %d..%d are valid\n", 1, sh.fileCount - 1)
			return
		}
		file, err := sh.GetFile(sh.telegramSet.Stickers[index].FileId)
		if err != nil {
			fmt.Println("Failed to get file:", err)
			return
		}
		os.WriteFile("downloaded", file, 0644)
	}
}
