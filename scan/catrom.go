package scan

import (
	"fmt"
	"image"
	"image/color"
	"math"
)

// Interpolate between four points a,b,c,d
// x is 0..1.0 between b and c
// adapted from ImageMagick CatromWeights() in pixel.c 20200409_121551
func catromWeights(x float64, weights []float64) {
	alpha := 1.0 - x
	beta := -0.5 * x * alpha
	weights[0] = alpha * beta
	weights[3] = x * beta
	gamma := weights[3] - weights[0]
	weights[1] = alpha - weights[0] + gamma
	weights[2] = x - weights[3] - gamma
}

func CatmullRom(a, b, c, d uint8, x float64) uint8 {
	return 0
}

var black = color.RGBA{0, 0, 0, 255}

type frgba struct {
	r, g, b, a float64
}

func (x frgba) String() string {
	return fmt.Sprintf("(%.4f,%.4f,%.4f,%.4f)", x.r, x.g, x.b, x.a)
}

func fclamp(x, min, max float64) float64 {
	if x < min {
		return min
	}
	if x > max {
		return max
	}
	return x
}

var dblimit = 10

func ImageBiCatrom(im image.Image, x, y float64) color.RGBA {
	bound := im.Bounds()
	fxf := math.Floor(x)
	fx := int(fxf)
	fyf := math.Floor(y)
	fy := int(fyf)
	if fx < bound.Min.X || fx >= bound.Max.X || fy < bound.Min.Y || fy >= bound.Max.Y {
		return black
	}
	if fx < bound.Min.X+1 || fx >= bound.Max.X-1 || fy < bound.Min.Y+1 || fy >= bound.Max.Y-1 {
		// linear interpolation at the edges.
		// TODO!
		return black
	}
	// first interpolate across each row to a point at x
	var xw [4]float64
	catromWeights(x-fxf, xw[:])
	var xpoints [4]frgba
	//var row [4]color.Color
	for iy := 0; iy < 4; iy++ {
		y := fy - 1 + iy
		for ix := 0; ix < 4; ix++ {
			r, g, b, a := im.At(fx-1+ix, y).RGBA()
			xpoints[iy].r += float64(r) * xw[ix]
			xpoints[iy].g += float64(g) * xw[ix]
			xpoints[iy].b += float64(b) * xw[ix]
			xpoints[iy].a += float64(a) * xw[ix]
		}
	}

	// fmt.Printf("x=%.4f %s %s %s %s\n", x, xpoints[0], xpoints[1], xpoints[2], xpoints[3])
	// dblimit--
	// if dblimit <= 0 {
	// 	os.Exit(1)
	// }

	// second interpolate along vertical line at x to point at y
	catromWeights(y-fyf, xw[:])
	outf := frgba{0, 0, 0, 0}
	for i := 0; i < 4; i++ {
		outf.r += xpoints[i].r * xw[i]
		outf.g += xpoints[i].g * xw[i]
		outf.b += xpoints[i].b * xw[i]
		outf.a += xpoints[i].a * xw[i]
	}
	outf.a = fclamp(outf.a/255.0, 0, 255)
	outf.r = fclamp(outf.r/255.0, 0, outf.a)
	outf.g = fclamp(outf.g/255.0, 0, outf.a)
	outf.b = fclamp(outf.b/255.0, 0, outf.a)
	return color.RGBA{uint8(outf.r), uint8(outf.g), uint8(outf.b), uint8(outf.a)}
}

func YBiCatrom(im *image.YCbCr, x, y float64) uint8 {
	bound := im.Bounds()
	fxf := math.Floor(x)
	fx := int(fxf)
	fyf := math.Floor(y)
	fy := int(fyf)
	if fx < bound.Min.X || fx >= bound.Max.X || fy < bound.Min.Y || fy >= bound.Max.Y {
		return 0
	}
	if fx < bound.Min.X+1 || fx >= bound.Max.X-1 || fy < bound.Min.Y+1 || fy >= bound.Max.Y-1 {
		// linear interpolation at the edges.
		// TODO!
		return 0
	}
	// first interpolate across each row to a point at x
	var xw [4]float64
	catromWeights(x-fxf, xw[:])
	//var xpoints [4]frgba
	var xyvals [4]float64
	//var row [4]color.Color
	for iy := 0; iy < 4; iy++ {
		y := fy - 1 + iy
		for ix := 0; ix < 4; ix++ {
			xyvals[iy] += float64(im.Y[(im.YStride*y)+(fx-1+ix)]) * xw[ix]
		}
	}

	// second interpolate along vertical line at x to point at y
	catromWeights(y-fyf, xw[:])
	outy := 0.0
	for i := 0; i < 4; i++ {
		outy += xyvals[i] * xw[i]
	}
	outb := uint8(fclamp(outy, 0, 255))
	//fmt.Printf("%f -> %d\n", outy, outb)
	return outb
}
