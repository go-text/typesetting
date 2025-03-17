// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package tables

import (
	"encoding/binary"
	"fmt"
)

// Code generated by binarygen from xvar_src.go. DO NOT EDIT

func (item *AxisValueMap) mustParse(src []byte) {
	_ = src[3] // early bound checking
	item.FromCoordinate = Coord(binary.BigEndian.Uint16(src[0:]))
	item.ToCoordinate = Coord(binary.BigEndian.Uint16(src[2:]))
}

func ParseAvar(src []byte) (Avar, int, error) {
	var item Avar
	n := 0
	if L := len(src); L < 8 {
		return item, 0, fmt.Errorf("reading Avar: "+"EOF: expected length: 8, got %d", L)
	}
	_ = src[7] // early bound checking
	item.majorVersion = binary.BigEndian.Uint16(src[0:])
	item.minorVersion = binary.BigEndian.Uint16(src[2:])
	item.reserved = binary.BigEndian.Uint16(src[4:])
	arrayLengthAxisSegmentMaps := int(binary.BigEndian.Uint16(src[6:]))
	n += 8

	{

		offset := 8
		for i := 0; i < arrayLengthAxisSegmentMaps; i++ {
			elem, read, err := ParseSegmentMaps(src[offset:])
			if err != nil {
				return item, 0, fmt.Errorf("reading Avar: %s", err)
			}
			item.AxisSegmentMaps = append(item.AxisSegmentMaps, elem)
			offset += read
		}
		n = offset
	}
	return item, n, nil
}

func ParseDeltaSetMapping(src []byte) (DeltaSetMapping, int, error) {
	var item DeltaSetMapping
	n := 0
	if L := len(src); L < 2 {
		return item, 0, fmt.Errorf("reading DeltaSetMapping: "+"EOF: expected length: 2, got %d", L)
	}
	_ = src[1] // early bound checking
	item.format = src[0]
	item.entryFormat = src[1]
	n += 2

	{

		err := item.parseMap(src[2:])
		if err != nil {
			return item, 0, fmt.Errorf("reading DeltaSetMapping: %s", err)
		}
	}
	return item, n, nil
}

func ParseFvar(src []byte) (Fvar, int, error) {
	var item Fvar
	n := 0
	if L := len(src); L < 16 {
		return item, 0, fmt.Errorf("reading Fvar: "+"EOF: expected length: 16, got %d", L)
	}
	_ = src[15] // early bound checking
	item.majorVersion = binary.BigEndian.Uint16(src[0:])
	item.minorVersion = binary.BigEndian.Uint16(src[2:])
	item.axesArrayOffset = Offset16(binary.BigEndian.Uint16(src[4:]))
	item.reserved = binary.BigEndian.Uint16(src[6:])
	item.axisCount = binary.BigEndian.Uint16(src[8:])
	item.axisSize = binary.BigEndian.Uint16(src[10:])
	item.instanceCount = binary.BigEndian.Uint16(src[12:])
	item.instanceSize = binary.BigEndian.Uint16(src[14:])
	n += 16

	{

		err := item.parseFvarRecords(src[:])
		if err != nil {
			return item, 0, fmt.Errorf("reading Fvar: %s", err)
		}
	}
	return item, n, nil
}

func ParseFvarRecords(src []byte, axisCount int, instanceCount int, instanceSize int) (FvarRecords, int, error) {
	var item FvarRecords
	n := 0
	{

		if L := len(src); L < axisCount*20 {
			return item, 0, fmt.Errorf("reading FvarRecords: "+"EOF: expected length: %d, got %d", axisCount*20, L)
		}

		item.Axis = make([]VariationAxisRecord, axisCount) // allocation guarded by the previous check
		for i := range item.Axis {
			item.Axis[i].mustParse(src[i*20:])
		}
		n += axisCount * 20
	}
	{

		err := item.parseInstances(src[n:], axisCount, instanceCount, instanceSize)
		if err != nil {
			return item, 0, fmt.Errorf("reading FvarRecords: %s", err)
		}
	}
	return item, n, nil
}

