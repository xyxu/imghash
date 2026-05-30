package imgproc

import (
	"image"
	"math"
	"slices"
)

var _ = getSize

func DCT(mat [][]float32) [][]float32 {
	mat64 := matf32Tof64(mat)
	for i := range mat64 {
		mat64[i] = dctOrthogonal(mat64[i])
	}
	mat64 = transpose(mat64)
	for i := range mat64 {
		mat64[i] = dctOrthogonal(mat64[i])
	}
	mat64 = transpose(mat64)
	return matf64Tof32(mat64)
}

func dctOrthogonal(x []float64) []float64 {
	n := len(x)
	c0 := math.Sqrt(1.0 / float64(n))
	c1 := math.Sqrt(2.0 / float64(n))
	result := make([]float64, n)
	for k := range n {
		var sum float64
		for i := range n {
			sum += x[i] * math.Cos(math.Pi*float64(2*i+1)*float64(k)/float64(2*n))
		}
		if k == 0 {
			result[k] = sum * c0
		} else {
			result[k] = sum * c1
		}
	}
	return result
}

func matf32Tof64(mat [][]float32) [][]float64 {
	res := make([][]float64, len(mat))
	for i := range mat {
		res[i] = make([]float64, len(mat[i]))
		for j := 0; j < len(mat[i]); j++ {
			res[i][j] = float64(mat[i][j])
		}
	}
	return res
}

func matf64Tof32(mat [][]float64) [][]float32 {
	res := make([][]float32, len(mat))
	for i := range mat {
		res[i] = make([]float32, len(mat[i]))
		for j := 0; j < len(mat[i]); j++ {
			res[i][j] = float32(mat[i][j])
		}
	}
	return res
}

// --- float32 Haar DWT (used by non-WHash algorithms) ---

func HaarDWT2D(mat [][]float32, levels int) {
	rows := len(mat)
	if rows == 0 {
		return
	}
	cols := len(mat[0])
	for l := range levels {
		h := rows >> uint(l)
		w := cols >> uint(l)
		if h < 2 || w < 2 {
			break
		}
		haarRows(mat, h, w)
		haarCols(mat, h, w)
	}
}

func haarRows(mat [][]float32, rows, cols int) {
	half := cols / 2
	tmp := make([]float32, cols)
	for r := range rows {
		for c := range half {
			tmp[c] = (mat[r][2*c] + mat[r][2*c+1]) / 2
			tmp[half+c] = (mat[r][2*c] - mat[r][2*c+1]) / 2
		}
		copy(mat[r][:cols], tmp)
	}
}

func haarCols(mat [][]float32, rows, cols int) {
	half := rows / 2
	tmp := make([]float32, rows)
	for c := range cols {
		for r := range half {
			tmp[r] = (mat[2*r][c] + mat[2*r+1][c]) / 2
			tmp[half+r] = (mat[2*r][c] - mat[2*r+1][c]) / 2
		}
		for r := range rows {
			mat[r][c] = tmp[r]
		}
	}
}

func InverseHaarDWT2D(mat [][]float32, levels int) {
	rows := len(mat)
	if rows == 0 {
		return
	}
	cols := len(mat[0])
	for l := levels - 1; l >= 0; l-- {
		h := rows >> uint(l)
		w := cols >> uint(l)
		if h < 2 || w < 2 {
			continue
		}
		inverseHaarCols(mat, h, w)
		inverseHaarRows(mat, h, w)
	}
}

func inverseHaarRows(mat [][]float32, rows, cols int) {
	half := cols / 2
	tmp := make([]float32, cols)
	for r := range rows {
		copy(tmp, mat[r][:cols])
		for c := range half {
			avg := tmp[c]
			diff := tmp[half+c]
			mat[r][2*c] = avg + diff
			mat[r][2*c+1] = avg - diff
		}
	}
}

func inverseHaarCols(mat [][]float32, rows, cols int) {
	half := rows / 2
	for c := range cols {
		tmp := make([]float32, rows)
		for r := range rows {
			tmp[r] = mat[r][c]
		}
		for r := range half {
			avg := tmp[r]
			diff := tmp[half+r]
			mat[2*r][c] = avg + diff
			mat[2*r+1][c] = avg - diff
		}
	}
}

