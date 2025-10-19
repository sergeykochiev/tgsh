package main

import (
	"io"
	"github.com/google/uuid"
	"bytes"
	"net/http"
	"fmt"
	"errors"
	"os"
	"github.com/sergeykochiev/tgsh/png"
	"encoding/json"
)

const Token string = ""
const ConfigPath string = "config.json"
var (
	ErrNotStickerHub = errors.New("Not a sticker hub")
)

type SHFileInfo struct {
	
}

type TelegramFile struct {
	FilePath string `json:"file_path"`
}

type Config struct {
	SetName string
	UserId int
}

type TelegramSticker struct {
	FileId string `json:"file_id"`
}

type TelegramSet struct {
	Name string `json:"name"`
	Title string `json:"title"`
	StickerType string `json:"sticker_type"`
	Stickers []TelegramSticker `json:"stickers"`
}

type StickerHub struct {
	fileCount int
	info map[string]uint16
	telegramSet TelegramSet
}

func hiword(n uint16) byte {
	return byte(n >> 8)
}

func loword(n uint16) byte {
	return byte(n & 0xFF)
}
	
func (sh* StickerHub) OfNewSet(userId int, title string) error {
	err := createNewStickerSet(TelegramCreateNewStickerSetParams{
		UserId
		Name
		Title
		Stickers
	})
}

func (sh StickerHub) CreateEmptyInfo() string {
	sh.info = make(map[string]uint16)
	return ""
}

func (sh* StickerHub) ListFiles() {
	if sh.fileCount < 0 {
		fmt.Printf("Stickerhub \"%s\" is empty\n", sh.telegramSet.Title)
	}
	fmt.Printf("Files in stickerhub \"%s\" (%d total):", sh.telegramSet.Title)
	for filename, _ := range(sh.info) {
		fmt.Println(filename)
	}
}

func (sh* StickerHub) ParseInfoFile() error {
	var p png.PngImage
	var fileData []byte
	var err error
	var concatData []byte
	for i := range(sh.fileCount) {
		fileData, err = getFileData(sh.telegramSet.Stickers[i].FileId)
		if err != nil {
			return err
		}
		_, err = p.From(fileData)
		if !p.IsFullOfData() {
			break
		}
		concatData = append(concatData, fileData...)
	}
	err = json.Unmarshal(concatData, &sh.info)
	if err != nil {
		return ErrNotStickerHub
	}
	return nil
}

func (sh* StickerHub) OfFetchedSet(name string) (err error) {
	sh.telegramSet, err = getStickerSet(name)
	if err != nil {
		return
	}
	sh.fileCount = len(sh.telegramSet.Stickers)
	return
}

type TelegramGetStickerSetParams struct {
	Name string `json:"name"`
}

type TelegramGetFileParams struct {
	FileId string `json:"file_id"`
}

type TelegramCreateNewStickerSetParams struct {
	UserId int `json:"user_id"`
	Name string `json:"name"`
	Title string `json:"title"`
	Stickers []TelegramSticker `json:"stickers"`
}


func botUrl(token string, endpoint string) string {
	return "https://api.telegram.org/bot" + token + "/" + endpoint
}

func botFileUrl(token string, filePath string) string {
	return "https://api.telegram.org/file/bot" + token + "/" + filePath
}

func fetch(url string, method string, body any) (*http.Response, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch: %s", err)
	}
	req, err := http.NewRequest(method, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch: %s", err)
	}
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch: %s", err)
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("failed to fetch: status is %d", res.StatusCode)
	}
	return res, nil
}

func getStickerSet(name string) (TelegramSet, error) {
	var set TelegramSet
	res, err := fetch(botUrl(Token, "getStickerSet"), "GET", TelegramGetStickerSetParams{
		Name: name,
	})
	if err != nil {
		return set, fmt.Errorf("failed to get sticker set: %s", err)
	}
	err = json.NewDecoder(res.Body).Decode(&set)
	if err != nil {
		return set, fmt.Errorf("failed to get sticker set: %s", err)
	}
	return set, nil
}

func getFile(fileId string) (TelegramFile, error) {
	var file TelegramFile
	res, err := fetch(botUrl(Token, "getFile"), "GET", TelegramGetFileParams{
		FileId: fileId,
	})
	if err != nil {
		return file, fmt.Errorf("failed to get file: %s", err)
	}
	err = json.NewDecoder(res.Body).Decode(&file)
	if err != nil {
		return file, fmt.Errorf("failed to get file: %s", err)
	}
	return file, nil
}

// TODO
func uploadFile() {
}

func createNewStickerSet(params TelegramCreateNewStickerSetParams) error {
	_, err := fetch(botUrl(Token, "createNewStickerSet"), "POST", params)
	if err != nil {
		return fmt.Errorf("failed to get file: %s", err)
	}
	return nil
}	

func downloadFile(file TelegramFile) ([]byte, error) {
	res, err := fetch(botFileUrl(Token, file.FilePath), "GET", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %s", err)
	}
	var body []byte
	body, err = io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %s", err)
	}
	return body, nil
}

func getFileData(fileId string) ([]byte, error) {
	file, err := getFile(fileId)
	if err != nil {
		return nil, err
	}
	return downloadFile(file)
}

