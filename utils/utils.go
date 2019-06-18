package utils

import (
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"strings"

	"github.com/pkg/errors"
)

func SplitEscapedString(s string) []string {
	var result []string
	parts := strings.Split(s, " ")
	previousPartEndsWithSlash := false
	for _, part := range parts {
		if previousPartEndsWithSlash {
			result[len(result)-1] = strings.TrimSuffix(result[len(result)-1], `\`) + " " + part
		} else {
			result = append(result, part)
		}
		previousPartEndsWithSlash = strings.HasSuffix(part, `\`)
	}
	return result
}

// PrettyPrint outputs v as JSON with indentation
func PrettyPrint(v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func LoadPngImage(reader io.Reader) (rgbaImage *image.RGBA, err error) {
	defer func() {
		if err != nil {
			err = errors.Wrap(err, "Unable to load image")
		}
	}()
	var img image.Image
	img, err = png.Decode(reader)
	if err != nil {
		return
	}
	rgbaImage = image.NewRGBA(img.Bounds())
	draw.Draw(rgbaImage, rgbaImage.Bounds(), img, image.Point{0, 0}, draw.Src)
	return
}

func QuoteString(s string) string {
	return `"` + s + `"`
}
