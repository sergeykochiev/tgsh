package main

import (
	"fmt"
	"strconv"
	"errors"
	"os"
)

func usage() {
	fmt.Println("Usage:")
	fmt.Println("-", "set", "<user id> <sticker set name | \"new\">", ":", "configure hub")
	fmt.Println("-", "put", "<filename>", ":", "put file into hub")
	fmt.Println("-", "list", ":", "list files in hub")
	fmt.Println("-", "get", "<file index>", ":", "download file from hub by index")
}

func cmdset(c *Config, sh *StickerHub, argc int, argv []string) error {
	var err error
	if argc < 4 {
		usage()
		return nil
	}
	c.UserId, err = strconv.Atoi(argv[2])
	if err != nil {
		return errors.New("Unparsable user id")
	}
	err = sh.OfUser(c.UserId)
	if err != nil {
		return errors.New("Invalid user id")
	}
	if argv[3] == "new" {
		err = sh.FromNewSet(promptString("Title:"))
	} else {
		err = sh.FromExistingSet(argv[3])
	}
	if err != nil {
	  return err
	}
	c.SetName = sh.telegramSet.Name
	c.WriteFile()
	return nil
}

func cmdput(c *Config, sh *StickerHub, argc int, argv []string) error {
	if !c.IsConfigured() {
		return errors.New("Use set to configure first")
	}
	if argc < 3 {
		usage()
		return nil
	}
	return sh.UploadFile(argv[2])
}

func cmdlist(c *Config, sh *StickerHub) error {
	if !c.IsConfigured() {
		return errors.New("Use set to configure first")
	}
	sh.ListFiles()
	return nil
}

func cmdget(c *Config, sh *StickerHub, argc int, argv []string) error {
	if !c.IsConfigured() {
		return errors.New("Use set to configure first")
	}
	if argc < 3 {
		usage()
		return nil
	}
	index, err := strconv.Atoi(argv[2])
	if err != nil {
		return errors.New("Unparsable file index")
	}
	if index > sh.fileCount - 1 || index <= 0 {
		return errors.New("Invalid index")
	}
	file, err := sh.GetFile(sh.telegramSet.Stickers[index].FileId)
	if err != nil {
		return err
	}
	return os.WriteFile(sh.GetInfoEntry(index - 1).Filename, file, 0644)
}

func cmd(c *Config, sh *StickerHub, argc int, argv []string) error {
	if argc < 2 {
		usage()
		return nil
	}
	switch os.Args[1] {
	case "set": return cmdset(c, sh, argc, argv);
	case "get": return cmdget(c, sh, argc, argv);
	case "put": return cmdput(c, sh, argc, argv);
	case "list": return cmdlist(c, sh);
	default:
		usage()
		return nil
	}
}

func main() {
	var c Config
	err := c.GetOrCreate()
	if err != nil {
		fmt.Println("Failed to get config:", err)
		return
	}

	var sh StickerHub
	err = sh.GetUsername()
	if err != nil {
		fmt.Println("Failed to get bot username:", err)
		return
	}

	if c.IsConfigured() {
		sh.OfUser(c.UserId)
		sh.FromExistingSet(c.SetName)
	}

	argc := len(os.Args)

	err = cmd(&c, &sh, argc, os.Args)
	if err != nil {
		fmt.Println(err)
	}
}
