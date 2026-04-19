package core

import (
	"math"
	"sort"
	"strconv"
	"strings"
)

// Locator computes tick positions for a numeric range.
type Locator interface {
	Ticks(minVal, maxVal float64, targetCount int) []float64
}

// Formatter converts numeric tick values to strings.
type Formatter interface {
	Format(x float64) string
}

// LinearLocator places ticks at nice multiples of 1,2,2.5,5×10^k.
type LinearLocator struct{}

// Ticks returns a strictly increasing slice of ticks that cover [min,max]
// using a step chosen from {1,2,2.5,5}×10^k close to span/targetCount.
func (LinearLocator) Ticks(minVal, maxVal float64, targetCount int) []float64 {
	if targetCount <= 0 {
		targetCount = 1
	}
	if math.IsNaN(minVal) || math.IsNaN(maxVal) {
		return nil
	}
	if minVal == maxVal {
		return []float64{minVal}
	}
	if minVal > maxVal {
		minVal, maxVal = maxVal, minVal
	}
	span := maxVal - minVal
	raw := span / float64(targetCount)
	if raw <= 0 || math.IsInf(raw, 0) || math.IsNaN(raw) {
		return []float64{minVal, maxVal}
	}
	// Determine exponent of 10 for raw step.
	exp := math.Floor(math.Log10(raw))
	base := math.Pow(10, exp)
	candidates := []float64{1 * base, 2 * base, 2.5 * base, 5 * base, 10 * base}
	step := candidates[0]
	best := math.Abs(candidates[0] - raw)
	for _, c := range candidates[1:] {
		if d := math.Abs(c - raw); d < best {
			best = d
			step = c
		}
	}
	// Align start/end to cover [min,max]
	start := math.Floor(minVal/step) * step
	end := math.Ceil(maxVal/step) * step
	// Generate ticks
	// Guard against pathological loops
	nmax := int(2*float64(targetCount) + 20)
	var ticks []float64
	for v, i := start, 0; v <= end+0.5*step && i < nmax; v, i = v+step, i+1 {
		// Avoid negative zero
		if v == 0 {
			v = 0
		}
		ticks = append(ticks, v)
	}
	// Ensure strictly increasing and within coverage
	// Remove potential duplicates due to floating rounding
	out := make([]float64, 0, len(ticks))
	var last float64
	for i, v := range ticks {
		if i == 0 || v > last {
			out = append(out, v)
			last = v
		}
	}
	return out
}

func scalarUsesScientific(x float64) bool {
	ax := math.Abs(x)
	return (ax >= 1e6) || (ax > 0 && ax <= 1e-4)
}

func scalarStepPrecision(step float64) int {
	step = math.Abs(step)
	if step == 0 || math.IsNaN(step) || math.IsInf(step, 0) {
		return 0
	}

	pow10 := 1.0
	for prec := 0; prec <= 12; prec++ {
		scaled := step * pow10
		if approx(scaled, math.Round(scaled), 1e-9*math.Max(1, math.Abs(scaled))) {
			return prec
		}
		pow10 *= 10
	}
	return 6
}

func formatScalarTickLabel(f ScalarFormatter, x, step float64) string {
	if scalarUsesScientific(x) {
		return f.Format(x)
	}

	prec := scalarStepPrecision(step)
	if prec < 0 {
		return f.Format(x)
	}

	if approx(x, 0, 1e-12*math.Max(1, math.Abs(step))) {
		x = 0
	}

	return strconv.FormatFloat(x, 'f', prec, 64)
}

// MinorLinearLocator subdivides the intervals between major ticks.
// N is the number of subdivisions per major interval (e.g. N=5 gives 4 minor ticks
// between each pair of major ticks). If N <= 1, defaults to 5.
type MinorLinearLocator struct {
	N int // subdivisions per major interval
}

func (m MinorLinearLocator) Ticks(minVal, maxVal float64, _ int) []float64 {
	n := m.N
	if n <= 1 {
		n = 5
	}
	// Get major ticks to subdivide between them
	majors := (LinearLocator{}).Ticks(minVal, maxVal, 6)
	if len(majors) < 2 {
		return nil
	}

	step := (majors[1] - majors[0]) / float64(n)
	if step <= 0 {
		return nil
	}

	// Generate minor ticks across the full range, excluding major positions
	start := majors[0]
	end := majors[len(majors)-1]
	var ticks []float64
	for v := start; v <= end+step*0.5; v += step {
		// Skip if this coincides with a major tick
		isMajor := false
		for _, mj := range majors {
			if math.Abs(v-mj) < step*0.01 {
				isMajor = true
				break
			}
		}
		if !isMajor && v >= minVal && v <= maxVal {
			ticks = append(ticks, v)
		}
	}
	return ticks
}

