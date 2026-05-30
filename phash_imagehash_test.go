package imghash_test

import (
	"fmt"
	"testing"

	"github.com/xyxu/imghash/v2"
	"github.com/xyxu/imghash/v2/hashtype"
	"github.com/xyxu/imghash/v2/similarity"
)

var pHashImgHashCalculateTests = []struct {
	filename string
	hash     hashtype.Binary
}{
	{"assets/lena.jpg", hashtype.Binary{153, 198, 86, 45, 117, 51, 162, 150}},
	{"assets/baboon.jpg", hashtype.Binary{223, 32, 96, 125, 31, 160, 216, 143}},
	{"assets/cat.jpg", hashtype.Binary{213, 203, 135, 188, 84, 72, 101, 170}},
	{"assets/monarch.jpg", hashtype.Binary{233, 123, 100, 252, 152, 150, 1, 99}},
	{"assets/peppers.jpg", hashtype.Binary{163, 191, 124, 16, 199, 17, 200, 217}},
	{"assets/tulips.jpg", hashtype.Binary{197, 174, 67, 186, 236, 94, 12, 164}},
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
	// Output: [213 203 135 188 84 72 101 170]
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
