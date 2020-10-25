package audio

import (
	"io"
	"os"

	"github.com/cryptix/wav"
	"github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto"
)

var sampleRate int = 44100

type Player struct {
	context *oto.Context
}

func New() *Player {
	c, err := oto.NewContext(sampleRate, 2, 2, 8192)
	if err != nil {
		panic(err)
	}

	return &Player{c}
}

func (p *Player) Play(file string, repeat bool) {
	var decoder io.Reader
	var close func() error
	isMp3 := file[len(file)-4:] == ".mp3"
	if isMp3 {
		decoder, close = getMp3Reader(file)
	} else {
		decoder, close = getWavReader(file)
	}
	defer close()

	player := p.context.NewPlayer()
	defer player.Close()
	for {
		if _, err := io.Copy(player, decoder); err != nil {
			panic(err)
		}

		if !repeat {
			break
		}

		if isMp3 {
			decoder, close = getMp3Reader(file)
		} else {
			decoder, close = getWavReader(file)
		}
		defer close()
	}
}

func getMp3Reader(file string) (io.Reader, func() error) {
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}

	decoder, err := mp3.NewDecoder(f)
	if err != nil {
		panic(err)
	}
	return decoder, f.Close
}

func getWavReader(file string) (io.Reader, func() error) {
	fInfo, err := os.Stat(file)
	if err != nil {
		panic(err)
	}

	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}

	w, err := wav.NewReader(f, fInfo.Size())
	if err != nil {
		panic(err)
	}
	decoder, err := w.GetDumbReader()
	if err != nil {
		panic(err)
	}
	return decoder, f.Close
}