// --- float64 Haar DWT (used by WHash for Python compatibility) ---

var invSqrt2 = 1.0 / math.Sqrt2

func HaarDWT2DF64(mat [][]float64, levels int) {
	rows := len(mat)
	if rows == 0 {
		return
	}
	cols := len(mat[0])
	for l := range levels {
		h := rows >> uint(l)
		w := cols >> uint(l)
		if h < 2 || w < 2 {
			break
		}
		haar2DF64level(mat, h, w)
	}
}

func haar2DF64level(mat [][]float64, rows, cols int) {
	halfR := rows / 2
	tmpC := make([]float64, rows)
	for c := range cols {
		for r := range halfR {
			a := mat[2*r][c]
			b := mat[2*r+1][c]
			tmpC[r] = a*invSqrt2 + b*invSqrt2
			tmpC[halfR+r] = a*invSqrt2 - b*invSqrt2
		}
		for r := range rows {
			mat[r][c] = tmpC[r]
		}
	}

	halfC := cols / 2
	tmpR := make([]float64, cols)
	for r := range rows {
		for c := range halfC {
			a := mat[r][2*c]
			b := mat[r][2*c+1]
			tmpR[c] = a*invSqrt2 + b*invSqrt2
			tmpR[halfC+c] = a*invSqrt2 - b*invSqrt2
		}
		copy(mat[r][:cols], tmpR)
	}
}

func InverseHaarDWT2DF64(mat [][]float64, levels int) {
	rows := len(mat)
	if rows == 0 {
		return
	}
	cols := len(mat[0])
	for l := levels - 1; l >= 0; l-- {
		h := rows >> uint(l)
		w := cols >> uint(l)
		if h < 2 || w < 2 {
			continue
		}
		inverseHaar2DF64level(mat, h, w)
	}
}

func inverseHaar2DF64level(mat [][]float64, rows, cols int) {
	halfC := cols / 2
	tmpR := make([]float64, cols)
	for r := range rows {
		copy(tmpR, mat[r][:cols])
		for c := range halfC {
			avg := tmpR[c]
			diff := tmpR[halfC+c]
			mat[r][2*c] = avg*invSqrt2 + diff*invSqrt2
			mat[r][2*c+1] = avg*invSqrt2 - diff*invSqrt2
		}
	}

	halfR := rows / 2
	for c := range cols {
		tmpC := make([]float64, rows)
		for r := range rows {
			tmpC[r] = mat[r][c]
		}
		for r := range halfR {
			avg := tmpC[r]
			diff := tmpC[halfR+r]
			mat[2*r][c] = avg*invSqrt2 + diff*invSqrt2
			mat[2*r+1][c] = avg*invSqrt2 - diff*invSqrt2
		}
	}
}

func GrayToF64(img *image.Gray) [][]float64 {
	bounds := img.Bounds()
	width, height := getSize(img)
	f64Img := make([][]float64, height)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		f64Img[y-bounds.Min.Y] = make([]float64, width)
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pixel := img.GrayAt(x, y).Y
			f64Img[y-bounds.Min.Y][x-bounds.Min.X] = float64(pixel)
		}
	}
	return f64Img
}

func MedianF64(mat [][]float64) float64 {
	var n int
	for _, row := range mat {
		n += len(row)
	}
	vals := make([]float64, 0, n)
	for _, row := range mat {
		vals = append(vals, row...)
	}
	slices.Sort(vals)
	if n%2 == 0 {
		return (vals[n/2-1] + vals[n/2]) / 2
	}
	return vals[n/2]
}

func transpose(mat [][]float64) [][]float64 {
	height := len(mat)
	width := len(mat[0])
	res := make([][]float64, height)
	for i := range height {
		res[i] = make([]float64, width)
	}
	for i := range height {
		for j := range width {
			res[i][j] = mat[j][i]
		}
	}
	return res
}
