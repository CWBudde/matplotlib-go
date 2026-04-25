package core

// ParasiteAxes wraps an overlay Axes that shares its viewport with a host Axes.
type ParasiteAxes struct {
	Host *Axes
	Axes *Axes
}

type parasiteAxesOptions struct {
	shareX *Axes
	shareY *Axes
}

// ParasiteAxesOption configures parasite-axes creation.
type ParasiteAxesOption func(*parasiteAxesOptions)

// WithParasiteSharedX reuses x-axis scale and x-axis state from peer.
func WithParasiteSharedX(peer *Axes) ParasiteAxesOption {
	return func(cfg *parasiteAxesOptions) {
		cfg.shareX = peer
	}
}

// WithParasiteSharedY reuses y-axis scale and y-axis state from peer.
func WithParasiteSharedY(peer *Axes) ParasiteAxesOption {
	return func(cfg *parasiteAxesOptions) {
		cfg.shareY = peer
	}
}

// WithParasiteSharedAxes reuses both x and y states from peer.
func WithParasiteSharedAxes(peer *Axes) ParasiteAxesOption {
	return func(cfg *parasiteAxesOptions) {
		cfg.shareX = peer
		cfg.shareY = peer
	}
}

// ParasiteAxes creates an overlay Axes sharing the same viewport as the host.
func (a *Axes) ParasiteAxes(opts ...ParasiteAxesOption) *ParasiteAxes {
	if a == nil || a.figure == nil {
		return nil
	}

	cfg := parasiteAxesOptions{}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	parasite := a.newOverlayAxes()
	if parasite == nil {
		return nil
	}

	if cfg.shareX != nil {
		root := cfg.shareX.xScaleRoot()
		if root != nil {
			parasite.shareX = root
			parasite.XAxis = root.XAxis
		}
	}
	if cfg.shareY != nil {
		root := cfg.shareY.yScaleRoot()
		if root != nil {
			parasite.shareY = root
			parasite.YAxis = root.YAxis
		}
	}

	parasite.ShowFrame = false
	return &ParasiteAxes{
		Host: a,
		Axes: parasite,
	}
}

// NewParasiteAxes creates an overlay Axes for the requested host on this figure.
func (f *Figure) NewParasiteAxes(host *Axes, opts ...ParasiteAxesOption) *ParasiteAxes {
	if f == nil || host == nil || host.figure != f {
		return nil
	}
	return host.ParasiteAxes(opts...)
}
