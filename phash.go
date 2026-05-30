package imghash

import (
	"image"

	"github.com/xyxu/imghash/v2/hashtype"
	"github.com/xyxu/imghash/v2/internal/imgproc"
	"github.com/xyxu/imghash/v2/similarity"
)

// Size of the low-frequency DCT coefficient block used by PHash.
// An 8x8 block produces a 64-bit (8-byte) binary hash.
const dctCoefSize = 8

const phashHighfreqFactor = 4

// PHash is a perceptual hash that matches Python imagehash.phash.
// It uses the same algorithm: grayscale → PIL-compatible Lanczos resize
// (hash_size × highfreq_factor) → unnormalized 2D DCT → median threshold.
//
// See https://www.researchgate.net/publication/252340846_Rihamark_Perceptual_image_hash_benchmarking for more information.
type PHash struct {
	baseConfig
	// Per-byte weights for weighted Hamming distance.
	weights  []float64
	distFunc DistanceFunc
}

// NewPHash creates a new PHash matching Python imagehash.phash.
// Default hash_size is 8 (from 32/4), producing a 64-bit hash.
func NewPHash(opts ...PHashOption) (PHash, error) {
	p := PHash{
		baseConfig: baseConfig{width: 32, height: 32, interp: Bilinear},
	}
	for _, o := range opts {
		o.applyPHash(&p)
	}
	if err := p.validate(); err != nil {
		return PHash{}, err
	}
	if p.weights == nil {
		p.weights = make([]float64, dctCoefSize)
		for i := range p.weights {
			p.weights[i] = 1
		}
	}
	return p, nil
}

// Calculate returns a perceptual image hash matching Python imagehash.phash.
func (ph PHash) Calculate(img image.Image) (hashtype.Hash, error) {
	g, err := imgproc.Grayscale(img)
	if err != nil {
		return nil, err
	}
	r := imgproc.ResizePIL(g, int(ph.width), int(ph.height))
	fImg := imgproc.GrayToF32(r)
	dctImg := imgproc.DCTUnnormalized2D(fImg)
	hashSize := int(ph.width / phashHighfreqFactor)
	tLeft := ph.topLeft(dctImg, hashSize)
	med := imgproc.MedianF32(tLeft)
	bitImg := ph.compare(tLeft, med)
	return ph.computeHash(bitImg, hashSize), nil
}

// Computes the binary hash based on the binary image supplied.
func (ph PHash) computeHash(img [][]float32, hashSize int) hashtype.Binary {
	hash := make(hashtype.Binary, hashSize)
	var c uint
	for i := range img {
		for j := range img[i] {
			if img[i][j] != 0 {
				_ = hash.SetReverse(c)
			}
			c++
		}
	}
	return hash
}

// Extract top left block from supplied image.
func (ph PHash) topLeft(img [][]float32, size int) [][]float32 {
	tL := make([][]float32, size)
	for i := range tL {
		tL[i] = img[i][0:size]
	}
	return tL
}

// Build a binary image by comparing the value to the supplied image.
func (ph PHash) compare(img [][]float32, val float32) [][]float32 {
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

// Compare computes the weighted Hamming distance between two PHash hashes
// using the per-byte weights configured on this hasher.
func (ph PHash) Compare(h1, h2 hashtype.Hash) (similarity.Distance, error) {
	if err := validateBinaryCompareInputs(h1, h2); err != nil {
		return 0, err
	}
	if ph.distFunc != nil {
		return ph.distFunc(h1, h2)
	}
	return similarity.WeightedHamming(h1, h2, ph.weights)
}
