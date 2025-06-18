package tables

import "fmt"

func ParseCOLR(src []byte) (COLR1, error) {
	header, _, err := parseColr0(src)
	if err != nil {
		return COLR1{}, err
	}
	switch header.Version {
	case 0:
		return COLR1{colr0: header}, nil
	case 1:
		out, _, err := ParseCOLR1(src)
		return out, err
	default:
		return COLR1{}, fmt.Errorf("unsupported version for COLR: %d", header.Version)
	}
}

// https://learn.microsoft.com/en-us/typography/opentype/spec/colr#colr-table-formats
type colr0 struct {
	Version             uint16      // Table version number
	numBaseGlyphRecords uint16      // Number of BaseGlyph records.
	BaseGlyphRecords    []BaseGlyph `arrayCount:"ComputedField-numBaseGlyphRecords" offsetSize:"Offset32"` // Offset to baseGlyphRecords array, from beginning of COLR table.
	LayerRecords        []Layer     `arrayCount:"ComputedField-numLayerRecords" offsetSize:"Offset32"`     // Offset to layerRecords array, from beginning of COLR table.
	numLayerRecords     uint16      // Number of Layer records.
}

type COLR1 struct {
	colr0
	BaseGlyphList      BaseGlyphList    `offsetSize:"Offset32"` // Offset to BaseGlyphList table, from beginning of COLR table.
	LayerList          LayerList        `offsetSize:"Offset32"` // Offset to LayerList table, from beginning of COLR table (may be NULL).
	ClipList           ClipList         `offsetSize:"Offset32"` // Offset to ClipList table, from beginning of COLR table (may be NULL).
	VarIndexMap        *DeltaSetMapping `offsetSize:"Offset32"` // Offset to DeltaSetIndexMap table, from beginning of COLR table (may be NULL).
	ItemVariationStore *ItemVarStore    `offsetSize:"Offset32"` // Offset to ItemVariationStore, from beginning of COLR table (may be NULL).
}

type BaseGlyph struct {
	GlyphID         GlyphID // Glyph ID of the base glyph.
	FirstLayerIndex uint16  // Index (base 0) into the layerRecords array.
	NumLayers       uint16  // Number of color layers associated with this glyph.
}

type Layer struct {
	GlyphID      GlyphID // Glyph ID of the glyph used for a given layer.
	PaletteIndex uint16  // Index (base 0) for a palette entry in the CPAL table.
}

type BaseGlyphList struct {
	PaintRecords []BaseGlyphPaintRecord `arrayCount:"FirstUint32"` // numBaseGlyphPaintRecords
}

type BaseGlyphPaintRecord struct {
	GlyphID GlyphID    // Glyph ID of the base glyph.
	Paint   PaintTable `offsetSize:"Offset32" offsetRelativeTo:"Parent"` // Offset to a Paint table, from beginning of BaseGlyphList table.
}

type LayerList struct {
	PaintTables []PaintTable `arrayCount:"FirstUint32" offsetsArray:"Offset32"` // Offsets to Paint tables, from beginning of LayerList table.
}

type ClipList struct {
	format uint8  // Set to 1.
	Clips  []Clip `arrayCount:"FirstUint32"` // Clip records. Sorted by startGlyphID.
}

type Clip struct {
	StartGlyphID GlyphID // First glyph ID in the range.
	EndGlyphID   GlyphID // Last glyph ID in the range.
	ClipBox      ClipBox `offsetSize:"Offset24" offsetRelativeTo:"Parent"` // Offset to a ClipBox table, from beginning of ClipList table.
}

type ClipBox interface {
	isClipBox()
}

func (ClipBoxFormat1) isClipBox() {}
func (ClipBoxFormat2) isClipBox() {}

// static clip box
type ClipBoxFormat1 struct {
	format byte  `unionTag:"1"`
	XMin   int16 //	Minimum x of clip box.
	YMin   int16 //	Minimum y of clip box.
	XMax   int16 //	Maximum x of clip box.
	YMax   int16 //	Maximum y of clip box.
}

// variable clip box
type ClipBoxFormat2 struct {
	format       byte   `unionTag:"2"`
	XMin         int16  // 	Minimum x of clip box. For variation, use varIndexBase + 0.
	YMin         int16  // 	Minimum y of clip box. For variation, use varIndexBase + 1.
	XMax         int16  // 	Maximum x of clip box. For variation, use varIndexBase + 2.
	YMax         int16  // 	Maximum y of clip box. For variation, use varIndexBase + 3.
	VarIndexBase uint32 // 	Base index into DeltaSetIndexMap.
}

