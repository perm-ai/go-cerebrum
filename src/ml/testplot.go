package ml

import (
	"image/color"
	"math/rand"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

func TestPlot(x []float64, y []float64, target []float64) {
	var xA = make([]float64, len(x))
	var yA = make([]float64, len(x))
	PlotA := 0
	var xB = make([]float64, len(x))
	var yB = make([]float64, len(x))
	PlotB := 0
	for i := 0; i < len(x); i++ {
		if target[i] == 1 {
			xA[PlotA] = x[i]
			yA[PlotA] = y[i]
			PlotA++
		} else {
			xB[PlotA] = x[i]
			yB[PlotA] = y[i]
			PlotB++
		}
	}
	scatterDataA := assignPoints(xA, yA)
	scatterDataB := assignPoints(xB, yB)
	// Create a new plot, set its title and
	// axis labels.
	p := plot.New()

	p.Title.Text = "Points Example"
	p.X.Label.Text = "X"
	p.Y.Label.Text = "Y"
	// Draw a grid behind the data
	p.Add(plotter.NewGrid())

	// Make a scatter plotterA
	s1, err := plotter.NewScatter(scatterDataA)
	if err != nil {
		panic(err)
	}
	s1.GlyphStyle.Color = color.RGBA{R: 255, B: 128, A: 255}
	//B
	s2, err := plotter.NewScatter(scatterDataB)
	if err != nil {
		panic(err)
	}
	s2.GlyphStyle.Color = color.RGBA{R: 128, B: 255, A: 255}
	// Add the plotters to the plot, with a legend
	// entry for each
	p.Add(s1, s2)
	p.Legend.Add("scatterA", s1)
	p.Legend.Add("scatterB", s2)

	// Save the plot to a PNG file.
	if err := p.Save(4*vg.Inch, 4*vg.Inch, "points.png"); err != nil {
		panic(err)
	}
}

func randomPoints(n int) plotter.XYs {
	pts := make(plotter.XYs, n)
	for i := range pts {
		if i == 0 {
			pts[i].X = rand.Float64()
		} else {
			pts[i].X = pts[i-1].X + rand.Float64()
		}
		pts[i].Y = pts[i].X + 10*rand.Float64()
	}
	return pts
}
func assignPoints(x []float64, y []float64) plotter.XYs {
	pts := make(plotter.XYs, len(x))
	for i := range pts {
		pts[i].X = x[i]
		pts[i].Y = y[i]
	}
	return pts
}