func ParseGlyphVariationData(src []byte, axisCount int) (GlyphVariationData, int, error) {
	var item GlyphVariationData
	n := 0
	if L := len(src); L < 4 {
		return item, 0, fmt.Errorf("reading GlyphVariationData: "+"EOF: expected length: 4, got %d", L)
	}
	_ = src[3] // early bound checking
	item.tupleVariationCount = binary.BigEndian.Uint16(src[0:])
	offsetSerializedData := int(binary.BigEndian.Uint16(src[2:]))
	n += 4

	{

		if offsetSerializedData != 0 { // ignore null offset
			if L := len(src); L < offsetSerializedData {
				return item, 0, fmt.Errorf("reading GlyphVariationData: "+"EOF: expected length: %d, got %d", offsetSerializedData, L)
			}

			item.SerializedData = src[offsetSerializedData:]
		}
	}
	{
		arrayLength := int(item.tupleVariationCount & 0x0FFF)

		offset := 4
		for i := 0; i < arrayLength; i++ {
			elem, read, err := ParseTupleVariationHeader(src[offset:], axisCount)
			if err != nil {
				return item, 0, fmt.Errorf("reading GlyphVariationData: %s", err)
			}
			item.TupleVariationHeaders = append(item.TupleVariationHeaders, elem)
			offset += read
		}
		n = offset
	}
	return item, n, nil
}

func ParseGvar(src []byte) (Gvar, int, error) {
	var item Gvar
	n := 0
	if L := len(src); L < 20 {
		return item, 0, fmt.Errorf("reading Gvar: "+"EOF: expected length: 20, got %d", L)
	}
	_ = src[19] // early bound checking
	item.majorVersion = binary.BigEndian.Uint16(src[0:])
	item.minorVersion = binary.BigEndian.Uint16(src[2:])
	item.axisCount = binary.BigEndian.Uint16(src[4:])
	item.sharedTupleCount = binary.BigEndian.Uint16(src[6:])
	offsetSharedTuples := int(binary.BigEndian.Uint32(src[8:]))
	item.glyphCount = binary.BigEndian.Uint16(src[12:])
	item.flags = binary.BigEndian.Uint16(src[14:])
	item.glyphVariationDataArrayOffset = Offset32(binary.BigEndian.Uint32(src[16:]))
	n += 20

	{

		if offsetSharedTuples != 0 { // ignore null offset
			if L := len(src); L < offsetSharedTuples {
				return item, 0, fmt.Errorf("reading Gvar: "+"EOF: expected length: %d, got %d", offsetSharedTuples, L)
			}

			var err error
			item.SharedTuples, _, err = ParseSharedTuples(src[offsetSharedTuples:], int(item.sharedTupleCount), int(item.axisCount))
			if err != nil {
				return item, 0, fmt.Errorf("reading Gvar: %s", err)
			}

		}
	}
	{

		err := item.parseGlyphVariationDataOffsets(src[20:])
		if err != nil {
			return item, 0, fmt.Errorf("reading Gvar: %s", err)
		}
	}
	{

		err := item.parseGlyphVariationDatas(src[:])
		if err != nil {
			return item, 0, fmt.Errorf("reading Gvar: %s", err)
		}
	}
	return item, n, nil
}

func ParseHVAR(src []byte) (HVAR, int, error) {
	var item HVAR
	n := 0
	if L := len(src); L < 20 {
		return item, 0, fmt.Errorf("reading HVAR: "+"EOF: expected length: 20, got %d", L)
	}
	_ = src[19] // early bound checking
	item.majorVersion = binary.BigEndian.Uint16(src[0:])
	item.minorVersion = binary.BigEndian.Uint16(src[2:])
	offsetItemVariationStore := int(binary.BigEndian.Uint32(src[4:]))
	offsetAdvanceWidthMapping := int(binary.BigEndian.Uint32(src[8:]))
	offsetLsbMapping := int(binary.BigEndian.Uint32(src[12:]))
	offsetRsbMapping := int(binary.BigEndian.Uint32(src[16:]))
	n += 20

	{

		if offsetItemVariationStore != 0 { // ignore null offset
			if L := len(src); L < offsetItemVariationStore {
				return item, 0, fmt.Errorf("reading HVAR: "+"EOF: expected length: %d, got %d", offsetItemVariationStore, L)
			}

			var err error
			item.ItemVariationStore, _, err = ParseItemVarStore(src[offsetItemVariationStore:])
			if err != nil {
				return item, 0, fmt.Errorf("reading HVAR: %s", err)
			}

		}
	}
	{

		if offsetAdvanceWidthMapping != 0 { // ignore null offset
			if L := len(src); L < offsetAdvanceWidthMapping {
				return item, 0, fmt.Errorf("reading HVAR: "+"EOF: expected length: %d, got %d", offsetAdvanceWidthMapping, L)
			}

			var err error
			item.AdvanceWidthMapping, _, err = ParseDeltaSetMapping(src[offsetAdvanceWidthMapping:])
			if err != nil {
				return item, 0, fmt.Errorf("reading HVAR: %s", err)
			}

		}
	}
	{

		if offsetLsbMapping != 0 { // ignore null offset
			if L := len(src); L < offsetLsbMapping {
				return item, 0, fmt.Errorf("reading HVAR: "+"EOF: expected length: %d, got %d", offsetLsbMapping, L)
			}

			var tmpLsbMapping DeltaSetMapping
			var err error
			tmpLsbMapping, _, err = ParseDeltaSetMapping(src[offsetLsbMapping:])
			if err != nil {
				return item, 0, fmt.Errorf("reading HVAR: %s", err)
			}

			item.LsbMapping = &tmpLsbMapping
		}
	}
	{

		if offsetRsbMapping != 0 { // ignore null offset
			if L := len(src); L < offsetRsbMapping {
				return item, 0, fmt.Errorf("reading HVAR: "+"EOF: expected length: %d, got %d", offsetRsbMapping, L)
			}

			var tmpRsbMapping DeltaSetMapping
			var err error
			tmpRsbMapping, _, err = ParseDeltaSetMapping(src[offsetRsbMapping:])
			if err != nil {
				return item, 0, fmt.Errorf("reading HVAR: %s", err)
			}

			item.RsbMapping = &tmpRsbMapping
		}
	}
	return item, n, nil
}

