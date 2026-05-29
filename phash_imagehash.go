package imghash

import (
	"image"

	"github.com/ajdnik/imghash/v2/hashtype"
	"github.com/ajdnik/imghash/v2/internal/imgproc"
	"github.com/ajdnik/imghash/v2/similarity"
)

const pHashImgHashCoefSize = 8

type PHashImageHash struct {
	distFunc DistanceFunc
}

func NewPHashImageHash() PHashImageHash {
	return PHashImageHash{}
}

func (ph PHashImageHash) Calculate(img image.Image) (hashtype.Hash, error) {
	g, err := imgproc.Grayscale(img)
	if err != nil {
		return nil, err
	}
	g = imgproc.ResizePIL(g, 32, 32)
	fImg := imgproc.GrayToF32(g)
	dctImg := imgproc.DCTUnnormalized2D(fImg)
	lowFreq := pHashImgHashTopLeft(dctImg)
	med := imgproc.MedianF32(lowFreq)
	bitImg := pHashImgHashCompare(lowFreq, med)
	return pHashImgHashCompute(bitImg), nil
}

func (ph PHashImageHash) Compare(h1, h2 hashtype.Hash) (similarity.Distance, error) {
	if err := validateBinaryCompareInputs(h1, h2); err != nil {
		return 0, err
	}
	if ph.distFunc != nil {
		return ph.distFunc(h1, h2)
	}
	return similarity.Hamming(h1, h2)
}

func pHashImgHashTopLeft(img [][]float32) [][]float32 {
	tL := make([][]float32, pHashImgHashCoefSize)
	for i := range tL {
		tL[i] = img[i][0:pHashImgHashCoefSize]
	}
	return tL
}

func pHashImgHashCompare(img [][]float32, val float32) [][]float32 {
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

func pHashImgHashCompute(img [][]float32) hashtype.Binary {
	hash := make(hashtype.Binary, pHashImgHashCoefSize)
	var c uint
	for i := range img {
		for j := range img[i] {
			if img[i][j] != 0 {
				hash.Set(c)
			}
			c++
		}
	}
	return hash
}
