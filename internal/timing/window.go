package timing

import (
	"math"
	"sort"
	"time"
)

type TimedValue[T any] struct {
	At    time.Time
	Value T
}

type Pair[A, B any] struct {
	Left  TimedValue[A]
	Right TimedValue[B]
	Skew  time.Duration
}

type Window[A, B any] struct {
	maxSkew time.Duration
	left    []TimedValue[A]
	right   []TimedValue[B]
}

func NewWindow[A, B any](maxSkew time.Duration) *Window[A, B] {
	return &Window[A, B]{maxSkew: maxSkew}
}

func (w *Window[A, B]) AddLeft(value TimedValue[A]) {
	w.left = append(w.left, value)
	sort.Slice(w.left, func(i, j int) bool { return w.left[i].At.Before(w.left[j].At) })
}

func (w *Window[A, B]) AddRight(value TimedValue[B]) {
	w.right = append(w.right, value)
	sort.Slice(w.right, func(i, j int) bool { return w.right[i].At.Before(w.right[j].At) })
}

func (w *Window[A, B]) MatchNewest() (Pair[A, B], bool) {
	var best Pair[A, B]
	bestDistance := time.Duration(math.MaxInt64)
	bestLeft, bestRight := -1, -1
	for li := range w.left {
		for ri := range w.right {
			distance := w.left[li].At.Sub(w.right[ri].At)
			if distance < 0 {
				distance = -distance
			}
			if distance <= w.maxSkew && distance <= bestDistance {
				best = Pair[A, B]{Left: w.left[li], Right: w.right[ri], Skew: distance}
				bestDistance, bestLeft, bestRight = distance, li, ri
			}
		}
	}
	if bestLeft < 0 {
		return Pair[A, B]{}, false
	}
	w.left = append(w.left[:bestLeft], w.left[bestLeft+1:]...)
	w.right = append(w.right[:bestRight], w.right[bestRight+1:]...)
	return best, true
}

func (w *Window[A, B]) DropBefore(cutoff time.Time) {
	w.left = keepFrom(w.left, cutoff)
	w.right = keepFrom(w.right, cutoff)
}

func keepFrom[T any](values []TimedValue[T], cutoff time.Time) []TimedValue[T] {
	index := sort.Search(len(values), func(i int) bool { return !values[i].At.Before(cutoff) })
	return append([]TimedValue[T](nil), values[index:]...)
}