func ParseInstanceRecord(src []byte, coordinatesCount int) (InstanceRecord, int, error) {
	var item InstanceRecord
	n := 0
	if L := len(src); L < 4 {
		return item, 0, fmt.Errorf("reading InstanceRecord: "+"EOF: expected length: 4, got %d", L)
	}
	_ = src[3] // early bound checking
	item.SubfamilyNameID = NameID(binary.BigEndian.Uint16(src[0:]))
	item.flags = binary.BigEndian.Uint16(src[2:])
	n += 4

	{

		if L := len(src); L < 4+coordinatesCount*4 {
			return item, 0, fmt.Errorf("reading InstanceRecord: "+"EOF: expected length: %d, got %d", 4+coordinatesCount*4, L)
		}

		item.Coordinates = make([]float32, coordinatesCount) // allocation guarded by the previous check
		for i := range item.Coordinates {
			item.Coordinates[i] = Float1616FromUint(binary.BigEndian.Uint32(src[4+i*4:]))
		}
		n += coordinatesCount * 4
	}
	{

		read, err := item.parsePostScriptNameID(src[n:], coordinatesCount)
		if err != nil {
			return item, 0, fmt.Errorf("reading InstanceRecord: %s", err)
		}
		n += read
	}
	return item, n, nil
}

func ParseItemVarStore(src []byte) (ItemVarStore, int, error) {
	var item ItemVarStore
	n := 0
	if L := len(src); L < 8 {
		return item, 0, fmt.Errorf("reading ItemVarStore: "+"EOF: expected length: 8, got %d", L)
	}
	_ = src[7] // early bound checking
	item.format = binary.BigEndian.Uint16(src[0:])
	offsetVariationRegionList := int(binary.BigEndian.Uint32(src[2:]))
	arrayLengthItemVariationDatas := int(binary.BigEndian.Uint16(src[6:]))
	n += 8

	{

		if offsetVariationRegionList != 0 { // ignore null offset
			if L := len(src); L < offsetVariationRegionList {
				return item, 0, fmt.Errorf("reading ItemVarStore: "+"EOF: expected length: %d, got %d", offsetVariationRegionList, L)
			}

			var err error
			item.VariationRegionList, _, err = ParseVariationRegionList(src[offsetVariationRegionList:])
			if err != nil {
				return item, 0, fmt.Errorf("reading ItemVarStore: %s", err)
			}

		}
	}
	{

		if L := len(src); L < 8+arrayLengthItemVariationDatas*4 {
			return item, 0, fmt.Errorf("reading ItemVarStore: "+"EOF: expected length: %d, got %d", 8+arrayLengthItemVariationDatas*4, L)
		}

		item.ItemVariationDatas = make([]ItemVariationData, arrayLengthItemVariationDatas) // allocation guarded by the previous check
		for i := range item.ItemVariationDatas {
			offset := int(binary.BigEndian.Uint32(src[8+i*4:]))
			// ignore null offsets
			if offset == 0 {
				continue
			}

			if L := len(src); L < offset {
				return item, 0, fmt.Errorf("reading ItemVarStore: "+"EOF: expected length: %d, got %d", offset, L)
			}

			var err error
			item.ItemVariationDatas[i], _, err = ParseItemVariationData(src[offset:])
			if err != nil {
				return item, 0, fmt.Errorf("reading ItemVarStore: %s", err)
			}
		}
		n += arrayLengthItemVariationDatas * 4
	}
	return item, n, nil
}

