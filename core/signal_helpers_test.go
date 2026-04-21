package core

import (
	"math"
	"testing"
)

func TestAxesSpecgramFindsDominantFrequency(t *testing.T) {
	fig := NewFigure(400, 300)
	ax := fig.AddAxes(unitRect())

	samples := sineWave(8, 64, 128, 0)
	result := ax.Specgram(samples, SpecgramOptions{
		Fs:       64,
		NFFT:     32,
		NOverlap: 16,
	})
	if result == nil {
		t.Fatal("Specgram() returned nil")
	}
	if result.Image == nil {
		t.Fatal("Specgram() should create an image")
	}
	if got, want := len(result.Frequencies), 17; got != want {
		t.Fatalf("frequency bin count = %d, want %d", got, want)
	}
	if len(result.Times) < 2 {
		t.Fatalf("time bin count = %d, want >= 2", len(result.Times))
	}

	peak := dominantFrequency(result.Frequencies, result.Spectrum)
	if math.Abs(peak-8) > 2 {
		t.Fatalf("dominant spectrogram frequency = %v, want about 8", peak)
	}
}

func TestAxesSignalAnalysisHelpers(t *testing.T) {
	x := sineWave(5, 64, 128, 0)
	y := sineWave(5, 64, 128, math.Pi/4)
	opts := SignalSpectrumOptions{
		Fs:       64,
		NFFT:     64,
		NOverlap: 32,
	}

	fig := NewFigure(400, 300)
	ax := fig.AddAxes(unitRect())
	psd := ax.PSD(x, opts)
	if psd == nil || psd.Line == nil {
		t.Fatal("PSD() returned nil")
	}
	if got := dominantLineFrequency(psd); math.Abs(got-5) > 1 {
		t.Fatalf("PSD dominant frequency = %v, want about 5", got)
	}

	fig = NewFigure(400, 300)
	ax = fig.AddAxes(unitRect())
	csd := ax.CSD(x, y, opts)
	if csd == nil || csd.Line == nil {
		t.Fatal("CSD() returned nil")
	}
	if got := dominantLineFrequency(csd); math.Abs(got-5) > 1 {
		t.Fatalf("CSD dominant frequency = %v, want about 5", got)
	}

	fig = NewFigure(400, 300)
	ax = fig.AddAxes(unitRect())
	cohere := ax.Cohere(x, y, opts)
	if cohere == nil || cohere.Line == nil {
		t.Fatal("Cohere() returned nil")
	}
	peakIndex := argmax(cohere.Values)
	if peakIndex < 0 || cohere.Values[peakIndex] < 0.9 {
		t.Fatalf("coherence peak = %v, want >= 0.9", cohere.Values[peakIndex])
	}

	fig = NewFigure(400, 300)
	ax = fig.AddAxes(unitRect())
	acorr := ax.ACorr(x, CorrelationOptions{MaxLags: 8})
	if acorr == nil || acorr.Line == nil {
		t.Fatal("ACorr() returned nil")
	}
	if got, want := len(acorr.Lags), 17; got != want {
		t.Fatalf("lag count = %d, want %d", got, want)
	}
	zeroLag := indexOf(acorr.Lags, 0)
	if zeroLag < 0 {
		t.Fatal("zero lag not found in ACorr output")
	}
	if math.Abs(acorr.Values[zeroLag]-1) > 0.05 {
		t.Fatalf("ACorr zero-lag value = %v, want about 1", acorr.Values[zeroLag])
	}
}

func dominantFrequency(freqs []float64, spectrum [][]float64) float64 {
	bestIndex := -1
	bestValue := math.Inf(-1)
	for row := range spectrum {
		sum := 0.0
		for _, value := range spectrum[row] {
			sum += value
		}
		if sum > bestValue {
			bestValue = sum
			bestIndex = row
		}
	}
	if bestIndex < 0 {
		return 0
	}
	return freqs[bestIndex]
}

func dominantLineFrequency(result *SpectrumResult) float64 {
	if result == nil {
		return 0
	}
	index := argmax(result.Values)
	if index < 0 {
		return 0
	}
	return result.Frequencies[index]
}

func argmax(values []float64) int {
	best := -1
	bestValue := math.Inf(-1)
	for i, value := range values {
		if value > bestValue {
			bestValue = value
			best = i
		}
	}
	return best
}

func indexOf(values []float64, want float64) int {
	for i, value := range values {
		if value == want {
			return i
		}
	}
	return -1
}

func sineWave(freq, fs float64, count int, phase float64) []float64 {
	out := make([]float64, count)
	for i := range out {
		t := float64(i) / fs
		out[i] = math.Sin(2*math.Pi*freq*t + phase)
	}
	return out
}
