package speed

import (
	"reflect"
	"testing"
)

func TestDropExtremesRemovesSingleMinAndMaxSample(t *testing.T) {
	trimmed := dropExtremes([]uint64{1, 2, 3, 4, 5, 6, 7})

	expected := []uint64{2, 3, 4, 5, 6}
	if !reflect.DeepEqual(trimmed, expected) {
		t.Fatalf("expected trimmed speeds %v, got %v", expected, trimmed)
	}
}

func TestDropExtremesKeepsTwoOrFewerSamples(t *testing.T) {
	trimmed := dropExtremes([]uint64{1, 2})

	expected := []uint64{1, 2}
	if !reflect.DeepEqual(trimmed, expected) {
		t.Fatalf("expected untrimmed speeds %v, got %v", expected, trimmed)
	}
}