func promptString(message string) (string, error) {
	var out string
	fmt.Print(message)
	_, err := fmt.Scanln(&out)
	if err != nil {
		return out, fmt.Errorf("failed to propmt: %s", err)
	}
	return out, nil
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

func promptBool(message string, retryCount int) (bool, error) {
	var res string
	var err error
	fmt.Printf("%s (y|n)\n", message)
	for range(retryCount) {
		_, err = fmt.Scanln(&res)
		if res == "y" {
			return true, nil
		} else if res == "n" {
			return false, nil
		}
	}
	return false, fmt.Errorf("failed to prompt: %s", err)
}

func isConfigured() (bool, error) {
	_, err := os.Stat(ConfigPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return false, fmt.Errorf("failed to check if file exists: %s", err)
	}
	return !errors.Is(err, os.ErrNotExist), nil
}

func createConfigFile(config Config) error {
	file, err := os.OpenFile(ConfigPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to create config file: %s", err)
	}
	defer file.Close()
	err = json.NewEncoder(file).Encode(config)
	return err
}

func getConfigFromFile() (Config, error) {
	var config Config
	file, err := os.OpenFile(ConfigPath, os.O_RDONLY, 0644)
	if err != nil {
		return config, fmt.Errorf("failed to get config from file: %s", err)
	}
	defer file.Close()
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		return config, fmt.Errorf("failed to get config from file: %s", err)
	}
	return config, nil
}

func configureInit() (Config, error) {
	var config Config
	var err error
	// config.SetName, err = promptString("Configure set name:")
	// if err != nil {
	// 	return config, err
	// }
	config.UserId, err = promptInt("Configure user id:", 3)
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

func generateNewSetName() string {
	return "stickerhub_" + uuid.New().String()
}

func create(config Config, initFile []byte) error {
	if !config.HasUser() {
		fmt.Println("User not found in config")
		return nil
	}
	config.SetName = generateNewSetName()
	err := createNewStickerSet(TelegramCreateNewStickerSetParams{
		UserId: config.UserId,
		Name: config.SetName,
		Title: config.SetName,
		Stickers: []TelegramSticker{ initFile },
	})
	if err != nil {
		return err
	}
	err = createConfigFile(config)
	if err != nil {
		return err
	}
	fmt.Printf("Created a new hub (sticker set) with name \"%s\"", config.SetName)
	return nil
}

func getFileFromHub(config Config) ([]byte, error) {
	if !config.HasSet() {
		fmt.Println("Set not found in config: use 'setset' option to set a set first")
		return nil
	}
	var fileBytes []byte
	set, err := getStickerSet(config.SetName)
	if err != nil {
		return fileBytes, fmt.Errorf("failed to get file from hub: %s", err)
	}
	count := len(set.Stickers)
	if count == 0 {
		fmt.Println("No files found in the hub")
		return fileBytes, nil
	}
	index, err := promptInt(fmt.Sprintf("Select file index (%d total):", count), 3)
	if err != nil {
		return fileBytes, fmt.Errorf("failed to get file from hub: %s", err)
	}
	if index - 1 >= count || index - 1 <= 0 {
		return fileBytes, fmt.Errorf("failed to get file from hub: %s", err)
	}
	var file TelegramFile
	file, err = getFile(set.Stickers[index].FileId)
	if err != nil {
		return fileBytes, fmt.Errorf("failed to get file from hub: %s", err)
	}
	fileBytes, err = downloadFile(file)
	if err != nil {
		return fileBytes, fmt.Errorf("failed to get file from hub: %s", err)
	}
	return fileBytes, nil
}

func setSetName(config Config) error {
	config.SetName, err = promptString("Update set name:")
	if err != nil {
		return nil
	}
	_, err = getStickerSet(config.SetName)
	if err != nil {
		return err
	}
	return createConfigFile(config)
}

func createNewOrSetSet(config Config, initFileId string) error {
	createSet, err := promptBool("Set not found in config. Create new set?", 3)
	if err != nil {
		return err
	}
	if createSet {
		return createHub(config)
	}
	return errors.New("Set not found in config: use 'setset' option to set a set first")
}

func putFileIntoHub(config Config, filename string) error {
	var err error
	fileBytes, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	if !config.HasSet() {
		err = createNewOrSetSet()
		if err != nil {
			return err
		}
		config, err = getConfigFromFile()
		if err != nil {
			return err
		}
	}
	fmt.Println("NOT IMPLEMENTED")
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("not enough arguments")
		return
	}
	config, err := getConfig()
	if err != nil {
		fmt.Println(err)
		return
	}
	switch os.Args[1] {
		case "put":
			if len(os.Args) < 3 {
				fmt.Println("not enough arguments")
				break
			}
			err = putFileIntoHub(config, os.Args[2])
			if err != nil {
				fmt.Println(err)
			}
		case "get":
			if !config.HaveSet() {
				fmt.Println("use 'set' to set a set first")
				break
			}
			var file []byte
			file, err = getFileFromHub(config)
			if err != nil {
				fmt.Println(err)
				break
			}
			fmt.Println(file)
		case "setset": 
			err = setSetName(config)
			if err != nil {
				fmt.Println(err)
			}
		default:
			fmt.Println("unknown option")
	}
}
