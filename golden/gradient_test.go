package golden_test

import (
	"bytes"
	"flag"
	"image"
	"image/draw"
	"image/png"
	"io/ioutil"
	"os"
	"testing"

	"github.com/joncalhoun/twg/golden"
)

var updateFlag bool

func init() {
	flag.BoolVar(&updateFlag, "update", false, "set the update flag to update the expected output of any golden file tests")
}

func TestFibGradient(t *testing.T) {
	const (
		wantFile = "TestFibGradient.want.png"
		gotFile  = "TestFibGradient.got.png"
	)

	var im draw.Image = image.NewRGBA(image.Rect(0, 0, 400, 400))
	golden.FibGradient(im)

	if updateFlag {
		f, err := os.Create(wantFile)
		if err != nil {
			t.Fatalf("os.Create() err = %s; want nil", err)
		}
		png.Encode(f, im)
		f.Close()
	}

	f, err := os.Open(wantFile)
	if err != nil {
		t.Fatalf("os.Open() err = %s; want nil", err)
	}
	want, err := ioutil.ReadAll(f)
	f.Close()
	if err != nil {
		t.Fatalf("ioutil.ReadAll() err = %s; want nil", err)
	}
	got := bytes.NewBuffer(nil)
	png.Encode(got, im)
	if !bytesEq(t, got.Bytes(), want) {
		f, err := os.Create(gotFile)
		if err != nil {
			t.Fatalf("os.Create() err = %s; want nil", err)
		}
		png.Encode(f, im)
		f.Close()
		t.Errorf("image does not match golden file. See %s and %s.", gotFile, wantFile)
	}
}

func bytesEq(t *testing.T, a, b []byte) bool {
	if len(a) != len(b) {
		t.Logf("bytesEq: len(a) = %d; want %d", len(a), len(b))
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			t.Logf("bytesEq: difference exists at index %d", i)
			return false
		}
	}
	return true
}