type PaintTable interface {
	isPaintTable()
}

func (PaintColrLayers) isPaintTable()                  {}
func (PaintSolid) isPaintTable()                       {}
func (PaintVarSolid) isPaintTable()                    {}
func (PaintLinearGradient) isPaintTable()              {}
func (PaintVarLinearGradient) isPaintTable()           {}
func (PaintRadialGradient) isPaintTable()              {}
func (PaintVarRadialGradient) isPaintTable()           {}
func (PaintSweepGradient) isPaintTable()               {}
func (PaintVarSweepGradient) isPaintTable()            {}
func (PaintGlyph) isPaintTable()                       {}
func (PaintColrGlyph) isPaintTable()                   {}
func (PaintTransform) isPaintTable()                   {}
func (PaintVarTransform) isPaintTable()                {}
func (PaintTranslate) isPaintTable()                   {}
func (PaintVarTranslate) isPaintTable()                {}
func (PaintScale) isPaintTable()                       {}
func (PaintVarScale) isPaintTable()                    {}
func (PaintScaleAroundCenter) isPaintTable()           {}
func (PaintVarScaleAroundCenter) isPaintTable()        {}
func (PaintScaleUniform) isPaintTable()                {}
func (PaintVarScaleUniform) isPaintTable()             {}
func (PaintScaleUniformAroundCenter) isPaintTable()    {}
func (PaintVarScaleUniformAroundCenter) isPaintTable() {}
func (PaintRotate) isPaintTable()                      {}
func (PaintVarRotate) isPaintTable()                   {}
func (PaintRotateAroundCenter) isPaintTable()          {}
func (PaintVarRotateAroundCenter) isPaintTable()       {}
func (PaintSkew) isPaintTable()                        {}
func (PaintVarSkew) isPaintTable()                     {}
func (PaintSkewAroundCenter) isPaintTable()            {}
func (PaintVarSkewAroundCenter) isPaintTable()         {}
func (PaintComposite) isPaintTable()                   {}

type Offset24 [3]byte // TODO:

// (format 1)
type PaintColrLayers struct {
	format          byte   `unionTag:"1"`
	NumLayers       uint8  // Number of offsets to paint tables to read from LayerList.
	FirstLayerIndex uint32 // Index (base 0) into the LayerList.
}

// (format 2)
type PaintSolid struct {
	format       byte     `unionTag:"2"`
	PaletteIndex uint16   // Index for a CPAL palette entry.
	Alpha        Fixed214 // Alpha value.
}

// (format 3)
type PaintVarSolid struct {
	format       byte     `unionTag:"3"`
	PaletteIndex uint16   // Index for a CPAL palette entry.
	Alpha        Fixed214 // Alpha value. For variation, use varIndexBase + 0.
	VarIndexBase uint32   // Base index into DeltaSetIndexMap.
}

// (format 4)
type PaintLinearGradient struct {
	format          byte     `unionTag:"4"`
	ColorLineOffset Offset24 // Offset to ColorLine table, from beginning of PaintLinearGradient table.
	X0              int16    // Start point (p₀) x coordinate.
	Y0              int16    // Start point (p₀) y coordinate.
	X1              int16    // End point (p₁) x coordinate.
	Y1              int16    // End point (p₁) y coordinate.
	X2              int16    // Rotation point (p₂) x coordinate.
	Y2              int16    // Rotation point (p₂) y coordinate.
}

// (format 5)
type PaintVarLinearGradient struct {
	format          byte     `unionTag:"5"`
	ColorLineOffset Offset24 // Offset to VarColorLine table, from beginning of PaintVarLinearGradient table.
	X0              int16    // Start point (p₀) x coordinate. For variation, use varIndexBase + 0.
	Y0              int16    // Start point (p₀) y coordinate. For variation, use varIndexBase + 1.
	X1              int16    // End point (p₁) x coordinate. For variation, use varIndexBase + 2.
	Y1              int16    // End point (p₁) y coordinate. For variation, use varIndexBase + 3.
	X2              int16    // Rotation point (p₂) x coordinate. For variation, use varIndexBase + 4.
	Y2              int16    // Rotation point (p₂) y coordinate. For variation, use varIndexBase + 5.
	VarIndexBase    uint32   // Base index into DeltaSetIndexMap.
}