func ParseItemVariationData(src []byte) (ItemVariationData, int, error) {
	var item ItemVariationData
	n := 0
	if L := len(src); L < 6 {
		return item, 0, fmt.Errorf("reading ItemVariationData: "+"EOF: expected length: 6, got %d", L)
	}
	_ = src[5] // early bound checking
	item.itemCount = binary.BigEndian.Uint16(src[0:])
	item.wordDeltaCount = binary.BigEndian.Uint16(src[2:])
	item.regionIndexCount = binary.BigEndian.Uint16(src[4:])
	n += 6

	{
		arrayLength := int(item.regionIndexCount)

		if L := len(src); L < 6+arrayLength*2 {
			return item, 0, fmt.Errorf("reading ItemVariationData: "+"EOF: expected length: %d, got %d", 6+arrayLength*2, L)
		}

		item.RegionIndexes = make([]uint16, arrayLength) // allocation guarded by the previous check
		for i := range item.RegionIndexes {
			item.RegionIndexes[i] = binary.BigEndian.Uint16(src[6+i*2:])
		}
		n += arrayLength * 2
	}
	{

		err := item.parseDeltaSets(src[n:])
		if err != nil {
			return item, 0, fmt.Errorf("reading ItemVariationData: %s", err)
		}
	}
	return item, n, nil
}

func ParseMVAR(src []byte) (MVAR, int, error) {
	var item MVAR
	n := 0
	if L := len(src); L < 12 {
		return item, 0, fmt.Errorf("reading MVAR: "+"EOF: expected length: 12, got %d", L)
	}
	_ = src[11] // early bound checking
	item.majorVersion = binary.BigEndian.Uint16(src[0:])
	item.minorVersion = binary.BigEndian.Uint16(src[2:])
	item.reserved = binary.BigEndian.Uint16(src[4:])
	item.valueRecordSize = binary.BigEndian.Uint16(src[6:])
	item.valueRecordCount = binary.BigEndian.Uint16(src[8:])
	offsetItemVariationStore := int(binary.BigEndian.Uint16(src[10:]))
	n += 12

	{

		if offsetItemVariationStore != 0 { // ignore null offset
			if L := len(src); L < offsetItemVariationStore {
				return item, 0, fmt.Errorf("reading MVAR: "+"EOF: expected length: %d, got %d", offsetItemVariationStore, L)
			}

			var err error
			item.ItemVariationStore, _, err = ParseItemVarStore(src[offsetItemVariationStore:])
			if err != nil {
				return item, 0, fmt.Errorf("reading MVAR: %s", err)
			}

		}
	}
	{

		err := item.parseValueRecords(src[12:])
		if err != nil {
			return item, 0, fmt.Errorf("reading MVAR: %s", err)
		}
	}
	return item, n, nil
}

func ParseSegmentMaps(src []byte) (SegmentMaps, int, error) {
	var item SegmentMaps
	n := 0
	if L := len(src); L < 2 {
		return item, 0, fmt.Errorf("reading SegmentMaps: "+"EOF: expected length: 2, got %d", L)
	}
	arrayLengthAxisValueMaps := int(binary.BigEndian.Uint16(src[0:]))
	n += 2

	{

		if L := len(src); L < 2+arrayLengthAxisValueMaps*4 {
			return item, 0, fmt.Errorf("reading SegmentMaps: "+"EOF: expected length: %d, got %d", 2+arrayLengthAxisValueMaps*4, L)
		}

		item.AxisValueMaps = make([]AxisValueMap, arrayLengthAxisValueMaps) // allocation guarded by the previous check
		for i := range item.AxisValueMaps {
			item.AxisValueMaps[i].mustParse(src[2+i*4:])
		}
		n += arrayLengthAxisValueMaps * 4
	}
	return item, n, nil
}

func ParseSharedTuples(src []byte, sharedTuplesCount int, valuesCount int) (SharedTuples, int, error) {
	var item SharedTuples
	n := 0
	{

		offset := 0
		for i := 0; i < sharedTuplesCount; i++ {
			elem, read, err := ParseTuple(src[offset:], valuesCount)
			if err != nil {
				return item, 0, fmt.Errorf("reading SharedTuples: %s", err)
			}
			item.SharedTuples = append(item.SharedTuples, elem)
			offset += read
		}
		n = offset
	}
	return item, n, nil
}

func ParseTuple(src []byte, valuesCount int) (Tuple, int, error) {
	var item Tuple
	n := 0
	{

		if L := len(src); L < valuesCount*2 {
			return item, 0, fmt.Errorf("reading Tuple: "+"EOF: expected length: %d, got %d", valuesCount*2, L)
		}

		item.Values = make([]Coord, valuesCount) // allocation guarded by the previous check
		for i := range item.Values {
			item.Values[i] = Coord(binary.BigEndian.Uint16(src[i*2:]))
		}
		n += valuesCount * 2
	}
	return item, n, nil
}

