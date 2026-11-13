package main

import (
	"testing"
	"time"
)

func TestLatencyHistogramQuantiles(t *testing.T) {
	h := NewLatencyHistogram()
	for i := 1; i <= 10000; i++ {
		h.Record(1500*time.Millisecond + time.Duration(i)*10*time.Microsecond)
	}

	if h.Count() != 10000 {
		t.Fatalf("count = %d, want 10000", h.Count())
	}
	if h.Max() != 1600*time.Millisecond {
		t.Fatalf("max = %s, want 1.6s", h.Max())
	}
	for _, c := range []struct {
		q    float64
		want time.Duration
	}{
		{0.50, 1550 * time.Millisecond},
		{0.95, 1595 * time.Millisecond},
		{0.99, 1599 * time.Millisecond},
	} {
		got := h.Quantile(c.q)
		if diff := got - c.want; diff < 0 || diff > 200*time.Microsecond {
			t.Errorf("p%g = %s, want %s within 0.2 ms above", c.q*100, got, c.want)
		}
	}
}

func TestLatencyHistogramBucketsCoverEveryRecord(t *testing.T) {
	h := NewLatencyHistogram()
	values := []time.Duration{0, 999 * time.Microsecond, 1501 * time.Millisecond, 1501 * time.Millisecond, 19 * time.Second}
	for _, v := range values {
		h.Record(v)
	}

	var total uint64
	var prevTo float64
	for _, b := range h.Buckets() {
		if b.FromMs < prevTo {
			t.Fatalf("bucket %v overlaps the previous one ending at %v", b, prevTo)
		}
		if b.ToMs <= b.FromMs {
			t.Fatalf("bucket %v has no width", b)
		}
		total += b.Count
		prevTo = b.ToMs
	}
	if total != uint64(len(values)) {
		t.Fatalf("buckets hold %d records, want %d", total, len(values))
	}
	for _, v := range values {
		ms := float64(v) / float64(time.Millisecond)
		found := false
		for _, b := range h.Buckets() {
			if ms >= b.FromMs && ms < b.ToMs {
				found = true
			}
		}
		if !found {
			t.Errorf("no bucket holds %v ms", ms)
		}
	}
}

func TestLatencyHistogramEmpty(t *testing.T) {
	h := NewLatencyHistogram()
	if h.Quantile(0.99) != 0 || h.Max() != 0 || h.Mean() != 0 || len(h.Buckets()) != 0 {
		t.Fatal("empty histogram should report zeros and no buckets")
	}
}
