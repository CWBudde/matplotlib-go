package style

import "testing"

func TestSupportedMPLStyleKeysReturnSortedCopy(t *testing.T) {
	keys := SupportedMPLStyleKeys()
	if len(keys) == 0 {
		t.Fatal("supported key list is empty")
	}
	for i := 1; i < len(keys); i++ {
		if keys[i-1] > keys[i] {
			t.Fatalf("keys are not sorted at %d: %q > %q", i, keys[i-1], keys[i])
		}
	}

	keys[0] = "mutated"
	fresh := SupportedMPLStyleKeys()
	if fresh[0] == "mutated" {
		t.Fatal("SupportedMPLStyleKeys returned mutable backing storage")
	}
}

func TestMPLStyleParamsApplySupportedKeysAndReportUnsupported(t *testing.T) {
	params := Params{
		"figure.dpi":         "144",
		"unsupported.option": "value",
	}
	rc, report, err := applyMPLStyleParams(Default, params)
	if err != nil {
		t.Fatal(err)
	}
	if rc.DPI != 144 {
		t.Fatalf("DPI = %v, want 144", rc.DPI)
	}
	if len(report.Applied) != 1 || report.Applied[0] != "figure.dpi" {
		t.Fatalf("applied report = %+v", report.Applied)
	}
	if len(report.Unsupported) != 1 || report.Unsupported[0].Key != "unsupported.option" {
		t.Fatalf("unsupported report = %+v", report.Unsupported)
	}
}