func ParseTupleVariationHeader(src []byte, axisCount int) (TupleVariationHeader, int, error) {
	var item TupleVariationHeader
	n := 0
	if L := len(src); L < 4 {
		return item, 0, fmt.Errorf("reading TupleVariationHeader: "+"EOF: expected length: 4, got %d", L)
	}
	_ = src[3] // early bound checking
	item.VariationDataSize = binary.BigEndian.Uint16(src[0:])
	item.tupleIndex = binary.BigEndian.Uint16(src[2:])
	n += 4

	{

		read, err := item.parsePeakTuple(src[4:], axisCount)
		if err != nil {
			return item, 0, fmt.Errorf("reading TupleVariationHeader: %s", err)
		}
		n += read
	}
	{

		read, err := item.parseIntermediateTuples(src[n:], axisCount)
		if err != nil {
			return item, 0, fmt.Errorf("reading TupleVariationHeader: %s", err)
		}
		n += read
	}
	return item, n, nil
}

func ParseVarValueRecord(src []byte) (VarValueRecord, int, error) {
	var item VarValueRecord
	n := 0
	if L := len(src); L < 8 {
		return item, 0, fmt.Errorf("reading VarValueRecord: "+"EOF: expected length: 8, got %d", L)
	}
	item.mustParse(src)
	n += 8
	return item, n, nil
}

func ParseVariationRegion(src []byte, regionAxesCount int) (VariationRegion, int, error) {
	var item VariationRegion
	n := 0
	{

		if L := len(src); L < regionAxesCount*6 {
			return item, 0, fmt.Errorf("reading VariationRegion: "+"EOF: expected length: %d, got %d", regionAxesCount*6, L)
		}

		item.RegionAxes = make([]RegionAxisCoordinates, regionAxesCount) // allocation guarded by the previous check
		for i := range item.RegionAxes {
			item.RegionAxes[i].mustParse(src[i*6:])
		}
		n += regionAxesCount * 6
	}
	return item, n, nil
}

func ParseVariationRegionList(src []byte) (VariationRegionList, int, error) {
	var item VariationRegionList
	n := 0
	if L := len(src); L < 4 {
		return item, 0, fmt.Errorf("reading VariationRegionList: "+"EOF: expected length: 4, got %d", L)
	}
	_ = src[3] // early bound checking
	item.axisCount = binary.BigEndian.Uint16(src[0:])
	arrayLengthVariationRegions := int(binary.BigEndian.Uint16(src[2:]))
	n += 4

	{

		offset := 4
		for i := 0; i < arrayLengthVariationRegions; i++ {
			elem, read, err := ParseVariationRegion(src[offset:], int(item.axisCount))
			if err != nil {
				return item, 0, fmt.Errorf("reading VariationRegionList: %s", err)
			}
			item.VariationRegions = append(item.VariationRegions, elem)
			offset += read
		}
		n = offset
	}
	return item, n, nil
}

func (item *RegionAxisCoordinates) mustParse(src []byte) {
	_ = src[5] // early bound checking
	item.StartCoord = Coord(binary.BigEndian.Uint16(src[0:]))
	item.PeakCoord = Coord(binary.BigEndian.Uint16(src[2:]))
	item.EndCoord = Coord(binary.BigEndian.Uint16(src[4:]))
}

func (item *VarValueRecord) mustParse(src []byte) {
	_ = src[7] // early bound checking
	item.ValueTag = Tag(binary.BigEndian.Uint32(src[0:]))
	item.Index.mustParse(src[4:])
}

func (item *VariationAxisRecord) mustParse(src []byte) {
	_ = src[19] // early bound checking
	item.Tag = Tag(binary.BigEndian.Uint32(src[0:]))
	item.Minimum = Float1616FromUint(binary.BigEndian.Uint32(src[4:]))
	item.Default = Float1616FromUint(binary.BigEndian.Uint32(src[8:]))
	item.Maximum = Float1616FromUint(binary.BigEndian.Uint32(src[12:]))
	item.flags = binary.BigEndian.Uint16(src[16:])
	item.strid = NameID(binary.BigEndian.Uint16(src[18:]))
}

func (item *VariationStoreIndex) mustParse(src []byte) {
	_ = src[3] // early bound checking
	item.DeltaSetOuter = binary.BigEndian.Uint16(src[0:])
	item.DeltaSetInner = binary.BigEndian.Uint16(src[2:])
}
