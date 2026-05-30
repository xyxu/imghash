package imghash

import (
	"image"

	"github.com/xyxu/imghash/v2/hashtype"
	"github.com/xyxu/imghash/v2/internal/imgproc"
	"github.com/xyxu/imghash/v2/similarity"
)

const dctCoefSize = 8

type PHash struct {
	distFunc DistanceFunc
}

func NewPHash() PHash {
	return PHash{}
}

func (ph PHash) Calculate(img image.Image) (hashtype.Hash, error) {
	g, err := imgproc.Grayscale(img)
	if err != nil {
		return nil, err
	}
	g = imgproc.ResizePIL(g, 32, 32)
	fImg := imgproc.GrayToF32(g)
	dctImg := imgproc.DCTUnnormalized2D(fImg)
	lowFreq := phashTopLeft(dctImg)
	med := imgproc.MedianF32(lowFreq)
	bitImg := phashCompare(lowFreq, med)
	return phashCompute(bitImg), nil
}

func (ph PHash) Compare(h1, h2 hashtype.Hash) (similarity.Distance, error) {
	if err := validateBinaryCompareInputs(h1, h2); err != nil {
		return 0, err
	}
	if ph.distFunc != nil {
		return ph.distFunc(h1, h2)
	}
	return similarity.Hamming(h1, h2)
}

func phashTopLeft(img [][]float32) [][]float32 {
	tL := make([][]float32, dctCoefSize)
	for i := range tL {
		tL[i] = img[i][0:dctCoefSize]
	}
	return tL
}

func phashCompare(img [][]float32, val float32) [][]float32 {
	bit := make([][]float32, len(img))
	for i := range img {
		bit[i] = make([]float32, len(img[i]))
		for j := range img[i] {
			if img[i][j] > val {
				bit[i][j] = 1
			}
		}
	}
	return bit
}

func phashCompute(img [][]float32) hashtype.Binary {
	hash := make(hashtype.Binary, dctCoefSize)
	var c uint
	for i := range img {
		for j := range img[i] {
			if img[i][j] != 0 {
				hash.SetReverse(c)
			}
			c++
		}
	}
	return hash
}
