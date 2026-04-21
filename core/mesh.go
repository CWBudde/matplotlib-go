package core

import (
	"math"

	"matplotlib-go/internal/geom"
	"matplotlib-go/render"
)

// MeshOptions configures rectilinear mesh plots such as pcolor/pcolormesh.
type MeshOptions struct {
	XEdges    []float64
	YEdges    []float64
	Colormap  *string
	VMin      *float64
	VMax      *float64
	Alpha     *float64
	EdgeColor *render.Color
	EdgeWidth *float64
	Label     string
}

// Hist2DOptions configures 2D histogram binning and rendering.
type Hist2DOptions struct {
	XBins     int
	YBins     int
	XBinEdges []float64
	YBinEdges []float64
	Colormap  *string
	VMin      *float64
	VMax      *float64
	Alpha     *float64
	EdgeColor *render.Color
	EdgeWidth *float64
	Label     string
}

// Hist2DResult stores the rendered mesh and the computed counts/edges.
type Hist2DResult struct {
	Mesh   *QuadMesh
	Counts [][]float64
	XEdges []float64
	YEdges []float64
}

// PColor renders a scalar matrix as a rectilinear quad mesh.
func (a *Axes) PColor(data [][]float64, opts ...MeshOptions) *QuadMesh {
	return a.PColorMesh(data, opts...)
}

// PColorMesh renders a scalar matrix as a rectilinear quad mesh.
func (a *Axes) PColorMesh(data [][]float64, opts ...MeshOptions) *QuadMesh {
	rows, cols, ok := finiteMatrixSize(data)
	if !ok {
		return nil
	}

	var opt MeshOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	xEdges := resolvedMeshEdges(opt.XEdges, cols)
	yEdges := resolvedMeshEdges(opt.YEdges, rows)
	if len(xEdges) != cols+1 || len(yEdges) != rows+1 {
		return nil
	}

	cmap := ""
	if opt.Colormap != nil {
		cmap = *opt.Colormap
	}
	mapping := resolveScalarMapGrid(data, cmap, opt.VMin, opt.VMax)
	alpha := meshAlpha(opt.Alpha)
	edgeWidth := 0.0
	if opt.EdgeWidth != nil {
		edgeWidth = *opt.EdgeWidth
	}

	edgeColor := render.Color{}
	if opt.EdgeColor != nil {
		edgeColor = *opt.EdgeColor
	}

	cellCount := rows * cols
	faceColors := make([]render.Color, 0, cellCount)
	edgeColors := make([]render.Color, 0, cellCount)
	for _, row := range data {
		for _, value := range row {
			if !isFinite(value) {
				faceColors = append(faceColors, render.Color{})
				edgeColors = append(edgeColors, render.Color{})
				continue
			}
			faceColors = append(faceColors, mapping.Color(value, alpha))
			edgeColors = append(edgeColors, edgeColor)
		}
	}

	mesh := &QuadMesh{
		PatchCollection: PatchCollection{
			Collection: Collection{
				Coords:   Coords(CoordData),
				Label:    opt.Label,
				Alpha:    1,
				Colormap: mapping.Colormap,
				VMin:     mapping.VMin,
				VMax:     mapping.VMax,
			},
			FaceColors: faceColors,
			EdgeColors: edgeColors,
			EdgeWidth:  edgeWidth,
			LineJoin:   render.JoinMiter,
			LineCap:    render.CapButt,
		},
		XEdges: append([]float64(nil), xEdges...),
		YEdges: append([]float64(nil), yEdges...),
	}
	a.Add(mesh)
	return mesh
}

// Hist2D bins paired samples into a 2D count matrix and renders the result as
// a QuadMesh.
func (a *Axes) Hist2D(x, y []float64, opts ...Hist2DOptions) *Hist2DResult {
	if len(x) == 0 || len(y) == 0 {
		return nil
	}

	var opt Hist2DOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	n := len(x)
	if len(y) < n {
		n = len(y)
	}
	if n == 0 {
		return nil
	}

	xData := make([]float64, 0, n)
	yData := make([]float64, 0, n)
	for i := 0; i < n; i++ {
		if !isFinite(x[i]) || !isFinite(y[i]) {
			continue
		}
		xData = append(xData, x[i])
		yData = append(yData, y[i])
	}
	if len(xData) == 0 {
		return nil
	}

	xBins := opt.XBins
	if xBins <= 0 {
		xBins = 10
	}
	yBins := opt.YBins
	if yBins <= 0 {
		yBins = 10
	}

	xEdges := resolvedHistogramEdges(xData, xBins, opt.XBinEdges)
	yEdges := resolvedHistogramEdges(yData, yBins, opt.YBinEdges)
	if len(xEdges) < 2 || len(yEdges) < 2 {
		return nil
	}

	counts := make([][]float64, len(yEdges)-1)
	for row := range counts {
		counts[row] = make([]float64, len(xEdges)-1)
	}
	for i := range xData {
		xBin := findBin(xData[i], xEdges)
		yBin := findBin(yData[i], yEdges)
		if xBin < 0 || yBin < 0 {
			continue
		}
		counts[yBin][xBin]++
	}

	meshOpt := MeshOptions{
		XEdges:    xEdges,
		YEdges:    yEdges,
		Colormap:  opt.Colormap,
		VMin:      opt.VMin,
		VMax:      opt.VMax,
		Alpha:     opt.Alpha,
		EdgeColor: opt.EdgeColor,
		EdgeWidth: opt.EdgeWidth,
		Label:     opt.Label,
	}
	mesh := a.PColorMesh(counts, meshOpt)
	if mesh == nil {
		return nil
	}
	return &Hist2DResult{
		Mesh:   mesh,
		Counts: counts,
		XEdges: append([]float64(nil), xEdges...),
		YEdges: append([]float64(nil), yEdges...),
	}
}

func resolvedMeshEdges(edges []float64, cellCount int) []float64 {
	if len(edges) > 0 {
		return append([]float64(nil), edges...)
	}
	out := make([]float64, cellCount+1)
	for i := range out {
		out[i] = float64(i)
	}
	return out
}

func resolvedHistogramEdges(data []float64, bins int, explicit []float64) []float64 {
	if len(explicit) > 1 {
		return append([]float64(nil), explicit...)
	}
	return computeBinEdges(data, bins, BinStrategyAuto)
}

func meshAlpha(alpha *float64) float64 {
	if alpha == nil {
		return 1
	}
	return clampOneToOne(*alpha)
}

func quadMeshCellCenters(xEdges, yEdges []float64) []geom.Pt {
	centers := make([]geom.Pt, 0, max(0, (len(xEdges)-1)*(len(yEdges)-1)))
	for yi := 0; yi+1 < len(yEdges); yi++ {
		centerY := (yEdges[yi] + yEdges[yi+1]) * 0.5
		for xi := 0; xi+1 < len(xEdges); xi++ {
			centers = append(centers, geom.Pt{
				X: (xEdges[xi] + xEdges[xi+1]) * 0.5,
				Y: centerY,
			})
		}
	}
	return centers
}

func meshValueAverage(values []float64) float64 {
	sum := 0.0
	count := 0
	for _, value := range values {
		if !isFinite(value) {
			continue
		}
		sum += value
		count++
	}
	if count == 0 {
		return math.NaN()
	}
	return sum / float64(count)
}