// (format 6)
type PaintRadialGradient struct {
	format          byte     `unionTag:"6"`
	ColorLineOffset Offset24 // Offset to ColorLine table, from beginning of PaintRadialGradient table.
	X0              int16    // Start circle center x coordinate.
	Y0              int16    // Start circle center y coordinate.
	Radius0         uint16   // Start circle radius.
	X1              int16    // End circle center x coordinate.
	Y1              int16    // End circle center y coordinate.
	Radius1         uint16   // End circle radius.
}

// (format 7)
type PaintVarRadialGradient struct {
	format          byte     `unionTag:"7"`
	ColorLineOffset Offset24 // Offset to VarColorLine table, from beginning of PaintVarRadialGradient table.
	X0              int16    // Start circle center x coordinate. For variation, use varIndexBase + 0.
	Y0              int16    // Start circle center y coordinate. For variation, use varIndexBase + 1.
	Radius0         uint16   // Start circle radius. For variation, use varIndexBase + 2.
	X1              int16    // End circle center x coordinate. For variation, use varIndexBase + 3.
	Y1              int16    // End circle center y coordinate. For variation, use varIndexBase + 4.
	Radius1         uint16   // End circle radius. For variation, use varIndexBase + 5.
	VarIndexBase    uint32   // Base index into DeltaSetIndexMap.
}

// (format 8)
type PaintSweepGradient struct {
	format          byte     `unionTag:"8"`
	ColorLineOffset Offset24 // Offset to ColorLine table, from beginning of PaintSweepGradient table.
	CenterX         int16    // Center x coordinate.
	CenterY         int16    // Center y coordinate.
	StartAngle      Fixed214 // Start of the angular range of the gradient: add 1.0 and multiply by 180° to retrieve counter-clockwise degrees.
	EndAngle        Fixed214 // End of the angular range of the gradient: add 1.0 and multiply by 180° to retrieve counter-clockwise degrees.
}

// (format 9)
type PaintVarSweepGradient struct {
	format          byte     `unionTag:"9"`
	ColorLineOffset Offset24 // Offset to VarColorLine table, from beginning of PaintVarSweepGradient table.
	CenterX         int16    // Center x coordinate. For variation, use varIndexBase + 0.
	CenterY         int16    // Center y coordinate. For variation, use varIndexBase + 1.
	StartAngle      Fixed214 // Start of the angular range of the gradient: add 1.0 and multiply by 180° to retrieve counter-clockwise degrees. For variation, use varIndexBase + 2.
	EndAngle        Fixed214 // End of the angular range of the gradient: add 1.0 and multiply by 180° to retrieve counter-clockwise degrees. For variation, use varIndexBase + 3.
	VarIndexBase    uint32   // Base index into DeltaSetIndexMap.
}

// (format 10)
type PaintGlyph struct {
	format      byte     `unionTag:"10"`
	PaintOffset Offset24 // Offset to a Paint table, from beginning of PaintGlyph table.
	GlyphID     uint16   // Glyph ID for the source outline.
}

// (format 11)
type PaintColrGlyph struct {
	format  byte   `unionTag:"11"`
	GlyphID uint16 // Glyph ID for a BaseGlyphList base glyph.
}

// (format 12)
type PaintTransform struct {
	format          byte     `unionTag:"12"`
	PaintOffset     Offset24 // Offset to a Paint subtable, from beginning of PaintTransform table.
	TransformOffset Offset24 // Offset to an Affine2x3 table, from beginning of PaintTransform table.
}

// (format 13)
type PaintVarTransform struct {
	format          byte     `unionTag:"13"`
	PaintOffset     Offset24 // Offset to a Paint subtable, from beginning of PaintVarTransform table.
	TransformOffset Offset24 // Offset to a VarAffine2x3 table, from beginning of PaintVarTransform table.
}

// (format 14)
type PaintTranslate struct {
	format      byte     `unionTag:"14"`
	PaintOffset Offset24 // Offset to a Paint subtable, from beginning of PaintTranslate table.
	Dx          int16    // Translation in x direction.
	Dy          int16    // Translation in y direction.
}

