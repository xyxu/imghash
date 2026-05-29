package imgproc

import (
	"image"
	"image/color"
	"math"
)

const precisionBits = 22

var clip8Table [1280]uint8

func init() {
	for i := 0; i < len(clip8Table); i++ {
		v := i - 640
		switch {
		case v < 0:
			clip8Table[i] = 0
		case v > 255:
			clip8Table[i] = 255
		default:
			clip8Table[i] = uint8(v)
		}
	}
}

func clip8(val int) uint8 {
	return clip8Table[(val>>precisionBits)+640]
}

func lanczos(x float64) float64 {
	if x < 0 {
		x = -x
	}
	if x == 0 {
		return 1
	}
	if x >= 3 {
		return 0
	}
	pix := math.Pi * x
	return (math.Sin(pix) / pix) * (math.Sin(pix/3) / (pix / 3))
}

type coeffRow struct {
	offset int
	count  int
	coeffs []int32
}

func computeCoeffs(inSize int, outSize int) []coeffRow {
	scale := float64(inSize) / float64(outSize)
	fscale := scale
	if fscale < 1.0 {
		fscale = 1.0
	}
	support := 3.0 * fscale
	ss := 1.0 / fscale
	ksize := int(math.Ceil(support))*2 + 1

	rows := make([]coeffRow, outSize)
	for o := 0; o < outSize; o++ {
		center := (float64(o) + 0.5) * scale
		xmin := int(math.Floor(center - support + 0.5))
		if xmin < 0 {
			xmin = 0
		}
		xmax := int(math.Floor(center + support + 0.5))
		if xmax > inSize {
			xmax = inSize
		}
		count := xmax - xmin
		if count > ksize {
			count = ksize
		}

		weights := make([]float64, count)
		var wsum float64
		for x := 0; x < count; x++ {
			dist := (float64(x+xmin) + 0.5 - center) * ss
			w := lanczos(dist)
			weights[x] = w
			wsum += w
		}

		coeffs := make([]int32, ksize)
		for x := 0; x < count; x++ {
			if wsum != 0 {
				weights[x] /= wsum
			}
			coeffs[x] = int32(math.Floor(0.5 + weights[x]*float64(1<<precisionBits)))
		}
		rows[o] = coeffRow{offset: xmin, count: count, coeffs: coeffs}
	}
	return rows
}

func ResizePIL(gray *image.Gray, dstW, dstH int) *image.Gray {
	srcW := gray.Bounds().Dx()
	srcH := gray.Bounds().Dy()

	hCoef := computeCoeffs(srcW, dstW)
	vCoef := computeCoeffs(srcH, dstH)

	yboxFirst := vCoef[0].offset
	yboxLast := vCoef[len(vCoef)-1].offset + vCoef[len(vCoef)-1].count
	tempH := yboxLast - yboxFirst
	if tempH > srcH {
		tempH = srcH
	}

	temp := image.NewGray(image.Rect(0, 0, dstW, tempH))
	for yy := 0; yy < tempH; yy++ {
		srcY := yy + yboxFirst
		for xx := 0; xx < dstW; xx++ {
			c := hCoef[xx]
			acc := int32(1 << (precisionBits - 1))
			for x := 0; x < c.count; x++ {
				acc += int32(gray.GrayAt(x+c.offset, srcY).Y) * c.coeffs[x]
			}
			temp.SetGray(xx, yy, color.Gray{Y: clip8(int(acc))})
		}
	}

	dst := image.NewGray(image.Rect(0, 0, dstW, dstH))
	for yy := 0; yy < dstH; yy++ {
		c := vCoef[yy]
		adjOffset := c.offset - yboxFirst
		for xx := 0; xx < dstW; xx++ {
			acc := int32(1 << (precisionBits - 1))
			for y := 0; y < c.count; y++ {
				acc += int32(temp.GrayAt(xx, y+adjOffset).Y) * c.coeffs[y]
			}
			dst.SetGray(xx, yy, color.Gray{Y: clip8(int(acc))})
		}
	}
	return dst
}
