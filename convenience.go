package imghash

import (
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	_ "image/jpeg" // register JPEG decoder
	_ "image/png"  // register PNG decoder
	"io"
	"os"
	"strings"

	"github.com/xyxu/imghash/v2/hashtype"
	"github.com/xyxu/imghash/v2/similarity"
)

// OpenImage reads and decodes an image from the given file path.
// It supports JPEG, PNG, and GIF formats.
// For GIF images, frames are composited onto a white canvas matching PIL behavior.
func OpenImage(path string) (image.Image, error) {
	if strings.HasSuffix(strings.ToLower(path), ".gif") {
		return decodeGIF(path)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	img, _, err := image.Decode(f)
	return img, err
}

// decodeGIF decodes a GIF file and composites the first frame onto a white
// canvas sized to the GIF's logical screen, matching PIL's behavior.
func decodeGIF(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	g, err := gif.DecodeAll(f)
	if err != nil {
		return nil, err
	}

	if len(g.Image) == 0 {
		return nil, ErrInvalidSize
	}

	w := g.Config.Width
	h := g.Config.Height
	if w <= 0 || h <= 0 {
		return nil, ErrInvalidSize
	}

	canvas := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
	frame := g.Image[0]
	draw.Draw(canvas, frame.Bounds(), frame, frame.Bounds().Min, draw.Over)
	return canvas, nil
}

// DecodeImage decodes an image from the given reader.
// It supports JPEG, PNG, and GIF formats.
func DecodeImage(r io.Reader) (image.Image, error) {
	img, _, err := image.Decode(r)
	return img, err
}

// HashFile is a convenience that opens an image file and computes its hash.
func HashFile(hasher Hasher, path string) (hashtype.Hash, error) {
	img, err := OpenImage(path)
	if err != nil {
		return nil, err
	}
	return hasher.Calculate(img)
}

// HashReader is a convenience that decodes an image from a reader and computes its hash.
func HashReader(hasher Hasher, r io.Reader) (hashtype.Hash, error) {
	img, err := DecodeImage(r)
	if err != nil {
		return nil, err
	}
	return hasher.Calculate(img)
}

// Compare computes the distance between two hashes.
// By default it uses the natural metric for their type: Hamming distance
// for Binary hashes, L2 (Euclidean) distance for UInt8 and Float64 hashes.
// Pass an optional DistanceFunc to override the metric, e.g.:
//
//	Compare(h1, h2, similarity.Cosine)
func Compare(h1, h2 hashtype.Hash, fn ...DistanceFunc) (similarity.Distance, error) {
	if len(fn) > 0 && fn[0] != nil {
		return fn[0](h1, h2)
	}
	_, h1Binary := h1.(hashtype.Binary)
	_, h2Binary := h2.(hashtype.Binary)
	if h1Binary || h2Binary {
		if !h1Binary || !h2Binary {
			return 0, ErrIncompatibleHash
		}
		return similarity.Hamming(h1, h2)
	}
	return similarity.L2(h1, h2)
}
