package utils

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

func Clamp[T ~int | ~int64 | ~float64](value, minValue, maxValue T) T {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func WithJitter(base time.Duration, rng *rand.Rand, percent float64) time.Duration {
	if base <= 0 {
		return 0
	}
	if percent <= 0 || rng == nil {
		return base
	}
	delta := float64(base) * percent
	shift := (rng.Float64()*2 - 1) * delta
	result := time.Duration(float64(base) + shift)
	if result < time.Millisecond {
		return time.Millisecond
	}
	return result
}

func Slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.NewReplacer(" ", "-", "_", "-", "/", "-", "\\", "-", ".", "-", ":", "-").Replace(value)
	value = strings.Trim(value, "-")
	if value == "" {
		return "item"
	}
	return value
}

func ShortID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}
