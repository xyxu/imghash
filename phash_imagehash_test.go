package imghash_test

import (
	"fmt"
	"testing"

	"github.com/ajdnik/imghash/v2"
	"github.com/ajdnik/imghash/v2/hashtype"
	"github.com/ajdnik/imghash/v2/similarity"
)

var pHashImgHashCalculateTests = []struct {
	filename string
	hash     hashtype.Binary
}{
	{"assets/lena.jpg", hashtype.Binary{153, 99, 106, 180, 174, 204, 69, 105}},
	{"assets/baboon.jpg", hashtype.Binary{251, 4, 6, 190, 248, 5, 27, 241}},
	{"assets/cat.jpg", hashtype.Binary{171, 211, 225, 61, 42, 18, 166, 85}},
	{"assets/monarch.jpg", hashtype.Binary{151, 222, 38, 63, 25, 105, 128, 198}},
	{"assets/peppers.jpg", hashtype.Binary{197, 253, 62, 8, 227, 136, 19, 155}},
	{"assets/tulips.jpg", hashtype.Binary{163, 117, 194, 93, 55, 122, 48, 37}},
}

func TestPHashImageHash_Calculate(t *testing.T) {
	for _, tt := range pHashImgHashCalculateTests {
		t.Run(tt.filename, func(t *testing.T) {
			hash := imghash.NewPHashImageHash()
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

func ExamplePHashImageHash_Calculate() {
	img, err := imghash.OpenImage("assets/cat.jpg")
	if err != nil {
		panic(err)
	}
	ph := imghash.NewPHashImageHash()
	hash, err := ph.Calculate(img)
	if err != nil {
		panic(err)
	}

	fmt.Println(hash)
	// Output: [171 211 225 61 42 18 166 85]
}

var pHashImgHashDistanceTests = []struct {
	firstImage  string
	secondImage string
	distance    similarity.Distance
}{
	{"assets/lena.jpg", "assets/cat.jpg", 30},
	{"assets/lena.jpg", "assets/monarch.jpg", 36},
	{"assets/baboon.jpg", "assets/cat.jpg", 34},
	{"assets/peppers.jpg", "assets/baboon.jpg", 32},
	{"assets/tulips.jpg", "assets/monarch.jpg", 30},
}

func TestPHashImageHash_Distance(t *testing.T) {
	for _, tt := range pHashImgHashDistanceTests {
		t.Run(fmt.Sprintf("%v %v", tt.firstImage, tt.secondImage), func(t *testing.T) {
			hash := imghash.NewPHashImageHash()
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