// LogLocator produces logarithmic ticks for positive domains. Major ticks
// at Base^k within [min,max]. If Minor is true, places minor ticks at
// 2×Base^k and 5×Base^k where they lie within [min,max].
type LogLocator struct {
	Base  float64
	Minor bool
}

func (l LogLocator) Ticks(minVal, maxVal float64, targetCount int) []float64 {
	base := l.Base
	if base <= 1 {
		return nil
	}
	if minVal > maxVal {
		minVal, maxVal = maxVal, minVal
	}
	if minVal <= 0 || maxVal <= 0 {
		return nil
	}
	// Find exponent range
	lb := math.Log(base)
	kmin := math.Ceil(math.Log(minVal) / lb)
	kmax := math.Floor(math.Log(maxVal)/lb + 1e-10) // Add small epsilon to handle floating point precision
	var ticks []float64
	// Majors
	for k := kmin; k <= kmax; k++ {
		v := math.Pow(base, k)
		if v >= minVal && v <= maxVal {
			ticks = append(ticks, v)
		}
		if l.Minor {
			// Minors at 2,5 per decade (common convention)
			m2 := 2 * math.Pow(base, k)
			m5 := 5 * math.Pow(base, k)
			if m2 > v && m2 < math.Pow(base, k+1) && m2 >= minVal && m2 <= maxVal {
				ticks = append(ticks, m2)
			}
			if m5 > v && m5 < math.Pow(base, k+1) && m5 >= minVal && m5 <= maxVal {
				ticks = append(ticks, m5)
			}
		}
	}
	sort.Float64s(ticks)
	// Deduplicate
	out := ticks[:0]
	var last float64
	first := true
	for _, v := range ticks {
		if first || v > last {
			out = append(out, v)
			last = v
			first = false
		}
	}
	return out
}

// ScalarFormatter formats numbers with fixed precision and trims trailing zeros.
// Uses scientific notation if |x| >= 1e6 or (0 < |x| <= 1e-4).
type ScalarFormatter struct{ Prec int }

func (f ScalarFormatter) Format(x float64) string {
	if math.IsNaN(x) {
		return "NaN"
	}
	if math.IsInf(x, 1) {
		return "+Inf"
	}
	if math.IsInf(x, -1) {
		return "-Inf"
	}
	p := f.Prec
	if p < 0 {
		p = 0
	}
	var s string
	if scalarUsesScientific(x) {
		s = strconv.FormatFloat(x, 'e', p, 64)
		// normalize exponent: remove leading zeros in e+00X
		if i := strings.LastIndexByte(s, 'e'); i >= 0 && i+2 < len(s) {
			sign := s[i+1]
			exp := strings.TrimLeft(s[i+2:], "0")
			if exp == "" {
				exp = "0"
			}
			s = s[:i+2] + string(sign) + exp
		}
	} else {
		s = strconv.FormatFloat(x, 'f', p, 64)
	}
	// Trim trailing zeros and possible dot
	if strings.ContainsAny(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}
	return s
}

// LogFormatter formats tick labels on a log axis. For Base==10 it prefers
// forms like 1e3, 2e3, 5e3 when values are exact multiples. Otherwise it
// falls back to ScalarFormatter.
type LogFormatter struct{ Base float64 }

func (f LogFormatter) Format(x float64) string {
	if f.Base == 10 {
		if x <= 0 {
			return ""
		}
		k := math.Floor(math.Log10(x))
		pow := math.Pow(10, k)
		m := x / pow
		// Tolerate small rounding
		if approx(m, 1, 1e-12) {
			return "1e" + strconv.FormatFloat(k, 'f', 0, 64)
		}
		if approx(m, 2, 1e-12) {
			return "2e" + strconv.FormatFloat(k, 'f', 0, 64)
		}
		if approx(m, 5, 1e-12) {
			return "5e" + strconv.FormatFloat(k, 'f', 0, 64)
		}
	}
	// Fallback
	return (ScalarFormatter{Prec: 6}).Format(x)
}

func approx(a, b, eps float64) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d <= eps
}
