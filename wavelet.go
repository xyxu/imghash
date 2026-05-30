package imghash

import (
	"image"
	"math"

	"github.com/xyxu/imghash/v2/hashtype"
	"github.com/xyxu/imghash/v2/internal/imgproc"
	"github.com/xyxu/imghash/v2/similarity"
)

// WHash is a perceptual hash based on the Haar wavelet transform.
// It matches the Python imagehash.whash implementation:
// grayscale → PIL-compatible Lanczos resize → /255 normalization →
// optional max Haar LL removal → Haar DWT → median threshold.
//
// See https://fullstackml.com/wavelet-image-hash-in-python-3504571f3b08 for more information.
type WHash struct {
	baseConfig
	removeMaxHaarLL bool
	level           int // kept for backward compat; not used in Python-compatible mode
	distFunc        DistanceFunc
}

// NewWHash creates a new WHash with the given options.
// Without options, uses 8×8 hash, removeMaxHaarLL=true matching Python imagehash.
func NewWHash(opts ...WHashOption) (WHash, error) {
	w := WHash{
		baseConfig:      baseConfig{width: 8, height: 8, interp: Bilinear},
		removeMaxHaarLL: true,
	}
	for _, o := range opts {
		o.applyWHash(&w)
	}
	if err := w.validate(); err != nil {
		return WHash{}, err
	}
	if w.width&(w.width-1) != 0 || w.width == 0 {
		return WHash{}, ErrInvalidSize
	}
	if w.width != w.height {
		return WHash{}, ErrInvalidSize
	}
	return w, nil
}

// Calculate returns a perceptual image hash matching Python imagehash.whash.
func (wh WHash) Calculate(img image.Image) (hashtype.Hash, error) {
	g, err := imgproc.Grayscale(img)
	if err != nil {
		return nil, err
	}

	minDim := min(g.Bounds().Dx(), g.Bounds().Dy())
	imageScale := 1
	for imageScale*2 <= minDim {
		imageScale *= 2
	}
	if imageScale < int(wh.width) {
		imageScale = int(wh.width)
	}

	r := imgproc.ResizePIL(g, imageScale, imageScale)
	mat := imgproc.GrayToF64(r)

	for i := range mat {
		for j := range mat[i] {
			mat[i][j] /= 255
		}
	}

	llMaxLevel := int(math.Log2(float64(imageScale)))
	level := int(math.Log2(float64(wh.width)))
	dwtLevel := llMaxLevel - level

	if wh.removeMaxHaarLL {
		imgproc.HaarDWT2DF64(mat, llMaxLevel)
		mat[0][0] = 0
		imgproc.InverseHaarDWT2DF64(mat, llMaxLevel)
	}

	imgproc.HaarDWT2DF64(mat, dwtLevel)

	ll := wh.extractLL(mat)
	med := wh.medianF64(ll)
	hash, err := wh.computeHashF64(ll, med)
	if err != nil {
		return nil, err
	}
	return hash, nil
}

func (wh WHash) extractLL(mat [][]float64) [][]float64 {
	ll := make([][]float64, wh.height)
	for r := uint(0); r < wh.height; r++ {
		ll[r] = make([]float64, wh.width)
		copy(ll[r], mat[r][:wh.width])
	}
	return ll
}

func (wh WHash) medianF64(mat [][]float64) float64 {
	return imgproc.MedianF64(mat)
}

func (wh WHash) computeHashF64(ll [][]float64, median float64) (hashtype.Binary, error) {
	hash := hashtype.NewBinary(wh.width * wh.height)
	var c uint
	for _, row := range ll {
		for _, v := range row {
			if v > median {
				if err := hash.SetReverse(c); err != nil {
					return nil, err
				}
			}
			c++
		}
	}
	return hash, nil
}

// Compare computes the Hamming distance between two WHash hashes.
func (wh WHash) Compare(h1, h2 hashtype.Hash) (similarity.Distance, error) {
	if err := validateBinaryCompareInputs(h1, h2); err != nil {
		return 0, err
	}
	if wh.distFunc != nil {
		return wh.distFunc(h1, h2)
	}
	return similarity.Hamming(h1, h2)
}
