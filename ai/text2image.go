/*
 * mastercoderk@gmail.com
 */

package ai

import (
	"bytes"
	"chloe/def"
	"context"
	"encoding/base64"
	"image/png"
	"os"
	"time"

	log "github.com/jeanphorn/log4go"
	"github.com/sashabaranov/go-openai"
)

const (
	ImageGenerateTimeout = 120 * time.Second
)

// / image generate
type dalle struct {
	client *openai.Client
}

func NewImageGenerator(apiKey string) def.ImageGenerator {
	client := getOpenAIClient(apiKey)
	return &dalle{
		client: client,
	}
}

func (d *dalle) Generate(desc, size string) (string, def.CleanFunc, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ImageGenerateTimeout)
	defer cancel()
	var imgSize string
	switch size {
	case "m":
		imgSize = openai.CreateImageSize512x512
	case "s":
		imgSize = openai.CreateImageSize256x256
	case "b":
		imgSize = openai.CreateImageSize1024x1024
	}
	reqBase64 := openai.ImageRequest{
		Prompt:         desc,
		Size:           imgSize,
		ResponseFormat: openai.CreateImageResponseFormatB64JSON,
		N:              1,
	}

	respBase64, err := d.client.CreateImage(ctx, reqBase64)
	if err != nil {
		log.Error(`image creation error on description " %s", %v`, desc, err)
		return "", nil, err
	}

	imgBytes, err := base64.StdEncoding.DecodeString(respBase64.Data[0].B64JSON)
	if err != nil {
		log.Error("base64 decode error: %v", err)
		return "", nil, err
	}

	r := bytes.NewReader(imgBytes)
	imgData, err := png.Decode(r)
	if err != nil {
		log.Error("PNG decode error: %v", err)
		return "", nil, err
	}

	f, err := os.CreateTemp("", "*.png")
	if err != nil {
		log.Error("File creation error: %v", err)
		return "", nil, err
	}
	defer f.Close()
	fname := f.Name()

	if err := png.Encode(f, imgData); err != nil {
		log.Error("PNG encode error: %v", err)
		return "", nil, err
	}

	return fname, func() {
		_ = os.Remove(fname)
	}, nil
}
