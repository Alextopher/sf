package main

import (
	"encoding/binary"
	"fmt"
	"io"
)

type SoundFontHydra struct {
	// Headers is a listing of all presets within the SoundFont compatible file.
	// It always contains a minimum of two records, one record for each preset and one for a terminal record.
	Headers []PresetHeader

	// PBag is a listing of all preset zones within the SoundFont compatible file.
	// It always contains a minimum of two records, one record for each preset and one for a terminal record.
	PBag []struct {
		GenIndex, ModIndex uint16
	}

	// PresetModulators is a listing all preset zone modulators within the SoundFont compatible file.
	PresetModulators []Modulator

	// Generators is a required listing of preset zone generators for each preset zone within the SoundFont compatible file.
	PresetGenerators []Generator

	// Instruments is a required listing of instrument zones for each instrument within the SoundFont.
	Instuments []Instrument

	// IBag is a listing of all instrument zones within the SoundFont compatible file.
	// It contains one record for each instrument zone plus one for a terminal record.
	IBag []struct {
		InstGenIndex, InstModIndex uint16
	}

	// InstrumentModulators is a listing all instrument zone modulators within the SoundFont compatible file.
	InstrumentModulators []Modulator

	// InstrumentGenerators is a required listing of zone generators for each instrument zone within the SoundFont compatible file.
	InstrumentGenerators []Generator

	// Samples is a required listing of all samples within the smpl sub-chunk and any referenced ROM samples.
	Samples []SampleHeader
}

type PresetHeader struct {
	// PresetName contains the name of the preset expressed in ASCII, with unused terminal characters filled with zero valued byte
	PresetName [20]byte
	// Preset contains the MIDI Preset Number
	Preset uint16
	// Bank contains the MIDI Bank Number which apply to this preset.
	// The special case of a General MIDI percussion bank is handled conventionally by a wBank value of 128.
	Bank uint16
	// PresetBagNdx is an index to the preset’s zone list in the PBAG sub-chunk.
	PresetBagNdx uint16
	// Library is reserved for future implementation in a preset library
	// management function and should be preserved as read, and created as zero.
	Library uint32
	// Genre is reserved for future implementation in a preset library
	// management function and should be preserved as read, and created as zero.
	Genre uint32
	// Morphology is reserved for future implementation in a preset library
	// management function and should be preserved as read, and created as zero.
	Morphology uint32
}

func (p PresetHeader) String() string {
	return fmt.Sprintf("PresetHeader{PresetName: %q, Preset: %d, Bank: %d, PresetBagNdx: %d, Library: %d, Genre: %d, Morphology: %d}", p.PresetName, p.Preset, p.Bank, p.PresetBagNdx, p.Library, p.Genre, p.Morphology)
}

type SFModulator uint16
type SFGenerator uint16
type SFTransform uint16

type Modulator struct {
	// ModSrcOper is a value of one of the SFModulator enumeration type values. Unknown or undefined values are
	// ignored. Modulators with sfModAmtSrcOper set to ‘link’ which have no other modulator linked to it are ignored.
	ModSrcOper SFModulator
	// ModDestOper indicates the destination of the modulator. The destination is either a value of one of the SFGenerator
	// enumeration type values or a link to the sfModSrcOper of another modulator block. The latter is indicated by the top bit of
	// the sfModDestOper field being set, the other 15 bits designates the index value of the modulator whose source should be the
	// output of the current modulator RELATIVE TO the first modulator in the instrument zone. Unknown or undefined values
	// are ignored. Modulators with links that point to a modulator index that exceeds the total number of modulators for a given
	// zone are ignored. Linked modulators that are part of circular links are ignored.
	ModDestOper SFGenerator
	// ModAmount is a signed value indicating the degree to which the source modulates the destination. A zero
	// value indicates there is no fixed amount.
	ModAmount int16
	// ModAmtSrcOper is a value of one of the SFModulator enumeration type values. Unknown or undefined values are
	// ignored. Modulators with sfModAmtSrcOper set to ‘link’ are ignored. This value indicates the degree to which the source
	// modulates the destination is to be controlled by the specified modulation source. Note that this enumeration is two bytes in
	// length.
	ModAmtSrcOper SFModulator
	// 	ModTransOper is a value of one of the SFTransform enumeration type values. Unknown or undefined values are
	// ignored. This value indicates that a transform of the specified type will be applied to the modulation source before
	// application to the modulator. Note that this enumeration is two bytes in length
	ModTransOper SFTransform
}

