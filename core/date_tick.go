package core

import (
	"math"
	"time"
)

type DateLocator struct {
	Location *time.Location
}

func (l DateLocator) Ticks(minVal, maxVal float64, targetCount int) []float64 {
	if math.IsNaN(minVal) || math.IsNaN(maxVal) || math.IsInf(minVal, 0) || math.IsInf(maxVal, 0) {
		return nil
	}
	if minVal > maxVal {
		minVal, maxVal = maxVal, minVal
	}
	if targetCount <= 0 {
		targetCount = 6
	}

	minTime := dateNumberToTime(minVal, l.location())
	maxTime := dateNumberToTime(maxVal, l.location())
	if !maxTime.After(minTime) {
		return []float64{minVal}
	}

	interval := chooseDateTickInterval(minTime, maxTime, targetCount)
	current := interval.align(minTime)
	if current.Before(minTime) {
		current = interval.next(current)
	}

	ticks := make([]float64, 0, targetCount+2)
	guard := targetCount*4 + 16
	for i := 0; i < guard && !current.After(maxTime); i++ {
		ticks = append(ticks, timeToDateNumber(current))
		current = interval.next(current)
	}

	if len(ticks) == 0 {
		ticks = append(ticks, minVal, maxVal)
	}
	return dedupeTicks(ticks)
}

func (l DateLocator) location() *time.Location {
	if l.Location != nil {
		return l.Location
	}
	return time.UTC
}

type AutoDateFormatter struct {
	Min      float64
	Max      float64
	Location *time.Location
}

func (f AutoDateFormatter) Format(x float64) string {
	layout := chooseDateLabelLayout(f.Min, f.Max)
	return DateFormatter{Layout: layout, Location: f.location()}.Format(x)
}

func (f AutoDateFormatter) location() *time.Location {
	if f.Location != nil {
		return f.Location
	}
	return time.UTC
}

type DateFormatter struct {
	Layout   string
	Location *time.Location
}

func (f DateFormatter) Format(x float64) string {
	layout := f.Layout
	if layout == "" {
		layout = time.RFC3339
	}
	return dateNumberToTime(x, f.location()).Format(layout)
}

func (f DateFormatter) location() *time.Location {
	if f.Location != nil {
		return f.Location
	}
	return time.UTC
}

type dateTickInterval struct {
	unit string
	step int
}

func chooseDateTickInterval(minTime, maxTime time.Time, targetCount int) dateTickInterval {
	if targetCount <= 0 {
		targetCount = 6
	}
	span := maxTime.Sub(minTime)
	if span <= 0 {
		return dateTickInterval{unit: "day", step: 1}
	}

	raw := span / time.Duration(targetCount)
	candidates := []struct {
		interval dateTickInterval
		approx   time.Duration
	}{
		{dateTickInterval{unit: "second", step: 1}, time.Second},
		{dateTickInterval{unit: "second", step: 5}, 5 * time.Second},
		{dateTickInterval{unit: "second", step: 15}, 15 * time.Second},
		{dateTickInterval{unit: "second", step: 30}, 30 * time.Second},
		{dateTickInterval{unit: "minute", step: 1}, time.Minute},
		{dateTickInterval{unit: "minute", step: 5}, 5 * time.Minute},
		{dateTickInterval{unit: "minute", step: 15}, 15 * time.Minute},
		{dateTickInterval{unit: "minute", step: 30}, 30 * time.Minute},
		{dateTickInterval{unit: "hour", step: 1}, time.Hour},
		{dateTickInterval{unit: "hour", step: 3}, 3 * time.Hour},
		{dateTickInterval{unit: "hour", step: 6}, 6 * time.Hour},
		{dateTickInterval{unit: "hour", step: 12}, 12 * time.Hour},
		{dateTickInterval{unit: "day", step: 1}, 24 * time.Hour},
		{dateTickInterval{unit: "day", step: 2}, 48 * time.Hour},
		{dateTickInterval{unit: "day", step: 7}, 7 * 24 * time.Hour},
		{dateTickInterval{unit: "day", step: 14}, 14 * 24 * time.Hour},
		{dateTickInterval{unit: "month", step: 1}, 30 * 24 * time.Hour},
		{dateTickInterval{unit: "month", step: 3}, 90 * 24 * time.Hour},
		{dateTickInterval{unit: "month", step: 6}, 180 * 24 * time.Hour},
		{dateTickInterval{unit: "year", step: 1}, 365 * 24 * time.Hour},
		{dateTickInterval{unit: "year", step: 2}, 2 * 365 * 24 * time.Hour},
		{dateTickInterval{unit: "year", step: 5}, 5 * 365 * 24 * time.Hour},
		{dateTickInterval{unit: "year", step: 10}, 10 * 365 * 24 * time.Hour},
	}

	for _, candidate := range candidates {
		if raw <= candidate.approx {
			return candidate.interval
		}
	}
	return dateTickInterval{unit: "year", step: 10}
}

func (i dateTickInterval) align(t time.Time) time.Time {
	switch i.unit {
	case "year":
		year := (t.Year() / i.step) * i.step
		return time.Date(year, time.January, 1, 0, 0, 0, 0, t.Location())
	case "month":
		monthIndex := int(t.Month()) - 1
		aligned := (monthIndex / i.step) * i.step
		return time.Date(t.Year(), time.Month(aligned+1), 1, 0, 0, 0, 0, t.Location())
	case "day":
		y, m, d := t.Date()
		d = ((d - 1) / i.step * i.step) + 1
		return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
	case "hour":
		y, m, d := t.Date()
		hour := (t.Hour() / i.step) * i.step
		return time.Date(y, m, d, hour, 0, 0, 0, t.Location())
	case "minute":
		y, m, d := t.Date()
		minute := (t.Minute() / i.step) * i.step
		return time.Date(y, m, d, t.Hour(), minute, 0, 0, t.Location())
	default:
		y, m, d := t.Date()
		second := (t.Second() / i.step) * i.step
		return time.Date(y, m, d, t.Hour(), t.Minute(), second, 0, t.Location())
	}
}

func (i dateTickInterval) next(t time.Time) time.Time {
	switch i.unit {
	case "year":
		return t.AddDate(i.step, 0, 0)
	case "month":
		return t.AddDate(0, i.step, 0)
	case "day":
		return t.AddDate(0, 0, i.step)
	case "hour":
		return t.Add(time.Duration(i.step) * time.Hour)
	case "minute":
		return t.Add(time.Duration(i.step) * time.Minute)
	default:
		return t.Add(time.Duration(i.step) * time.Second)
	}
}

func chooseDateLabelLayout(minVal, maxVal float64) string {
	span := math.Abs(maxVal - minVal)
	switch {
	case span >= 2*365*24*3600:
		return "2006"
	case span >= 90*24*3600:
		return "Jan 2006"
	case span >= 2*24*3600:
		return "02 Jan"
	case span >= 24*3600:
		return "02 Jan 15:04"
	case span >= 60:
		return "15:04"
	default:
		return "15:04:05"
	}
}

func timeToDateNumber(t time.Time) float64 {
	t = t.UTC()
	return float64(t.Unix()) + float64(t.Nanosecond())/1e9
}

func dateNumberToTime(v float64, loc *time.Location) time.Time {
	if loc == nil {
		loc = time.UTC
	}
	sec, frac := math.Modf(v)
	nsec := int64(math.Round(frac * 1e9))
	return time.Unix(int64(sec), nsec).In(loc)
}
