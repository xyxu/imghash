package imghash_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/xyxu/imghash/v2"
	"github.com/xyxu/imghash/v2/hashtype"
	"github.com/xyxu/imghash/v2/similarity"
)

var wHashCalculateTests = []struct {
	filename string
	hash     hashtype.Binary
	width    uint
	height   uint
}{
	{"assets/lena.jpg", hashtype.Binary{190, 152, 189, 137, 11, 11, 143, 140}, 8, 8},
	{"assets/baboon.jpg", hashtype.Binary{1, 195, 63, 60, 188, 152, 226, 188}, 8, 8},
	{"assets/cat.jpg", hashtype.Binary{255, 255, 248, 224, 192, 128, 0, 244}, 8, 8},
	{"assets/monarch.jpg", hashtype.Binary{128, 144, 200, 63, 253, 251, 103, 2}, 8, 8},
	{"assets/peppers.jpg", hashtype.Binary{239, 135, 97, 45, 124, 76, 68, 225}, 8, 8},
	{"assets/tulips.jpg", hashtype.Binary{176, 102, 58, 90, 95, 94, 60, 96}, 8, 8},
}

func TestWHash_Calculate(t *testing.T) {
	for _, tt := range wHashCalculateTests {
		t.Run(tt.filename, func(t *testing.T) {
			hash, err := imghash.NewWHash(imghash.WithSize(tt.width, tt.height))
			if err != nil {
				t.Fatalf("failed to create hasher: %v", err)
			}
			img, err := imghash.OpenImage(tt.filename)
			if err != nil {
				t.Fatalf("failed to open %s: %v", tt.filename, err)
			}
			result, err := hash.Calculate(img)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			res := result.(hashtype.Binary)
			if !res.Equal(tt.hash) {
				t.Errorf("got %v, want %v", res, tt.hash)
			}
		})
	}
}

func ExampleWHash_Calculate() {
	img, err := imghash.OpenImage("assets/cat.jpg")
	if err != nil {
		panic(err)
	}
	wh, err := imghash.NewWHash()
	if err != nil {
		panic(err)
	}
	hash, err := wh.Calculate(img)
	if err != nil {
		panic(err)
	}

	fmt.Println(hash)
	// Output: [255 255 248 224 192 128 0 244]
}

var wHashDistanceTests = []struct {
	firstImage  string
	secondImage string
	distance    similarity.Distance
	width       uint
	height      uint
}{
	{"assets/lena.jpg", "assets/cat.jpg", 32, 8, 8},
	{"assets/lena.jpg", "assets/monarch.jpg", 34, 8, 8},
	{"assets/baboon.jpg", "assets/cat.jpg", 34, 8, 8},
	{"assets/peppers.jpg", "assets/baboon.jpg", 30, 8, 8},
	{"assets/tulips.jpg", "assets/monarch.jpg", 32, 8, 8},
}

func TestWHash_Distance(t *testing.T) {
	for _, tt := range wHashDistanceTests {
		t.Run(fmt.Sprintf("%v %v", tt.firstImage, tt.secondImage), func(t *testing.T) {
			hash, err := imghash.NewWHash(imghash.WithSize(tt.width, tt.height))
			if err != nil {
				t.Fatalf("failed to create hasher: %v", err)
			}
			img1, err := imghash.OpenImage(tt.firstImage)
			if err != nil {
				t.Fatalf("failed to open %s: %v", tt.firstImage, err)
			}
			img2, err := imghash.OpenImage(tt.secondImage)
			if err != nil {
				t.Fatalf("failed to open %s: %v", tt.secondImage, err)
			}
			h1, err := hash.Calculate(img1)
			if err != nil {
				t.Fatalf("failed to calculate hash for %s: %v", tt.firstImage, err)
			}
			h2, err := hash.Calculate(img2)
			if err != nil {
				t.Fatalf("failed to calculate hash for %s: %v", tt.secondImage, err)
			}
			dist, err := hash.Compare(h1, h2)
			if err != nil {
				t.Fatalf("failed to compute distance: %v", err)
			}
			if !dist.Equal(tt.distance) {
				t.Errorf("got %v, want %v", dist, tt.distance)
			}
		})
	}
}

func TestNewWHash_invalidSize(t *testing.T) {
	_, err := imghash.NewWHash(imghash.WithSize(0, 8))
	if !errors.Is(err, imghash.ErrInvalidSize) {
		t.Errorf("got %v, want imghash.ErrInvalidSize", err)
	}
}


