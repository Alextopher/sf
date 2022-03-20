package main

import (
	"fmt"
	"io"
)

type SoundFontInfo struct {
	// SfVersion identifyies the SoundFont specification version level to which the file complies.
	// e.g. 2.1
	SfVersion struct {
		Major, Minor uint16
	} // made from the ifil subchunk

	// Engine is a mandatory field identifying the wavetable sound engine for which the file was optimized.
	// It contains an ASCII string of 256 or fewer bytes including one or two terminators of value zero, so as to make
	// the total byte count even.
	Engine string // made from the isng subchunk

	// Name is a mandatory field providing the name of the SoundFont compatible bank.
	// It contains an ASCII string of 256 or fewer bytes including one or two terminators of value zero, so as to make
	// the total byte count even.
	// e.g. "General MIDI\0\0"
	Name string // made from the INAM subchunk

	// ROM is an optional field identifying a particular wavetable sound data ROM to which any ROM samples refer.
	// It contains an ASCII string of 256 or fewer bytes including one or two terminators of value zero, so as to make
	// the total byte count even. Both ROM and ROMVer must be present if either is present.
	ROM string // made from the IROM subchunk

	// ROMVer is an optional field identifying the particular wavetable sound data ROM revision to which any
	// ROM samples refer. Both ROM and ROMVer must be present if either is present.
	// e.g. 1.0
	ROMVer struct {
		Major, Minor uint16
	} // made from the IVER subchunk

	// CreationDate is an optional field identifying the creation date of the SoundFont compatible bank.
	// It contains an ASCII string of 256 or fewer bytes including one or two terminators of value zero, so as to make
	// the total byte count even.
	// Conventionally, the format of the string is “Month Day, Year”
	// e.g. "January 1, 2000"
	CreationDate string // made from the ICRD subchunk

	// Engineers is an optional field identifying the engineers who created the SoundFont compatible bank.
	// It contains an ASCII string of 256 or fewer bytes including one or two terminators of value zero, so as to make
	// the total byte count even.
	// e.g. "Jane Doe\0\0"
	Engineers string // made from the IENG subchunk

	// Product is an optional field identifying any specific product for which the SoundFont compatible bank is intended.
	// It contains an ASCII string of 256 or fewer bytes including one or two terminators of value zero, so as to make
	// the total byte count even.
	// e.g. "SBAWE32\0\0"
	Product string // made from the IPRD subchunk

	// Copyright is an optional field containing any copyright assertion string associated with the SoundFont compatible bank.
	// It contains an ASCII string of 256 or fewer bytes including one or two terminators of value zero, so as to make
	// the total byte count even.
	// e.g. "Copyright (c) 1994-95, John Myles White. All rights reserved.\0"
	Copyright string // made from the ICOP subchunk

	// Comments is an optional field containing any comments associated with the SoundFont compatible bank.
	// It contains an ASCII string of 65,536 or fewer bytes including one or two terminators of value zero, so as to make
	// the total byte count even.
	// e.g. "This space unintentionally left blank.\0\0"
	Comments string // made from the ICMT subchunk

	// Software is an optional field identifying the SoundFont compatible tools used to create and most recently
	// modify the SoundFont compatible bank. It contains an ASCII string of 256 or fewer bytes including one or two
	// terminators of value zero, so as to make the total byte count even.
	// e.g. "Sonic Foundry's SoundFont Editor v2.01\0\0"
	Software string // made from the IFST subchunk
}

func (info SoundFontInfo) String() string {
	return fmt.Sprintf("SoundFontInfo{\n\tSfVersion: %d.%d\n\tEngine: %q\n\tName: %q\n\tROM: %q\n\tIVER: %d.%d\n\tCreationDate: %q\n\tEngineers: %q\n\tProduct: %q\n\tCopyright: %q\n\tComments: %q\n\tSoftware: %q\n\t}",
		info.SfVersion.Major,
		info.SfVersion.Minor,
		info.Engine,
		info.Name,
		info.ROM,
		info.ROMVer.Major,
		info.ROMVer.Minor,
		info.CreationDate,
		info.Engineers,
		info.Product,
		info.Copyright,
		info.Comments,
		info.Software)
}

