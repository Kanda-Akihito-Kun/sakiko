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

func TestWriteCounterTracksTotalBytes(t *testing.T) {
	counter := &writeCounter{}

	if _, err := counter.Write(make([]byte, 512)); err != nil {
		t.Fatalf("Write() first error = %v", err)
	}
	if _, err := counter.Write(make([]byte, 256)); err != nil {
		t.Fatalf("Write() second error = %v", err)
	}

	if got := counter.Total(); got != 768 {
		t.Fatalf("expected total bytes 768, got %d", got)
	}
	if got := counter.Take(); got != 768 {
		t.Fatalf("expected take bytes 768, got %d", got)
	}
	if got := counter.Take(); got != 0 {
		t.Fatalf("expected second take bytes 0, got %d", got)
	}
}