// (format 15)
type PaintVarTranslate struct {
	format       byte     `unionTag:"15"`
	PaintOffset  Offset24 // Offset to a Paint subtable, from beginning of PaintVarTranslate table.
	Dx           int16    // Translation in x direction. For variation, use varIndexBase + 0.
	Dy           int16    // Translation in y direction. For variation, use varIndexBase + 1.
	VarIndexBase uint32   // Base index into DeltaSetIndexMap.
}

// (format 16)
type PaintScale struct {
	format      byte     `unionTag:"16"`
	PaintOffset Offset24 // Offset to a Paint subtable, from beginning of PaintScale table.
	ScaleX      Fixed214 // Scale factor in x direction.
	ScaleY      Fixed214 // Scale factor in y direction.
}

// (format 17)
type PaintVarScale struct {
	format       byte     `unionTag:"17"`
	PaintOffset  Offset24 // Offset to a Paint subtable, from beginning of PaintVarScale table.
	ScaleX       Fixed214 // Scale factor in x direction. For variation, use varIndexBase + 0.
	ScaleY       Fixed214 // Scale factor in y direction. For variation, use varIndexBase + 1.
	VarIndexBase uint32   // Base index into DeltaSetIndexMap.
}

// (format 18)
type PaintScaleAroundCenter struct {
	format      byte     `unionTag:"18"`
	PaintOffset Offset24 // Offset to a Paint subtable, from beginning of PaintScaleAroundCenter table.
	ScaleX      Fixed214 // Scale factor in x direction.
	ScaleY      Fixed214 // Scale factor in y direction.
	CenterX     int16    // x coordinate for the center of scaling.
	CenterY     int16    // y coordinate for the center of scaling.
}

// (format 19)
type PaintVarScaleAroundCenter struct {
	format       byte     `unionTag:"19"`
	PaintOffset  Offset24 // Offset to a Paint subtable, from beginning of PaintVarScaleAroundCenter table.
	ScaleX       Fixed214 // Scale factor in x direction. For variation, use varIndexBase + 0.
	ScaleY       Fixed214 // Scale factor in y direction. For variation, use varIndexBase + 1.
	CenterX      int16    // x coordinate for the center of scaling. For variation, use varIndexBase + 2.
	CenterY      int16    // y coordinate for the center of scaling. For variation, use varIndexBase + 3.
	VarIndexBase uint32   // Base index into DeltaSetIndexMap.
}

// (format 20)
type PaintScaleUniform struct {
	format      byte     `unionTag:"20"`
	PaintOffset Offset24 // Offset to a Paint subtable, from beginning of PaintScaleUniform table.
	Scale       Fixed214 // Scale factor in x and y directions.
}

// (format 21)
type PaintVarScaleUniform struct {
	format       byte     `unionTag:"21"`
	PaintOffset  Offset24 // Offset to a Paint subtable, from beginning of PaintVarScaleUniform table.
	Scale        Fixed214 // Scale factor in x and y directions. For variation, use varIndexBase + 0.
	VarIndexBase uint32   // Base index into DeltaSetIndexMap.
}

// (format 22)
type PaintScaleUniformAroundCenter struct {
	format      byte     `unionTag:"22"`
	PaintOffset Offset24 // Offset to a Paint subtable, from beginning of PaintScaleUniformAroundCenter table.
	Scale       Fixed214 // Scale factor in x and y directions.
	CenterX     int16    // x coordinate for the center of scaling.
	CenterY     int16    // y coordinate for the center of scaling.
}

// (format 23)
type PaintVarScaleUniformAroundCenter struct {
	format       byte     `unionTag:"23"`
	PaintOffset  Offset24 // Offset to a Paint subtable, from beginning of PaintVarScaleUniformAroundCenter table.
	Scale        Fixed214 // Scale factor in x and y directions. For variation, use varIndexBase + 0.
	CenterX      int16    // x coordinate for the center of scaling. For variation, use varIndexBase + 1.
	CenterY      int16    // y coordinate for the center of scaling. For variation, use varIndexBase + 2.
	VarIndexBase uint32   // Base index into DeltaSetIndexMap.
}

// (format 24)
type PaintRotate struct {
	format      byte     `unionTag:"24"`
	PaintOffset Offset24 // Offset to a Paint subtable, from beginning of PaintRotate table.
	Angle       Fixed214 // Rotation angle, 180° in counter-clockwise degrees per 1.0 of value.
}