type Generator struct {
	// GenOper is a value of one of the SFGenerator enumeration type values. Unknown or undefined values are
	// ignored.
	GenOper SFGenerator

	// GenAmount is the value to be assigned to the specified generator. Note that this can be of three formats. Certain
	// generators specify a range of MIDI key numbers of MIDI velocities, with a minimum and maximum value. Other
	// generators specify an unsigned WORD value. Most generators, however, specify a signed 16 bit SHORT value.
	GenAmount int16
}

type Instrument struct {
	// Name is the instrument name expressed in ASCII, with unused terminal characters filled with zero valued bytes.
	Name [20]byte
	// InstBagNdx is an index to the instrument’s zone list in the IBAG sub-chunk.
	InstBagNdx uint16
}

func (inst Instrument) String() string {
	return fmt.Sprintf("PresetInstrument{Name: %s, InstBagNdx: %d}", string(inst.Name[:]), inst.InstBagNdx)
}

type SfSampleType uint16

const (
	SampleType_Mono      SfSampleType = 1
	SampleType_Right     SfSampleType = 2
	SampleType_Left      SfSampleType = 4
	SampleType_Link      SfSampleType = 8
	SampleType_Rom_Mono  SfSampleType = 0x8001
	SampleType_Rom_Right SfSampleType = 0x8002
	SampleType_Rom_Left  SfSampleType = 0x8004
	SampleType_Rom_Link  SfSampleType = 0x8008
)

func (s SfSampleType) String() string {
	switch s {
	case SampleType_Mono:
		return "Mono"
	case SampleType_Right:
		return "Right"
	case SampleType_Left:
		return "Left"
	case SampleType_Link:
		return "Link"
	case SampleType_Rom_Mono:
		return "Rom_Mono"
	case SampleType_Rom_Right:
		return "Rom_Right"
	case SampleType_Rom_Left:
		return "Rom_Left"
	case SampleType_Rom_Link:
		return "Rom_Link"
	}
	return fmt.Sprintf("Unknown(%d)", s)
}

type SampleHeader struct {
	// SampleName is the name of the sample expressed in ASCII, with unused terminal characters filled with zero valued bytes.
	SampleName [20]byte
	// Start contains the index, in sample data points, from the beginning of the sample data field to the first data
	// point of this sample.
	Start uint32
	// End contains the index, in sample data points, from the beginning of the sample data field to the first of
	// the set of 46 zero valued data points following this sample.
	End uint32
	// Startloop contains the index, in sample data points, from the beginning of the sample data field to the first
	// data point in the loop of this sample
	Startloop uint32
	// Endloop contains the index, in sample data points, from the beginning of the sample data field to the first
	// data point following the loop of this sample. Note that this is the data point “equivalent to” the first loop data point, and that
	// to produce portable artifact free loops, the eight proximal data points surrounding both the Startloop and Endloop points
	// should be identical.
	Endloop uint32
	// SampleRate contains the sample rate, in hertz, at which this sample was acquired or to which it was most recently converted
	SampleRate uint32
	// OriginalPitch contains the MIDI key number of the recorded pitch of the sample.
	// Values between 128 and 254 are illegal. Whenever an illegal value or a value of 255 is encountered, the value 60 should be used
	OriginalPitch uint8
	// PitchCorrection contains a pitch correction in cents that should be applied to the sample on playback. The
	// purpose of this field is to compensate for any pitch errors during the sample recording process. The correction value is that
	// of the correction to be applied.
	PitchCorrection int8
	// TODO SampleLink
	SampleLink uint16
	// SampleType is a value of one of the SampleType enumeration type values.
	SampleType SfSampleType
}

