package scan

/*
Find transform
[a b c [sx  [dx
 d e f  sy = dy
 g h i]  1]   1]

Matrix A, column X, column B
Ax=b

A is data
[sx1 sy1 1 0 0 0 0 0 0
 0 0 0 sx1 sy1 1 0 0 0
 0 0 0 0 0 0 sx1 sy1 1]
... for other source points

x is the column of unknown transform factors we solve for
[a b c d e f g h i]

b is the column
[dx1 dy1 1 ...]


biga = A.T() . A
atb = A.T() . b

diagonalize [biga | atb]

*/
// go get -u -t gonum.org/v1/gonum/...

import (
	"gonum.org/v1/gonum/mat"
)

type FPoint struct {
	X float64
	Y float64
}

func FPointFromInt(x, y int) FPoint {
	return FPoint{X: float64(x), Y: float64(y)}
}

func (fp *FPoint) SetInt(x, y int) {
	fp.X = float64(x)
	fp.Y = float64(y)
}

func FindTransform(sources, dests []FPoint) []float64 {
	if len(sources) != len(dests) {
		return nil
	}
	data := make([]float64, len(sources)*9*3)
	dest := make([]float64, len(sources)*3)
	for i, sp := range sources {
		dp := dests[i]
		rp := i * 3 * 9
		data[rp+0] = sp.X
		data[rp+1] = sp.Y
		data[rp+2] = 1.0
		dest[i*3] = dp.X
		rp = ((i * 3) + 1) * 9
		data[rp+3] = sp.X
		data[rp+4] = sp.Y
		data[rp+5] = 1.0
		dest[(i*3)+1] = dp.Y
		rp = ((i * 3) + 2) * 9
		data[rp+6] = sp.X
		data[rp+7] = sp.Y
		data[rp+8] = 1.0
		dest[(i*3)+2] = 1.0
	}
	A := mat.NewDense(len(sources)*3, 9, data)
	b := mat.NewDense(len(sources)*3, 1, dest)
	var x mat.Dense
	x.Solve(A, b)
	//fmt.Printf("solution ?\nx = %v\n", mat.Formatted(&x))
	return mat.Col(nil, 0, &x)
}
