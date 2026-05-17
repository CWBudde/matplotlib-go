package render

import (
	"strings"
	"sync"

	"golang.org/x/image/font/sfnt"
)

type gposMarkPair struct {
	base sfnt.GlyphIndex
	mark sfnt.GlyphIndex
}

type gposMarkOffset struct {
	x int16
	y int16
}

type gposAnchor struct {
	x int16
	y int16
}

type gposMarkRecord struct {
	class  uint16
	anchor gposAnchor
}

type gposMarkToBaseTable struct {
	Offsets map[gposMarkPair]gposMarkOffset
}

var (
	gposMarkToBaseCacheMu sync.RWMutex
	gposMarkToBaseCache   = map[string]gposMarkToBaseTable{}
)

func gposMarkToBaseTableForFace(face FontFace, opts TextShapingOptions) (gposMarkToBaseTable, bool) {
	if !openTypeFeatureEnabled(opts, "mark", true) {
		return gposMarkToBaseTable{}, false
	}
	key := fontFaceCacheKey(face) + "|mark"
	if key == "|mark" {
		return gposMarkToBaseTable{}, false
	}

	gposMarkToBaseCacheMu.RLock()
	if cached, ok := gposMarkToBaseCache[key]; ok {
		gposMarkToBaseCacheMu.RUnlock()
		return cached, len(cached.Offsets) > 0
	}
	gposMarkToBaseCacheMu.RUnlock()

	data, err := loadFontFaceData(face)
	if err != nil {
		return gposMarkToBaseTable{}, false
	}
	table := parseGPOSMarkToBaseTable(data, []string{"mark"})

	gposMarkToBaseCacheMu.Lock()
	gposMarkToBaseCache[key] = table
	gposMarkToBaseCacheMu.Unlock()
	return table, len(table.Offsets) > 0
}

func (t gposMarkToBaseTable) offset(base, mark sfnt.GlyphIndex) (gposMarkOffset, bool) {
	offset, ok := t.Offsets[gposMarkPair{base: base, mark: mark}]
	return offset, ok
}

func parseGPOSMarkToBaseTable(fontData []byte, tags []string) gposMarkToBaseTable {
	gpos, ok := sfntTable(fontData, "GPOS")
	if !ok || len(gpos) < 10 {
		return gposMarkToBaseTable{}
	}
	featureListOffset := int(be16(gpos, 6))
	lookupListOffset := int(be16(gpos, 8))
	if featureListOffset <= 0 || lookupListOffset <= 0 || featureListOffset >= len(gpos) || lookupListOffset >= len(gpos) {
		return gposMarkToBaseTable{}
	}

	wanted := map[string]bool{}
	for _, tag := range tags {
		wanted[normalizeOpenTypeTag(tag)] = true
	}
	lookupIndices := gsubFeatureLookupIndices(gpos[featureListOffset:], wanted)
	if len(lookupIndices) == 0 {
		return gposMarkToBaseTable{}
	}

	table := gposMarkToBaseTable{Offsets: map[gposMarkPair]gposMarkOffset{}}
	parseGPOSMarkToBaseLookups(gpos, lookupListOffset, lookupIndices, table.Offsets)
	return table
}

func parseGPOSMarkToBaseLookups(gpos []byte, lookupListOffset int, lookupIndices []uint16, offsets map[gposMarkPair]gposMarkOffset) {
	if lookupListOffset+2 > len(gpos) {
		return
	}
	lookupCount := int(be16(gpos, lookupListOffset))
	for _, lookupIndex := range lookupIndices {
		if int(lookupIndex) >= lookupCount {
			continue
		}
		lookupOffsetOffset := lookupListOffset + 2 + int(lookupIndex)*2
		if lookupOffsetOffset+2 > len(gpos) {
			continue
		}
		lookupOffset := lookupListOffset + int(be16(gpos, lookupOffsetOffset))
		if lookupOffset+6 > len(gpos) {
			continue
		}
		lookupType := be16(gpos, lookupOffset)
		subtableCount := int(be16(gpos, lookupOffset+4))
		for i := 0; i < subtableCount; i++ {
			subOffsetOffset := lookupOffset + 6 + i*2
			if subOffsetOffset+2 > len(gpos) {
				break
			}
			subtableOffset := lookupOffset + int(be16(gpos, subOffsetOffset))
			switch lookupType {
			case 4:
				parseGPOSMarkToBaseSubtable(gpos, subtableOffset, offsets)
			case 9:
				parseGPOSExtensionPositioningSubtable(gpos, subtableOffset, 4, offsets)
			}
		}
	}
}