func (s SampleHeader) String() string {
	return fmt.Sprintf("SampleHeader{SampleName: %s, Start: %d, End: %d, Startloop: %d, Endloop: %d, SampleRate: %d, OriginalPitch: %d, PitchCorrection: %d, SampleLink: %d, SampleType: %v}",
		string(s.SampleName[:]),
		s.Start,
		s.End,
		s.Startloop,
		s.Endloop,
		s.SampleRate,
		s.OriginalPitch,
		s.PitchCorrection,
		s.SampleLink,
		s.SampleType)
}

func ReadSoundFontHydra(r io.Reader) (*SoundFontHydra, error) {
	sound := &SoundFontHydra{}

	pdtaChunks := make(map[[4]byte]bool)
	pdtaChunks[[4]byte{'p', 'h', 'd', 'r'}] = false
	pdtaChunks[[4]byte{'p', 'b', 'a', 'g'}] = false
	pdtaChunks[[4]byte{'p', 'm', 'o', 'd'}] = false
	pdtaChunks[[4]byte{'p', 'g', 'e', 'n'}] = false
	pdtaChunks[[4]byte{'i', 'n', 's', 't'}] = false
	pdtaChunks[[4]byte{'i', 'b', 'a', 'g'}] = false
	pdtaChunks[[4]byte{'i', 'm', 'o', 'd'}] = false
	pdtaChunks[[4]byte{'i', 'g', 'e', 'n'}] = false
	pdtaChunks[[4]byte{'s', 'h', 'd', 'r'}] = false

	for {
		// parse a chunk
		var chunk chunk
		if err := chunk.parse(r); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		_, ok := pdtaChunks[chunk.id]
		if !ok {
			// skip unknown chunks
			fmt.Println("unknown chunk", string(chunk.id[:]))
			continue
		}
		pdtaChunks[chunk.id] = true
		fmt.Println("found chunk", string(chunk.id[:]))

		// make sense of the chunk
		switch chunk.id {
		case [4]byte{'p', 'h', 'd', 'r'}:
			// each preset header is 38 bytes long
			if chunk.size%38 != 0 {
				return nil, fmt.Errorf("invalid preset header size %d", chunk.size)
			}
			sound.Headers = make([]PresetHeader, chunk.size/38)

			chunkReader := chunk.newReader()
			for i := 0; i < len(sound.Headers); i++ {
				if err := binary.Read(chunkReader, binary.LittleEndian, &sound.Headers[i]); err != nil {
					return nil, err
				}
			}
		case [4]byte{'p', 'b', 'a', 'g'}:
			// each preset bag is 4 bytes long
			if chunk.size%4 != 0 {
				return nil, fmt.Errorf("invalid preset bag size %d", chunk.size)
			}
			sound.PBag = make([]struct {
				GenIndex, ModIndex uint16
			}, chunk.size/4)

			for i := 0; i < len(sound.PBag); i++ {
				// first 2 bytes represent the major version number
				sound.PBag[i].GenIndex = uint16(chunk.data[4*i+1])<<8 | uint16(chunk.data[4*i])

				// last 2 bytes represent the minor version number
				sound.PBag[i].ModIndex = uint16(chunk.data[4*i+3])<<8 | uint16(chunk.data[4*i+2])
			}
		case [4]byte{'p', 'm', 'o', 'd'}:
			// each preset modulator is 10 bytes long
			if chunk.size%10 != 0 {
				return nil, fmt.Errorf("invalid preset modulator size %d", chunk.size)
			}
			sound.PresetModulators = make([]Modulator, chunk.size/10)

			chunkReader := chunk.newReader()
			for i := 0; i < len(sound.PresetModulators); i++ {
				if err := binary.Read(chunkReader, binary.LittleEndian, &sound.PresetModulators[i]); err != nil {
					return nil, err
				}
			}
		case [4]byte{'p', 'g', 'e', 'n'}:
			// each preset generator is 4 bytes long
			if chunk.size%4 != 0 {
				return nil, fmt.Errorf("invalid preset generator size %d", chunk.size)
			}
			sound.PresetGenerators = make([]Generator, chunk.size/4)

			chunkReader := chunk.newReader()
			for i := 0; i < len(sound.PresetGenerators); i++ {
				if err := binary.Read(chunkReader, binary.LittleEndian, &sound.PresetGenerators[i]); err != nil {
					return nil, err
				}
			}
		case [4]byte{'i', 'n', 's', 't'}:
			// each instrument header is 22 bytes long
			if chunk.size%22 != 0 {
				return nil, fmt.Errorf("invalid instrument header size %d", chunk.size)
			}
			sound.Instuments = make([]Instrument, chunk.size/22)

			chunkReader := chunk.newReader()
			for i := 0; i < len(sound.Instuments); i++ {
				if err := binary.Read(chunkReader, binary.LittleEndian, &sound.Instuments[i]); err != nil {
					return nil, err
				}
			}
		case [4]byte{'i', 'b', 'a', 'g'}:
			// each instrument bag is 4 bytes long
			if chunk.size%4 != 0 {
				return nil, fmt.Errorf("invalid preset bag size %d", chunk.size)
			}
			sound.IBag = make([]struct {
				InstGenIndex, InstModIndex uint16
			}, chunk.size/4)

			for i := 0; i < len(sound.IBag); i++ {
				// first 2 bytes represent the major version number
				sound.IBag[i].InstGenIndex = uint16(chunk.data[4*i+1])<<8 | uint16(chunk.data[4*i])

				// last 2 bytes represent the minor version number
				sound.IBag[i].InstModIndex = uint16(chunk.data[4*i+3])<<8 | uint16(chunk.data[4*i+2])
			}
		case [4]byte{'i', 'm', 'o', 'd'}:
			// each preset modulator is 10 bytes long
			if chunk.size%10 != 0 {
				return nil, fmt.Errorf("invalid preset modulator size %d", chunk.size)
			}
			sound.InstrumentModulators = make([]Modulator, chunk.size/10)

			chunkReader := chunk.newReader()
			for i := 0; i < len(sound.InstrumentModulators); i++ {
				if err := binary.Read(chunkReader, binary.LittleEndian, &sound.InstrumentModulators[i]); err != nil {
					return nil, err
				}
			}
		case [4]byte{'i', 'g', 'e', 'n'}:
			// each preset generator is 4 bytes long
			if chunk.size%4 != 0 {
				return nil, fmt.Errorf("invalid preset generator size %d", chunk.size)
			}
			sound.InstrumentGenerators = make([]Generator, chunk.size/4)

			chunkReader := chunk.newReader()
			for i := 0; i < len(sound.InstrumentGenerators); i++ {
				if err := binary.Read(chunkReader, binary.LittleEndian, &sound.InstrumentGenerators[i]); err != nil {
					return nil, err
				}
			}
		case [4]byte{'s', 'h', 'd', 'r'}:
			// each sample header is 46 bytes long
			if chunk.size%46 != 0 {
				return nil, fmt.Errorf("invalid sample header size %d", chunk.size)
			}
			sound.Samples = make([]SampleHeader, chunk.size/46)

			chunkReader := chunk.newReader()
			for i := 0; i < len(sound.Samples); i++ {
				if err := binary.Read(chunkReader, binary.LittleEndian, &sound.Samples[i]); err != nil {
					return nil, err
				}
			}
		}
	}

	// All chunks must be present
	for ck, ok := range pdtaChunks {
		if !ok {
			return nil, fmt.Errorf("missing chunk %v", string(ck[:]))
		}

	}

	return sound, nil
}
