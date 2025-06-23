package tables

import "fmt"

func ParseCPAL(src []byte) (CPAL1, error) {
	header, _, err := parseCpal0(src)
	if err != nil {
		return CPAL1{}, err
	}
	switch header.Version {
	case 0:
		return CPAL1{cpal0: header}, nil
	case 1:
		out, _, err := ParseCPAL1(src)
		return out, err
	default:
		return CPAL1{}, fmt.Errorf("unsupported version for CPAL: %d", header.Version)
	}
}

// https://learn.microsoft.com/en-us/typography/opentype/spec/cpal
type cpal0 struct {
	Version            uint16        //	Table version number
	numPaletteEntries  uint16        //	Number of palette entries in each palette.
	numPalettes        uint16        //	Number of palettes in the table.
	numColorRecords    uint16        //	Total number of color records, combined for all palettes.
	ColorRecordsArray  []ColorRecord `arrayCount:"ComputedField-numColorRecords" offsetSize:"Offset32"` //	Offset from the beginning of CPAL table to the first ColorRecord.
	ColorRecordIndices []uint16      `arrayCount:"ComputedField-numPalettes"`                           //[numPalettes]	Index of each palette’s first color record in the combined color record array.
}

type CPAL1 struct {
	cpal0
	paletteTypesArrayOffset       Offset32 // Offset from the beginning of CPAL table to the Palette Types Array. Set to 0 if no array is provided.
	paletteLabelsArrayOffset      Offset32 // Offset from the beginning of CPAL table to the Palette Labels Array. Set to 0 if no array is provided.
	paletteEntryLabelsArrayOffset Offset32 // Offset from the beginning of CPAL table to the Palette Entry Labels Array. Set to 0 if no array is provided.
}

type ColorRecord struct {
	Blue  uint8 // Blue value (B0).
	Green uint8 // Green value (B1).
	Red   uint8 // Red value (B2).
	Alpha uint8 // Alpha value (B3).
}
