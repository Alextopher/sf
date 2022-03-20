package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

type SoundFont struct {
	// Info holds the sound font information present in the INFO chunk.
	Info *SoundFontInfo

	// Sound holds the sound data present in the SDTA chunk.
	// Can either be 16-bit or 24-bit samples.
	Samples *SoundFontSamples

	// The Preset, Instrument, and Sample Header data
	Hydra *SoundFontHydra
}

// Expect reads len(b) bytes from r and checks that they match b.
func Expect(r io.Reader, b []byte) (bool, error) {
	buf := make([]byte, len(b))
	if _, err := io.ReadFull(r, buf); err != nil {
		return false, err
	}
	return bytes.Equal(buf, b), nil
}

func ReadSoundFont(r io.Reader) (*SoundFont, error) {
	// Read the RIFF header.
	var riffHeader chunk
	if err := riffHeader.expect(r, [4]byte{'R', 'I', 'F', 'F'}); err != nil {
		return nil, err
	}
	r = riffHeader.newReader()

	// read "sfbk" from the RIFF header
	ok, err := Expect(r, []byte{'s', 'f', 'b', 'k'})
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("expected sfbk")
	}

	// read the "LIST" header
	var listHeader chunk
	if err := listHeader.expect(r, [4]byte{'L', 'I', 'S', 'T'}); err != nil {
		return nil, err
	}
	listReader := listHeader.newReader()

	info, err := ReadSoundFontInfo(listReader)
	if err != nil {
		return nil, err
	}

	// read the next "LIST" header
	if err := listHeader.expect(r, [4]byte{'L', 'I', 'S', 'T'}); err != nil {
		return nil, err
	}
	listReader = listHeader.newReader()

	// read "sdta" from the "LIST" header
	ok, err = Expect(listReader, []byte{'s', 'd', 't', 'a'})
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("expected sdta")
	}
	sound, err := ReadSoundFontSamples(listReader)
	if err != nil {
		return nil, err
	}

	// read the last "LIST" header
	if err := listHeader.expect(r, [4]byte{'L', 'I', 'S', 'T'}); err != nil {
		return nil, err
	}
	listReader = listHeader.newReader()

	// read "pdta" from the "LIST" header
	ok, err = Expect(listReader, []byte{'p', 'd', 't', 'a'})
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("expected pdta")
	}

	hydra, err := ReadSoundFontHydra(listReader)
	if err != nil {
		return nil, err
	}

	// sink remaining data
	n, err := io.Copy(io.Discard, listReader)
	if err != nil {
		return nil, err
	}
	fmt.Println("sunk", n, "bytes")

	return &SoundFont{
		Info:    info,
		Samples: sound,
		Hydra:   hydra,
	}, nil
}

func main() {
	// open the test file
	f, err := os.Open("test.sf2")
	if err != nil {
		panic(err)
	}

	defer f.Close()

	// read the file
	_, err = ReadSoundFont(f)
	if err != nil {
		panic(err)
	}

	// do something with the sound font
	// ...
	// fmt.Println(sf)
}
