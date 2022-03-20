package main

import "io"

type SoundFontSamples struct {
	// Samples the Digital Audio Samples for the upper 16 bits
	SamplesHigher []int16

	// SamplesLower optionally holds the Digital Audio Samples for the lower 8 bits
	// of the samples.
	SamplesLower []int8
}

func ReadSoundFontSamples(r io.Reader) (*SoundFontSamples, error) {
	sound := &SoundFontSamples{}

	// read the "smpl" header
	var smplHeader chunk
	if err := smplHeader.expect(r, [4]byte{'s', 'm', 'p', 'l'}); err != nil {
		return nil, err
	}

	// The smpl sub-chunk, if present, contains one or more “samples” of digital audio information in the form of linearly coded
	// sixteen bit, signed, little endian (least significant byte first) words.
	sound.SamplesHigher = make([]int16, smplHeader.size/2)
	for i := 0; i < len(sound.SamplesHigher); i++ {
		sound.SamplesHigher[i] = int16(smplHeader.data[i*2+1])<<8 | int16(smplHeader.data[i*2])<<8
	}

	// optionally read the "sm24" sub-chunk
	var sm24Header chunk
	if err := sm24Header.expect(r, [4]byte{'s', 'm', '2', '4'}); err != nil {
		if err == io.EOF {
			return sound, nil
		}
		return nil, err
	}

	// The sm24 sub-chunk, if present, contains the least significant byte counterparts to each sample data point contained in the
	// smpl chunk. Note this means for every two bytes in the [smpl] sub-chunk there is a 1-byte counterpart in [sm24] sub-chunk.
	sound.SamplesLower = make([]int8, sm24Header.size)
	for i := 0; i < len(sound.SamplesLower); i++ {
		sound.SamplesLower[i] = int8(sm24Header.data[i])
	}

	return sound, nil
}