// ReadSoundFontInfo parses a SoundFont info list.
func ReadSoundFontInfo(r io.Reader) (*SoundFontInfo, error) {
	info := &SoundFontInfo{}

	// TODO refactor this out
	// read "INFO" from the "LIST" header
	ok, err := Expect(r, []byte{'I', 'N', 'F', 'O'})
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("expected \"INFO\"")
	}

	// Keep track of known chunks and if we've seen them already
	infoChunks := make(map[[4]byte]bool)
	infoChunks[[4]byte{'i', 'f', 'i', 'l'}] = false
	infoChunks[[4]byte{'i', 's', 'n', 'g'}] = false
	infoChunks[[4]byte{'I', 'N', 'A', 'M'}] = false
	infoChunks[[4]byte{'i', 'r', 'o', 'm'}] = false
	infoChunks[[4]byte{'i', 'v', 'e', 'r'}] = false
	infoChunks[[4]byte{'I', 'C', 'R', 'D'}] = false
	infoChunks[[4]byte{'I', 'E', 'N', 'G'}] = false
	infoChunks[[4]byte{'I', 'P', 'R', 'D'}] = false
	infoChunks[[4]byte{'I', 'C', 'O', 'P'}] = false
	infoChunks[[4]byte{'I', 'C', 'M', 'T'}] = false
	infoChunks[[4]byte{'I', 'S', 'F', 'T'}] = false

	for {
		// parse a chunk
		var chunk chunk
		if err := chunk.parse(r); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		// check if we know how to parse this chunk and if we've seen it already
		seen, ok := infoChunks[chunk.id]
		if !ok {
			// skip unknown chunks
			fmt.Println("unknown chunk", chunk.id)
			continue
		}
		if seen {
			return nil, fmt.Errorf("duplicate chunk %v", chunk.id)
		}
		infoChunks[chunk.id] = true

		// make sense of the chunk
		switch chunk.id {
		case [4]byte{'i', 'f', 'i', 'l'}:
			// must contain 4 bytes
			if chunk.size != 4 {
				return nil, fmt.Errorf("ifil subchunk must contain 4 bytes")
			}

			// first 2 bytes represent the major version number
			info.SfVersion.Major = uint16(chunk.data[1])<<8 | uint16(chunk.data[0])

			// last 2 bytes represent the minor version number
			info.SfVersion.Minor = uint16(chunk.data[3])<<8 | uint16(chunk.data[2])
		case [4]byte{'i', 's', 'n', 'g'}:
			// must contain 256 of fewer bytes
			if chunk.size > 256 {
				return nil, fmt.Errorf("isng subchunk must contain 256 or fewer bytes")
			}

			info.Engine = string(chunk.data)
		case [4]byte{'I', 'N', 'A', 'M'}:
			// must contain 256 of fewer bytes
			if chunk.size > 256 {
				return nil, fmt.Errorf("Inam subchunk must contain 256 or fewer bytes")
			}

			info.Name = string(chunk.data)
		case [4]byte{'i', 'r', 'o', 'm'}:
			// must contain 256 of fewer bytes
			if chunk.size > 256 {
				return nil, fmt.Errorf("irom subchunk must contain 256 or fewer bytes")
			}

			info.ROM = string(chunk.data)
		case [4]byte{'i', 'v', 'e', 'r'}:
			// must contain 4 bytes
			if chunk.size != 4 {
				return nil, fmt.Errorf("iver subchunk must contain 4 bytes")
			}

			// first 2 bytes represent the major version number
			info.ROMVer.Major = uint16(chunk.data[1])<<8 | uint16(chunk.data[0])

			// last 2 bytes represent the minor version number
			info.ROMVer.Minor = uint16(chunk.data[3])<<8 | uint16(chunk.data[2])
		case [4]byte{'I', 'C', 'R', 'D'}:
			// must contain 256 of fewer bytes
			if chunk.size > 256 {
				return nil, fmt.Errorf("ICRD subchunk must contain 256 or fewer bytes")
			}

			info.CreationDate = string(chunk.data)
		case [4]byte{'I', 'E', 'N', 'G'}:
			// must contain 256 of fewer bytes
			if chunk.size > 256 {
				return nil, fmt.Errorf("IENG subchunk must contain 256 or fewer bytes")
			}

			info.Engineers = string(chunk.data)
		case [4]byte{'I', 'P', 'R', 'D'}:
			// must contain 256 of fewer bytes
			if chunk.size > 256 {
				return nil, fmt.Errorf("IPRD subchunk must contain 256 or fewer bytes")
			}

			info.Product = string(chunk.data)
		case [4]byte{'I', 'C', 'O', 'P'}:
			// must contain 256 of fewer bytes
			if chunk.size > 256 {
				return nil, fmt.Errorf("ICOP subchunk must contain 256 or fewer bytes")
			}

			info.Copyright = string(chunk.data)
		case [4]byte{'I', 'C', 'M', 'T'}:
			// must contain 65536 of fewer bytes
			if chunk.size > 65536 {
				return nil, fmt.Errorf("ICMT subchunk must contain 65536 or fewer bytes")
			}

			info.Comments = string(chunk.data)
		case [4]byte{'I', 'S', 'F', 'T'}:
			// must contain 256 of fewer bytes
			if chunk.size > 256 {
				return nil, fmt.Errorf("ISFT subchunk must contain 256 or fewer bytes")
			}

			info.Software = string(chunk.data)
		}
	}

	// If the ifil sub-chunk is missing, or its size is not four bytes, the file should be rejected as structurally unsound.
	if ok := infoChunks[[4]byte{'i', 'f', 'i', 'l'}]; !ok {
		return nil, fmt.Errorf("ifil chunk is missing")
	}

	// If the isng sub-chunk is missing, or is not terminated with a zero valued byte, or its contents are an unknown sound engine,
	// the field should be ignored and EMU8000 assumed.
	if ok := infoChunks[[4]byte{'i', 's', 'n', 'g'}]; !ok {
		info.Engine = "EMU8000"
	}

	return info, nil
}
