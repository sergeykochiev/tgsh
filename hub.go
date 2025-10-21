package main

import (
	"fmt"
	"os"
	"encoding/json"
	"github.com/sergeykochiev/tgsh/png"
	"github.com/sergeykochiev/tgsh/webp"
)

type StickerHub struct {
	botUsername string
	fileCount int
	userId int
	info map[string]uint16
	telegramSet TelegramSet
}

func (sh* StickerHub) GetUsername() error {
	user, err := getMe()
	if err != nil {
		return err
	}
	sh.botUsername = user.Username
	return nil
}

func (sh StickerHub) UserId() int {
	return sh.userId
}

func (sh* StickerHub) FromNewSet(userId int, title string) error {
	headerData := sh.createEmptyHeader()
	file, err := uploadStickerFile(userId, "header", headerData)
	if err != nil {
		return fmt.Errorf("upload sticker file: %s", err)
	}
	name := generateNewSetName(sh.botUsername)
	fmt.Printf("Creating set with name \"%s\"\n", name)
	ok, err := createNewStickerSet(TelegramParamsCreateNewStickerSet{
 		UserId: userId,
 		Name: name,
 		Title: title,
 		Stickers: []TelegramInputSticker{
			{
				FileId: file.Id,
				Format: "static",
				EmojiList: []string{ DefaultEmoji },
			},
		},
 	})
	if err != nil {
		return fmt.Errorf("create new sticker set: %s", err)
	}
	if !ok {
		return fmt.Errorf("create new sticker set: returned false")
	}
	sh.userId = userId
	fmt.Printf("Created new set. Name is \"%s\"\n", name)
	return sh.FromExistingSet(userId, name)
}

func (sh StickerHub) encodeDataToPng(data []byte) []byte {
	var p png.PngImage
	fmt.Println(len(data))
	p.Default(512, 512, data)
	return p.Encode()
}

func (sh StickerHub) createEmptyHeader() []byte {
	return sh.encodeDataToPng(append([]byte(HubSignature), []byte{'{', '}'}...))
}

func (sh* StickerHub) ListFiles() {
	if sh.fileCount < 0 {
		fmt.Printf("Stickerhub \"%s\" is empty\n", sh.telegramSet.Title)
	}
	fmt.Printf("Files in stickerhub \"%s\" (%d total):", sh.telegramSet.Title, sh.fileCount)
	for filename, _ := range(sh.info) {
		fmt.Println(filename)
	}
}

func (sh* StickerHub) decodeFileData(fileData []byte) ([]byte, error) {
	return webp.Decode(fileData)
}

func (sh* StickerHub) getHeaderData() ([]byte, error) {
	var concatData []byte
	if sh.fileCount == 0 {
		return nil, fmt.Errorf("file count is 0")
	}
	fileData, err := getFileData(sh.telegramSet.Stickers[0].FileId)
	if err != nil {
		return nil, fmt.Errorf("get file data: %s", err)
	}
	decoded, err := sh.decodeFileData(fileData)
	if err != nil {
		return nil, fmt.Errorf("decode file data: %s", err)
	}
	concatData = append(concatData, decoded...) 
	return concatData, nil
}

func (sh* StickerHub) UploadFile(filename string) error {
	fileData, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("read file: %s", err)
	}
	file, err := uploadStickerFile(sh.userId, filename, sh.encodeDataToPng(fileData))
	if err != nil {
		return fmt.Errorf("upload sticker file: %s", err)
	}
	ok, err := addStickerToSet(TelegramParamsAddStickerToSet{
		UserId: sh.userId,
		Name: sh.telegramSet.Name,
		Sticker: TelegramInputSticker{
			FileId: file.Id,
			Format: "static",
			EmojiList: []string{ DefaultEmoji },
		},
	})
	if err != nil {
		return fmt.Errorf("add sticker to set: %s", err)
	}
	if !ok {
		return fmt.Errorf("add sticker to set: returned false")
	}
	return nil
}

func (sh* StickerHub) parseHeader() error {
	data, err := sh.getHeaderData()
	if err != nil {
		return fmt.Errorf("get header data: %s", err)
	}
	if len(data) < HubSignatureLength {
		return fmt.Errorf("data is shorter than signature: %s", ErrNotStickerHub)
	}
	signature := data[:HubSignatureLength]
	if string(signature) != HubSignature {
		return fmt.Errorf("invalid or nonexistent signature: %s", ErrNotStickerHub)
	}
	err = json.Unmarshal(data[HubSignatureLength:], &sh.info)
	if err != nil {
		return fmt.Errorf("json decode Hub Info: %s", ErrNotStickerHub)
	}
	return nil
}

func (sh* StickerHub) FromExistingSet(userId int, name string) error {
	var err error
	sh.telegramSet, err = getStickerSet(name)
	if err != nil {
		return fmt.Errorf("get sticker set: %s", err)
	}
	sh.fileCount = len(sh.telegramSet.Stickers)
	sh.userId = userId
	err = sh.parseHeader()
	if err != nil {
		return fmt.Errorf("parse header: %s", err)
	}
	return nil
}
