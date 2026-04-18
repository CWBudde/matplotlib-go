package core

import "matplotlib-go/internal/geom"

// SubplotOptions controls automatic subplot placement and axis sharing.
type SubplotOptions struct {
	// Figure-normalized margins.
	Left, Right, Bottom, Top float64

	// Normalized inter-subplot spacing.
	WSpace float64
	HSpace float64

	// ShareAxes controls whether all subplots reuse the same x/y scales and axis controls.
	ShareX bool
	ShareY bool
}

// SubplotOption configures the behavior of Subplots.
type SubplotOption func(*SubplotOptions)

// WithSubplotPadding overrides the subplot figure margins.
func WithSubplotPadding(left, right, bottom, top float64) SubplotOption {
	return func(cfg *SubplotOptions) {
		cfg.Left = left
		cfg.Right = right
		cfg.Bottom = bottom
		cfg.Top = top
	}
}

// WithSubplotSpacing overrides the spacing between subplot cells.
func WithSubplotSpacing(wspace, hspace float64) SubplotOption {
	return func(cfg *SubplotOptions) {
		cfg.WSpace = wspace
		cfg.HSpace = hspace
	}
}

// WithSubplotShareX shares x scales and x-axis settings across all subplots.
func WithSubplotShareX() SubplotOption {
	return func(cfg *SubplotOptions) {
		cfg.ShareX = true
	}
}

// WithSubplotShareY shares y scales and y-axis settings across all subplots.
func WithSubplotShareY() SubplotOption {
	return func(cfg *SubplotOptions) {
		cfg.ShareY = true
	}
}

// WithSubplotShareBoth shares x and y scales and axis settings across all subplots.
func WithSubplotShareBoth() SubplotOption {
	return func(cfg *SubplotOptions) {
		cfg.ShareX = true
		cfg.ShareY = true
	}
}

func defaultSubplotOptions() SubplotOptions {
	return SubplotOptions{
		Left:   0.10,
		Right:  0.95,
		Bottom: 0.10,
		Top:    0.90,
		WSpace: 0.05,
		HSpace: 0.06,
	}
}

// Subplots creates an nRows x nCols grid of axes with automatic layout.
func (f *Figure) Subplots(nRows, nCols int, opts ...SubplotOption) [][]*Axes {
	cfg := defaultSubplotOptions()
	for _, opt := range opts {
		opt(&cfg)
	}

	if nRows <= 0 || nCols <= 0 {
		return nil
	}
	if cfg.Right <= cfg.Left || cfg.Top <= cfg.Bottom {
		return nil
	}

	// Effective cell sizes including spacing.
	availableW := cfg.Right - cfg.Left - cfg.WSpace*float64(nCols-1)
	availableH := cfg.Top - cfg.Bottom - cfg.HSpace*float64(nRows-1)
	if availableW <= 0 || availableH <= 0 {
		return nil
	}
	cellW := availableW / float64(nCols)
	cellH := availableH / float64(nRows)

	grid := make([][]*Axes, nRows)
	var sharedX, sharedY *Axes

	for row := 0; row < nRows; row++ {
		rowAxes := make([]*Axes, nCols)
		for col := 0; col < nCols; col++ {
			minX := cfg.Left + float64(col)*(cellW+cfg.WSpace)
			maxX := minX + cellW

			maxY := cfg.Top - float64(row)*(cellH+cfg.HSpace)
			minY := maxY - cellH

			ax := f.AddAxes(geom.Rect{
				Min: geom.Pt{X: minX, Y: minY},
				Max: geom.Pt{X: maxX, Y: maxY},
			})
			if cfg.ShareX {
				if sharedX == nil {
					sharedX = ax
				} else {
					ax.shareX = sharedX
					ax.XAxis = sharedX.XAxis
				}
			}
			if cfg.ShareY {
				if sharedY == nil {
					sharedY = ax
				} else {
					ax.shareY = sharedY
					ax.YAxis = sharedY.YAxis
				}
			}
			rowAxes[col] = ax
		}
		grid[row] = rowAxes
	}

	return grid
}
