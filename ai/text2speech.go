/*
 * mastercoderk@gmail.com
 */

package ai

import (
	"chloe/def"
	"context"
	"io/ioutil"
	"os"
	"time"

	pys "chloe/external/service"

	htgotts "github.com/hegedustibor/htgo-tts"
	"github.com/hegedustibor/htgo-tts/voices"
	log "github.com/jeanphorn/log4go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type googleTransTTS struct {
}

func NewGoogleTranslateTTS() def.TextToSpeech {
	return &googleTransTTS{}
}

func (tts *googleTransTTS) Convert(text string) (string, def.CleanFunc, error) {
	dir, err := ioutil.TempDir("", "tts*")
	if err != nil {
		log.Error("directory creation error: %v", err)
		return "", nil, err
	}

	fBasename := "speech"

	speech := htgotts.Speech{Folder: dir, Language: voices.English}
	fname, err := speech.CreateSpeechFile(text, fBasename)
	if err != nil {
		log.Error("failed to convert text to speech, %v", err)
		return "", nil, err
	}

	return fname, func() {
		_ = os.Remove(fname)
		_ = os.Remove(dir)
	}, nil
}

type pyServiceTTS struct {
	client pys.GoogleTranslateTTSClient
}

func NewPyServiceTTS() def.TextToSpeech {
	conn, err := grpc.Dial(
		"127.0.0.1:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Error("can not dial to external grpc service, %v", err)
		return nil
	}
	return &pyServiceTTS{
		client: pys.NewGoogleTranslateTTSClient(conn),
	}
}

func (tts *pyServiceTTS) Convert(text string) (string, def.CleanFunc, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	resp, err := tts.client.TextToSpeech(ctx, &pys.Text{
		Text: text,
	})
	if err != nil {
		log.Error("grpc call failed, %v", err)
		return "", nil, err
	}

	f, err := os.CreateTemp("", "*.mp3")
	if err != nil {
		log.Error("file creation error: %v", err)
		return "", nil, err
	}
	defer f.Close()
	fname := f.Name()

	speech := resp.FileResponse.Data
	if _, err := f.Write(speech); err != nil {
		log.Error("save mp3 failed, %v", err)
		return "", nil, err
	}

	return fname, func() {
		_ = os.Remove(fname)
	}, nil
}