func parseGPOSExtensionPositioningSubtable(gpos []byte, offset int, wantedType uint16, offsets map[gposMarkPair]gposMarkOffset) {
	if offset+8 > len(gpos) || be16(gpos, offset) != 1 || be16(gpos, offset+2) != wantedType {
		return
	}
	extensionOffset := offset + int(be32(gpos, offset+4))
	if extensionOffset <= offset || extensionOffset >= len(gpos) {
		return
	}
	parseGPOSMarkToBaseSubtable(gpos, extensionOffset, offsets)
}

func parseGPOSMarkToBaseSubtable(gpos []byte, offset int, offsets map[gposMarkPair]gposMarkOffset) {
	if offset+12 > len(gpos) || be16(gpos, offset) != 1 {
		return
	}
	markCoverage := parseCoverageGlyphs(gpos, offset+int(be16(gpos, offset+2)))
	baseCoverage := parseCoverageGlyphs(gpos, offset+int(be16(gpos, offset+4)))
	classCount := int(be16(gpos, offset+6))
	markArrayOffset := offset + int(be16(gpos, offset+8))
	baseArrayOffset := offset + int(be16(gpos, offset+10))
	if len(markCoverage) == 0 || len(baseCoverage) == 0 || classCount <= 0 {
		return
	}

	marks := parseGPOSMarkArray(gpos, markArrayOffset, markCoverage)
	bases := parseGPOSBaseArray(gpos, baseArrayOffset, baseCoverage, classCount)
	for markGlyph, mark := range marks {
		for baseGlyph, baseAnchors := range bases {
			baseAnchor, ok := baseAnchors[mark.class]
			if !ok {
				continue
			}
			offsets[gposMarkPair{base: baseGlyph, mark: markGlyph}] = gposMarkOffset{
				x: baseAnchor.x - mark.anchor.x,
				y: baseAnchor.y - mark.anchor.y,
			}
		}
	}
}

func parseGPOSMarkArray(gpos []byte, offset int, coverage []sfnt.GlyphIndex) map[sfnt.GlyphIndex]gposMarkRecord {
	if offset+2 > len(gpos) {
		return nil
	}
	count := int(be16(gpos, offset))
	out := map[sfnt.GlyphIndex]gposMarkRecord{}
	for i := 0; i < count && i < len(coverage); i++ {
		recordOffset := offset + 2 + i*4
		if recordOffset+4 > len(gpos) {
			break
		}
		anchorOffset := int(be16(gpos, recordOffset+2))
		if anchorOffset == 0 {
			continue
		}
		anchor, ok := parseGPOSAnchor(gpos, offset+anchorOffset)
		if !ok {
			continue
		}
		out[coverage[i]] = gposMarkRecord{
			class:  be16(gpos, recordOffset),
			anchor: anchor,
		}
	}
	return out
}

func parseGPOSBaseArray(gpos []byte, offset int, coverage []sfnt.GlyphIndex, classCount int) map[sfnt.GlyphIndex]map[uint16]gposAnchor {
	if offset+2 > len(gpos) {
		return nil
	}
	count := int(be16(gpos, offset))
	out := map[sfnt.GlyphIndex]map[uint16]gposAnchor{}
	for i := 0; i < count && i < len(coverage); i++ {
		recordOffset := offset + 2 + i*classCount*2
		if recordOffset+classCount*2 > len(gpos) {
			break
		}
		for class := 0; class < classCount; class++ {
			anchorOffset := int(be16(gpos, recordOffset+class*2))
			if anchorOffset == 0 {
				continue
			}
			anchor, ok := parseGPOSAnchor(gpos, offset+anchorOffset)
			if !ok {
				continue
			}
			if out[coverage[i]] == nil {
				out[coverage[i]] = map[uint16]gposAnchor{}
			}
			out[coverage[i]][uint16(class)] = anchor
		}
	}
	return out
}

func parseGPOSAnchor(gpos []byte, offset int) (gposAnchor, bool) {
	if offset+6 > len(gpos) {
		return gposAnchor{}, false
	}
	switch be16(gpos, offset) {
	case 1, 3:
		return gposAnchor{
			x: beInt16(gpos, offset+2),
			y: beInt16(gpos, offset+4),
		}, true
	case 2:
		if offset+8 > len(gpos) {
			return gposAnchor{}, false
		}
		return gposAnchor{
			x: beInt16(gpos, offset+2),
			y: beInt16(gpos, offset+4),
		}, true
	default:
		return gposAnchor{}, false
	}
}

func beInt16(data []byte, offset int) int16 {
	return int16(be16(data, offset))
}

func enabledArabicFormFeature(formTag string, opts TextShapingOptions) bool {
	return formTag == "" || openTypeFeatureEnabled(opts, strings.TrimSpace(formTag), true)
}
