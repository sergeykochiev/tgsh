package main

type TelegramResponse[T any] struct {
	Ok bool `json:"ok"`
	Result T `json:"result"`
	Desc string `json:"description"`
}

type TelegramFile struct {
	Id string `json:"file_id"`
	UniqueId string `json:"file_unique_id"`
	Path string `json:"file_path"`
	Size int `json:"file_size"`
}

type TelegramUser struct {
	Username string `json:"username"`
}

type Config struct {
	UserId int
}

type TelegramSticker struct {
	FileId string `json:"file_id"`
}

type TelegramInputSticker struct {
	FileId string `json:"sticker"`
	Format string `json:"format"`
	EmojiList []string `json:"emoji_list"`
}

type TelegramSet struct {
	Name string `json:"name"`
	Title string `json:"title"`
	StickerType string `json:"sticker_type"`
	Stickers []TelegramSticker `json:"stickers"`
}
