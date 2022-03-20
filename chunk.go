package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// SoundFont uses the RIFF (Resource Interchange File Format) file format.
// The RIFF file format is a generic file format for storing data.
type chunk struct {
	// The chunk id, normally four ASCII characters.
	id [4]byte
	// 4 bytes, little endian, the size of the chunk data.
	size uint32
	// The chunk data.
	data []byte
}

// parse reads a chunk from the reader.
func (ck *chunk) parse(r io.Reader) error {
	// First read the chunk id and size.
	if _, err := io.ReadFull(r, ck.id[:]); err != nil {
		return err
	}

	// Read the chunk size.
	if err := binary.Read(r, binary.LittleEndian, &ck.size); err != nil {
		return err
	}

	// Read the chunk data.
	ck.data = make([]byte, ck.size)
	if _, err := io.ReadFull(r, ck.data); err != nil {
		return err
	}

	// fmt.Println(string(ck.id[:]), ck.size, len(ck.data))
	return nil
}

// expect reads a chunk from the reader and checks that it's id matches the
// expected id.
func (ch *chunk) expect(r io.Reader, id [4]byte) error {
	// Read the chunk
	if err := ch.parse(r); err != nil {
		return err
	}

	if ch.id != id {
		return fmt.Errorf("expected chunk id %v, got %v", id, ch.id)
	}

	return nil
}

// newReader returns a new reader of the chunk's data.
func (ch *chunk) newReader() io.Reader {
	return bytes.NewReader(ch.data)
}
