package main

import (
	"errors"
)

const (
 	ConfigPath string = "config.json"
	HubSignature string = "stickhub"
	HubSignatureLength = 8
	DefaultEmoji string = "🥰"
)

var (
	ErrNotStickerHub = errors.New("not a sticker hub")
)
