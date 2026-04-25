package core

import "testing"

func TestAxesParasiteUsesHostViewport(t *testing.T) {
	fig := NewFigure(500, 500)
	host := fig.AddAxes(unitRect())

	parasite := host.ParasiteAxes()
	if parasite == nil {
		t.Fatal("ParasiteAxes() returned nil")
	}
	if parasite.Host != host {
		t.Fatalf("host = %p, want %p", parasite.Host, host)
	}
	if parasite.Axes == nil {
		t.Fatal("ParasiteAxes.Axes is nil")
	}
	if parasite.Axes.RectFraction != host.RectFraction {
		t.Fatalf("parasite rect = %+v, want %+v", parasite.Axes.RectFraction, host.RectFraction)
	}
}

func TestAxesParasiteSharedAxesOptionUsesRootPeers(t *testing.T) {
	fig := NewFigure(500, 500)
	host := fig.AddAxes(unitRect())

	parasite := host.ParasiteAxes(
		WithParasiteSharedX(host),
		WithParasiteSharedY(host),
	)
	if parasite == nil {
		t.Fatal("ParasiteAxes() returned nil")
	}

	if parasite.Axes.shareX != host.xScaleRoot() {
		t.Fatal("expected parasite x scale root to be shared with host")
	}
	if parasite.Axes.shareY != host.yScaleRoot() {
		t.Fatal("expected parasite y scale root to be shared with host")
	}
	if parasite.Axes.XAxis != host.XAxis {
		t.Fatal("expected shared x axis object with host")
	}
	if parasite.Axes.YAxis != host.YAxis {
		t.Fatal("expected shared y axis object with host")
	}
}

func TestAxesParasiteWithoutSharingIsIndependent(t *testing.T) {
	fig := NewFigure(500, 500)
	host := fig.AddAxes(unitRect())

	parasite := host.ParasiteAxes()
	if parasite == nil {
		t.Fatal("ParasiteAxes() returned nil")
	}
	if parasite.Axes.shareX != nil {
		t.Fatalf("parasite shareX = %p, want nil", parasite.Axes.shareX)
	}
	if parasite.Axes.shareY != nil {
		t.Fatalf("parasite shareY = %p, want nil", parasite.Axes.shareY)
	}
}
