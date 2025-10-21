package main

import (
	"fmt"
	"encoding/json"
	"net/http"
	"io"
	"bytes"
	"mime/multipart"
)

type TelegramParamsReplaceStickerInSet struct {
	UserId int `json:"user_id"`
	Name string `json:"name"`
	OldFileId string `json:"old_sticker"`
	Sticker TelegramInputSticker `json:"sticker"`
}

type TelegramParamsGetStickerSet struct {
	Name string `json:"name"`
}

type TelegramParamsGetFile struct {
	FileId string `json:"file_id"`
}

type TelegramParamsCreateNewStickerSet struct {
	UserId int `json:"user_id"`
	Name string `json:"name"`
	Title string `json:"title"`
	Stickers []TelegramInputSticker `json:"stickers"`
}

type TelegramParamsAddStickerToSet struct {
	UserId int `json:"user_id"`
	Name string `json:"name"`
	Sticker TelegramInputSticker `json:"sticker"`
}

func botUrl(endpoint string) string {
	return "https://api.telegram.org/bot" + getToken() + "/" + endpoint
}

func botFileUrl(filePath string) string {
	return "https://api.telegram.org/file/bot" + getToken() + "/" + filePath
}

func makeRequest(url string, method string, body io.Reader, header http.Header) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("\"%s\" create request: %s", url, err)
	}
	req.Header = header
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("\"%s\" make request: %s", url, err)
	}
	return res, nil
}

func makeJsonRequest(url string, body any) (*http.Response, error) {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return makeRequest(url, "POST", bytes.NewReader(jsonData), http.Header{
		"Content-Type": []string{ "application/json" },
	})
}

func fetch[T any](endpoint string, method string, params any) (T, error) {
	var resData TelegramResponse[T]
	var res *http.Response
	var err error
	if method == "POST" {
		res, err = makeJsonRequest(botUrl(endpoint), params)
	} else {
		res, err = makeRequest(botUrl(endpoint), "GET", nil, nil)
	}
	if err != nil {
		return resData.Result, fmt.Errorf("fetch: %s", err)
	}
	err = json.NewDecoder(res.Body).Decode(&resData)
	if err != nil {
		return resData.Result, fmt.Errorf("json decode Telegram Set: %s", err)
	}
	if !resData.Ok || res.StatusCode != 200 {
		return resData.Result, fmt.Errorf("response is not OK: status is %d, desc is %s", res.StatusCode, resData.Desc)
	}
	return resData.Result, nil
}

func getStickerSet(name string) (TelegramSet, error) {
	return fetch[TelegramSet]("getStickerSet", "POST", TelegramParamsGetStickerSet{ Name: name })
}

func getFile(fileId string) (TelegramFile, error) {
	return fetch[TelegramFile]("getFile", "POST", TelegramParamsGetFile{ FileId: fileId })
}

func getMe() (TelegramUser, error) {
	return fetch[TelegramUser]("getMe", "GET", nil)
}

func replaceStickerInSet(params TelegramParamsReplaceStickerInSet) (bool, error) {
	return fetch[bool]("replaceStickerInSet", "POST", params)
}

func uploadStickerFile(userId int, filename string, fileData []byte) (TelegramFile, error) {
	var resData TelegramResponse[TelegramFile]
	var b bytes.Buffer
	var vw io.Writer
	var err error
	w := multipart.NewWriter(&b)
	if vw, err = w.CreateFormField("user_id"); err != nil {
		return resData.Result, err
	}
	if _, err := vw.Write(fmt.Appendf([]byte{}, "%d", userId)); err != nil {
		return resData.Result, err
	}
	if vw, err = w.CreateFormField("sticker_format"); err != nil {
		return resData.Result, err
	}
	if _, err := vw.Write([]byte("static")); err != nil {
		return resData.Result, err
	}
	if vw, err = w.CreateFormFile("sticker", filename); err != nil {
		return resData.Result, err
	}
	if _, err := vw.Write(fileData); err != nil {
		return resData.Result, err
	}
	h := make(http.Header)
	h.Add("Content-Type", w.FormDataContentType())
	w.Close()
	res, err := makeRequest(botUrl("uploadStickerFile"), "POST", &b, h)
	if err != nil {
		return resData.Result, fmt.Errorf("fetch: %s", err)
	}
	err = json.NewDecoder(res.Body).Decode(&resData)
	if err != nil {
		return resData.Result, fmt.Errorf("json decode Telegram File: %s", err)
	}
	if !resData.Ok || res.StatusCode != 200 {
		return resData.Result, fmt.Errorf("response is not OK: status is %d, desc is %s", res.StatusCode, resData.Desc)
	}
	return resData.Result, nil
}

func addStickerToSet(params TelegramParamsAddStickerToSet) (bool, error) {
	return fetch[bool]("addStickerToSet", "POST", params)
}

func createNewStickerSet(params TelegramParamsCreateNewStickerSet) (bool, error) {
	return fetch[bool]("createNewStickerSet", "POST", params)
}

func downloadFile(file TelegramFile) ([]byte, error) {
	res, err := makeRequest(botFileUrl(file.Path), "GET", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("fetch: %s", err)
	}
	var body []byte
	body, err = io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %s", err)
	}
	return body, nil
}