// (format 25)
type PaintVarRotate struct {
	format       byte     `unionTag:"25"`
	PaintOffset  Offset24 // Offset to a Paint subtable, from beginning of PaintVarRotate table.
	Angle        Fixed214 // Rotation angle, 180° in counter-clockwise degrees per 1.0 of value. For variation, use varIndexBase + 0.
	VarIndexBase uint32   // Base index into DeltaSetIndexMap.
}

// (format 26)
type PaintRotateAroundCenter struct {
	format      byte     `unionTag:"26"`
	PaintOffset Offset24 // Offset to a Paint subtable, from beginning of PaintRotateAroundCenter table.
	Angle       Fixed214 // Rotation angle, 180° in counter-clockwise degrees per 1.0 of value.
	CenterX     int16    // x coordinate for the center of rotation.
	CenterY     int16    // y coordinate for the center of rotation.
}

// (format 27)
type PaintVarRotateAroundCenter struct {
	format       byte     `unionTag:"27"`
	PaintOffset  Offset24 // Offset to a Paint subtable, from beginning of PaintVarRotateAroundCenter table.
	Angle        Fixed214 // Rotation angle, 180° in counter-clockwise degrees per 1.0 of value. For variation, use varIndexBase + 0.
	CenterX      int16    // x coordinate for the center of rotation. For variation, use varIndexBase + 1.
	CenterY      int16    // y coordinate for the center of rotation. For variation, use varIndexBase + 2.
	VarIndexBase uint32   // Base index into DeltaSetIndexMap.
}

// (format 28)
type PaintSkew struct {
	format      byte     `unionTag:"28"`
	PaintOffset Offset24 // Offset to a Paint subtable, from beginning of PaintSkew table.
	XSkewAngle  Fixed214 // Angle of skew in the direction of the x-axis, 180° in counter-clockwise degrees per 1.0 of value.
	YSkewAngle  Fixed214 // Angle of skew in the direction of the y-axis, 180° in counter-clockwise degrees per 1.0 of value.
}

// (format 29)
type PaintVarSkew struct {
	format       byte     `unionTag:"29"`
	PaintOffset  Offset24 // Offset to a Paint subtable, from beginning of PaintVarSkew table.
	XSkewAngle   Fixed214 // Angle of skew in the direction of the x-axis, 180° in counter-clockwise degrees per 1.0 of value. For variation, use varIndexBase + 0.
	YSkewAngle   Fixed214 // Angle of skew in the direction of the y-axis, 180° in counter-clockwise degrees per 1.0 of value. For variation, use varIndexBase + 1.
	VarIndexBase uint32   // Base index into DeltaSetIndexMap.
}

// (format 30)
type PaintSkewAroundCenter struct {
	format      byte     `unionTag:"30"`
	PaintOffset Offset24 // Offset to a Paint subtable, from beginning of PaintSkewAroundCenter table.
	XSkewAngle  Fixed214 // Angle of skew in the direction of the x-axis, 180° in counter-clockwise degrees per 1.0 of value.
	YSkewAngle  Fixed214 // Angle of skew in the direction of the y-axis, 180° in counter-clockwise degrees per 1.0 of value.
	CenterX     int16    // x coordinate for the center of rotation.
	CenterY     int16    // y coordinate for the center of rotation.
}

// (format 31)
type PaintVarSkewAroundCenter struct {
	format       byte     `unionTag:"31"`
	PaintOffset  Offset24 // Offset to a Paint subtable, from beginning of PaintVarSkewAroundCenter table.
	XSkewAngle   Fixed214 // Angle of skew in the direction of the x-axis, 180° in counter-clockwise degrees per 1.0 of value. For variation, use varIndexBase + 0.
	YSkewAngle   Fixed214 // Angle of skew in the direction of the y-axis, 180° in counter-clockwise degrees per 1.0 of value. For variation, use varIndexBase + 1.
	CenterX      int16    // x coordinate for the center of rotation. For variation, use varIndexBase + 2.
	CenterY      int16    // y coordinate for the center of rotation. For variation, use varIndexBase + 3.
	VarIndexBase uint32   // Base index into DeltaSetIndexMap.
}

// (format 32)
type PaintComposite struct {
	format              byte     `unionTag:"32"`
	SourcePaintOffset   Offset24 // Offset to a source Paint table, from beginning of PaintComposite table.
	CompositeMode       uint8    // A CompositeMode enumeration value.
	BackdropPaintOffset Offset24 // Offset to a backdrop Paint table, from beginning of PaintComposite table.
}
