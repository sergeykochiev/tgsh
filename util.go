package main

import (
	"fmt"
	"os"
	"github.com/google/uuid"
	"strings"
)

func printBytesHex(data []byte) {
	for _, b := range(data) {
		fmt.Printf("%02x ", b)
	}
	fmt.Println()
}

func getToken() string {
	token := os.Getenv("TOKEN")
	if token == "" {
		panic("Provide token in env")
	}
	return token
}

func hiword(n uint16) byte {
	return byte(n >> 8)
}

func loword(n uint16) byte {
	return byte(n & 0xFF)
}

func generateNewSetName(username string) string {
	return "stickerhub_" + strings.ReplaceAll(uuid.New().String(), "-", "_") + "_by_" + username
}

func getFileData(fileId string) ([]byte, error) {
	file, err := getFile(fileId)
	if err != nil {
		return nil, err
	}
	return downloadFile(file)
}

func promptBool(message string, retryCount int) (bool, error) {
	var res string
	var err error
	fmt.Printf("%s (y|n)\n", message)
	for range(retryCount) {
		_, err = fmt.Scanln(&res)
		switch res {
		case "y": return true, nil
		case "n": return false, nil
		}
	}
	return false, fmt.Errorf("failed to prompt: %s", err)
}

func promptString(message string) string {
	var out string
	fmt.Print(message)
	fmt.Scanln(&out)
	return out
}

func promptInt(message string, retryCount int) (int, error) {
	var out int
	var err error
	fmt.Printf(message)
	for range(retryCount) {
		_, err = fmt.Scanln(&out)
		if err == nil {
			return out, nil
		}
	}
	return out, fmt.Errorf("failed to prompt: %s", err)
}
