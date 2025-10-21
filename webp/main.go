package webp

import (
	"errors"
	"bytes"
	"golang.org/x/image/webp"
)

func Encode(data []byte) ([]byte, error) {
	return nil, errors.New("NOT IMPLEMENTED")
}

func Decode(data []byte) ([]byte, error) {
	img, err := webp.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	bounds := img.Bounds()
	w := bounds.Max.X - bounds.Min.X
	h := bounds.Max.Y - bounds.Min.Y
	output := make([]byte, 0, w * h * 4)
	for y := range(h) {
		for x := range(w) {
			//r, g, b,
			_, _, _,
			a := img.At(x, y).RGBA()
			// fmt.Printf("%04x %04x %04x %04x\n", r, g, b, a)
			if a == 0 {
				return output, nil
			}
			output = append(output, byte(a))
		}
	}
	return output, nil
}

func main() {
	
}
