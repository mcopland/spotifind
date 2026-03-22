package sync

import "testing"

func TestParseYear_FullDate(t *testing.T) {
	if got := parseYear("2024-01-15"); got != 2024 {
		t.Errorf("expected 2024, got %d", got)
	}
}

func TestParseYear_YearOnly(t *testing.T) {
	if got := parseYear("2020"); got != 2020 {
		t.Errorf("expected 2020, got %d", got)
	}
}

func TestParseYear_NonNumericPrefix(t *testing.T) {
	if got := parseYear("abcd-01-01"); got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
}

func TestParseYear_Empty(t *testing.T) {
	if got := parseYear(""); got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
}

func TestParseYear_TooShort(t *testing.T) {
	if got := parseYear("202"); got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
}
