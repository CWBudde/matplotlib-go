package pdfcompare

import (
	"bytes"
	"compress/zlib"
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestParseAndDiffIgnoresObjectWhitespaceAndXRefOffsets(t *testing.T) {
	expected := []byte("%PDF-1.7\n" +
		"1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n" +
		"2 0 obj\n[ 1 2 3 ]\nendobj\n" +
		"xref\n0 3\n0000000000 65535 f \n0000000010 00000 n \n0000000060 00000 n \n" +
		"trailer\n<< /Size 3 /Root 1 0 R >>\nstartxref\n99\n%%EOF\n")
	actual := []byte("%PDF-1.7\n\n" +
		"1 0 obj\n<<\n/Type/Catalog/Pages 2 0 R\n>>\nendobj\n" +
		"2 0 obj\n[1 2 3]\nendobj\n" +
		"xref\n0 3\n0000000000 65535 f \n0000001234 00000 n \n0000005678 00000 n \n" +
		"trailer\n<< /Root 1 0 R /Size 3 >>\nstartxref\n42\n%%EOF\n")

	diff, err := ParseAndDiff(expected, actual)
	if err != nil {
		t.Fatalf("ParseAndDiff: %v", err)
	}
	if diff != "" {
		t.Fatalf("whitespace and xref offset differences should compare equal, got: %s", diff)
	}
}

func TestParseAndDiffReportsObjectBodyMismatch(t *testing.T) {
	expected := []byte("%PDF-1.7\n1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")
	actual := []byte("%PDF-1.7\n1 0 obj\n<< /Type /Page /Pages 2 0 R >>\nendobj\n")

	diff, err := ParseAndDiff(expected, actual)
	if err != nil {
		t.Fatalf("ParseAndDiff: %v", err)
	}
	if !strings.Contains(diff, "1 0 obj") || !strings.Contains(diff, "/Catalog") || !strings.Contains(diff, "/Page") {
		t.Fatalf("diff should report object and values, got: %s", diff)
	}
}

func TestParseAndDiffDecodesFlateStreamsBeforeComparing(t *testing.T) {
	expected := pdfWithFlateStream(t, "10 10 m\n20 20 l\nS\n")
	actual := pdfWithFlateStream(t, "10 10 m 20 20 l S")

	diff, err := ParseAndDiff(expected, actual)
	if err != nil {
		t.Fatalf("ParseAndDiff: %v", err)
	}
	if diff != "" {
		t.Fatalf("flate stream token whitespace should compare equal, got: %s", diff)
	}
}

func TestParseAndDiffReportsFlateStreamMismatch(t *testing.T) {
	expected := pdfWithFlateStream(t, "10 10 m\n20 20 l\nS\n")
	actual := pdfWithFlateStream(t, "10 10 m\n30 20 l\nS\n")

	diff, err := ParseAndDiff(expected, actual)
	if err != nil {
		t.Fatalf("ParseAndDiff: %v", err)
	}
	if !strings.Contains(diff, "stream differs") || !strings.Contains(diff, "20") || !strings.Contains(diff, "30") {
		t.Fatalf("diff should report decoded stream mismatch, got: %s", diff)
	}
}

func TestParseAndDiffTruncatesLargeStreamMismatch(t *testing.T) {
	expected := pdfWithFlateStream(t, strings.Repeat("10 10 m 20 20 l S\n", 4096))
	actual := pdfWithFlateStream(t, strings.Repeat("10 10 m 30 20 l S\n", 4096))

	diff, err := ParseAndDiff(expected, actual)
	if err != nil {
		t.Fatalf("ParseAndDiff: %v", err)
	}
	if !strings.Contains(diff, "truncated") {
		t.Fatalf("large diff should be truncated, got: %s", diff)
	}
	if len(diff) > 4096 {
		t.Fatalf("large diff length = %d, want bounded output", len(diff))
	}
}

func TestParseAndDiffHashesFlateImageStreams(t *testing.T) {
	pixels := bytes.Repeat([]byte{0, 10, 20, 30, 40, 50, 60}, 8192)
	expected := pdfWithFlateImageStream(t, pixels)
	actual := pdfWithFlateImageStream(t, append([]byte(nil), pixels...))

	diff, err := ParseAndDiff(expected, actual)
	if err != nil {
		t.Fatalf("ParseAndDiff: %v", err)
	}
	if diff != "" {
		t.Fatalf("matching image streams should compare equal, got: %s", diff)
	}

	changed := append([]byte(nil), pixels...)
	changed[len(changed)-1] ^= 0xff
	diff, err = ParseAndDiff(expected, pdfWithFlateImageStream(t, changed))
	if err != nil {
		t.Fatalf("ParseAndDiff changed image: %v", err)
	}
	wantHash := fmt.Sprintf("%x", sha256.Sum256(changed))
	if !strings.Contains(diff, "stream-bytes") || !strings.Contains(diff, wantHash) {
		t.Fatalf("image stream mismatch should report byte hashes, got: %s", diff)
	}
	if len(diff) > 4096 {
		t.Fatalf("image stream diff length = %d, want bounded output", len(diff))
	}
}

func TestParseReportsMissingPDFHeader(t *testing.T) {
	if _, err := Parse([]byte("1 0 obj\n<< >>\nendobj\n")); err == nil {
		t.Fatal("missing PDF header should produce an error")
	}
}

func pdfWithFlateStream(t *testing.T, stream string) []byte {
	t.Helper()
	var compressed bytes.Buffer
	zw := zlib.NewWriter(&compressed)
	if _, err := zw.Write([]byte(stream)); err != nil {
		t.Fatalf("zlib write: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zlib close: %v", err)
	}
	var out bytes.Buffer
	out.WriteString("%PDF-1.7\n")
	out.WriteString("1 0 obj\n")
	out.WriteString("<< /Length ")
	out.WriteString(strconv.Itoa(compressed.Len()))
	out.WriteString(" /Filter /FlateDecode >>\nstream\n")
	out.Write(compressed.Bytes())
	out.WriteString("\nendstream\nendobj\n")
	return out.Bytes()
}

func pdfWithFlateImageStream(t *testing.T, stream []byte) []byte {
	t.Helper()
	var compressed bytes.Buffer
	zw := zlib.NewWriter(&compressed)
	if _, err := zw.Write(stream); err != nil {
		t.Fatalf("zlib write: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zlib close: %v", err)
	}
	var out bytes.Buffer
	out.WriteString("%PDF-1.7\n")
	out.WriteString("1 0 obj\n")
	out.WriteString("<< /Type /XObject /Subtype /Image /Width 128 /Height 64 /ColorSpace /DeviceRGB /BitsPerComponent 8 /Length ")
	out.WriteString(strconv.Itoa(compressed.Len()))
	out.WriteString(" /Filter /FlateDecode >>\nstream\n")
	out.Write(compressed.Bytes())
	out.WriteString("\nendstream\nendobj\n")
	return out.Bytes()
}
