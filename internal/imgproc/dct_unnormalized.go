package imgproc

import "math"

func dctUnnormalized1D(x []float64) []float64 {
	n := len(x)
	result := make([]float64, n)
	for k := 0; k < n; k++ {
		var sum float64
		for i := 0; i < n; i++ {
			sum += x[i] * math.Cos(math.Pi*float64(k)*float64(2*i+1)/float64(2*n))
		}
		result[k] = 2 * sum
	}
	return result
}

func DCTUnnormalized2D(mat [][]float32) [][]float32 {
	h := len(mat)
	if h == 0 {
		return nil
	}

	mat64 := matf32Tof64(mat)

	for i := 0; i < h; i++ {
		mat64[i] = dctUnnormalized1D(mat64[i])
	}

	mat64 = transpose(mat64)

	w2 := len(mat64)
	for i := 0; i < w2; i++ {
		mat64[i] = dctUnnormalized1D(mat64[i])
	}

	mat64 = transpose(mat64)

	return matf64Tof32(mat64)
}
