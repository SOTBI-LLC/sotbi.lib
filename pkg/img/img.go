package img

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io"
	"math"
)

func New(reader io.Reader) (Resizer, error) {
	i, _, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}

	return &img{Image: i}, nil
}

type Resizer interface {
	Resize(length, width int) Resizer
	ToBytes() ([]byte, error)
	ToBase64() (string, error)
}

type img struct {
	image.Image
}

func (im *img) ToBytes() ([]byte, error) {
	opt := jpeg.Options{Quality: 80}

	buff := bytes.NewBuffer(nil)

	err := jpeg.Encode(buff, im, &opt)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func (im *img) ToBase64() (string, error) {
	bytes, err := im.ToBytes()
	if err != nil {
		return "", err
	}

	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(bytes), nil
}

func (im *img) Resize(length, width int) Resizer {
	if width <= 0 || length <= 0 || im.Bounds().Empty() {
		im.Image = image.NewRGBA(image.Rect(0, 0, 0, 0))

		return im
	}
	// truncate pixel size
	minX := im.Bounds().Min.X
	minY := im.Bounds().Min.Y
	maxX := im.Bounds().Max.X
	maxY := im.Bounds().Max.Y

	for (maxX-minX)%length != 0 {
		maxX--
	}

	for (maxY-minY)%width != 0 {
		maxY--
	}

	scaleX := (maxX - minX) / length
	scaleY := (maxY - minY) / width

	rect := image.Rect(0, 0, length, width)
	resImg := image.NewRGBA(rect)
	draw.Draw(resImg, resImg.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)

	for y := 0; y < width; y++ {
		for x := 0; x < length; x++ {
			averageColor := im.getAverageColor(
				minX+x*scaleX,
				minX+(x+1)*scaleX,
				minY+y*scaleY,
				minY+(y+1)*scaleY,
			)
			resImg.Set(x, y, averageColor)
		}
	}

	im.Image = resImg

	return im
}

func (im *img) getAverageColor(minX, maxX, minY, maxY int) color.Color {
	var red float64

	var green float64

	var blue float64

	var alpha float64

	scale := 1.0 / float64((maxX-minX)*(maxY-minY))

	for i := minX; i < maxX; i++ {
		for k := minY; k < maxY; k++ {
			r, g, b, a := im.Image.At(i, k).RGBA()
			red += float64(r) * scale
			green += float64(g) * scale
			blue += float64(b) * scale
			alpha += float64(a) * scale
		}
	}

	return color.RGBA{
		R: uint8(math.Sqrt(red)),
		G: uint8(math.Sqrt(green)),
		B: uint8(math.Sqrt(blue)),
		A: uint8(math.Sqrt(alpha)),
	}
}

// optional written to file
// err = ioutil.WriteFile("resources/test.jpg", imgBytes, 0777)
// if err != nil {
// log.Fatal(err)
// }
