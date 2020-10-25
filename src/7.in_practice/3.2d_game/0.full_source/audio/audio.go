package audio

import (
	"io"
	"os"

	"github.com/cryptix/wav"
	"github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto"
)

var sampleRate int = 44100

func Play(file string, repeat bool) {
	var decoder io.Reader
	var sampleRate int
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	isMp3 := file[len(file)-4:] == ".mp3"
	if isMp3 {
		d, err := mp3.NewDecoder(f)
		if err != nil {
			panic(err)
		}
		sampleRate = d.SampleRate()
		decoder = d
		/*
			c, err := oto.NewContext(d.SampleRate(), 2, 2, 8192)
			if err != nil {
				panic(err)
			}
			defer c.Close()
		*/
	} else {
		fInfo, err := os.Stat(file)
		if err != nil {
			panic(err)
		}

		w, err := wav.NewReader(f, fInfo.Size())
		if err != nil {
			panic(err)
		}

		decoder, err = w.GetDumbReader()
		if err != nil {
			panic(err)
		}
		sampleRate = int(w.GetSampleRate())
	}

	panic(sampleRate)

	c, err := oto.NewContext(sampleRate, 2, 2, 8192)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	p := c.NewPlayer()
	defer p.Close()
	for {
		if _, err := io.Copy(p, decoder); err != nil {
			panic(err)
		}

		if !repeat {
			break
		}

		if isMp3 {

		} else {
			//d, err = w.GetDumbReader()
		}
	}
}
