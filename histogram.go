package main

import (
	"math"
	"math/bits"
	"sort"
	"time"
)

const histogramSubBucketBits = 14

type latencyHistogram struct {
	counts map[uint64]uint64
	total  uint64
	sum    time.Duration
	max    time.Duration
}

type HistogramBucket struct {
	FromMs float64 `json:"from_ms"`
	ToMs   float64 `json:"to_ms"`
	Count  uint64  `json:"count"`
}

func newLatencyHistogram() *latencyHistogram {
	return &latencyHistogram{counts: make(map[uint64]uint64)}
}

func histogramBucketWidth(us uint64) uint64 {
	shift := bits.Len64(us) - histogramSubBucketBits
	if shift <= 0 {
		return 1
	}
	return 1 << shift
}

func (h *latencyHistogram) Record(d time.Duration) {
	if d < 0 {
		d = 0
	}
	us := uint64(d / time.Microsecond)
	width := histogramBucketWidth(us)
	h.counts[us/width*width]++
	h.total++
	h.sum += d
	if d > h.max {
		h.max = d
	}
}

func (h *latencyHistogram) Count() uint64 {
	return h.total
}

func (h *latencyHistogram) Max() time.Duration {
	return h.max
}

func (h *latencyHistogram) Mean() time.Duration {
	if h.total == 0 {
		return 0
	}
	return h.sum / time.Duration(h.total)
}

func (h *latencyHistogram) sortedStarts() []uint64 {
	starts := make([]uint64, 0, len(h.counts))
	for start := range h.counts {
		starts = append(starts, start)
	}
	sort.Slice(starts, func(i, j int) bool { return starts[i] < starts[j] })
	return starts
}

func (h *latencyHistogram) Quantile(q float64) time.Duration {
	if h.total == 0 {
		return 0
	}
	rank := uint64(math.Ceil(q * float64(h.total)))
	if rank < 1 {
		rank = 1
	}
	var seen uint64
	for _, start := range h.sortedStarts() {
		seen += h.counts[start]
		if seen >= rank {
			upper := time.Duration(start+histogramBucketWidth(start)) * time.Microsecond
			return min(upper, h.max)
		}
	}
	return h.max
}

func (h *latencyHistogram) Buckets() []HistogramBucket {
	starts := h.sortedStarts()
	buckets := make([]HistogramBucket, 0, len(starts))
	for _, start := range starts {
		buckets = append(buckets, HistogramBucket{
			FromMs: float64(start) / 1000,
			ToMs:   float64(start+histogramBucketWidth(start)) / 1000,
			Count:  h.counts[start],
		})
	}
	return buckets
}
