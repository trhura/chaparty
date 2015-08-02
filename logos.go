package app

import (
	"bufio"
	"bytes"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"

	"appengine"
)

var logoImages map[string]image.Image = nil // = loadLogos()

func check(e error, context appengine.Context) {
	if e != nil {
		context.Errorf("Error: %s", e)
	}
}

func addLogo(profilePtr *image.Image, logo string, context appengine.Context) []byte {
	profileImage := *profilePtr
	destImage := image.NewRGBA(profileImage.Bounds())
	draw.Draw(destImage, destImage.Bounds(), profileImage, image.ZP, draw.Src)

	if logoImages == nil {
		logoImages = loadLogos("./logos/*", context)
		context.Infof("%s", logoImages)
	}

	if logoImage, ok := logoImages[logo]; ok {
		start := profileImage.Bounds().Size()
		start = start.Sub(image.Pt(5, 7))
		start = start.Sub(logoImage.Bounds().Size())

		bounds := image.Rectangle{start, start.Add(logoImage.Bounds().Size())}
		draw.Draw(destImage, bounds, logoImage, image.ZP, draw.Over)

	} else {
		context.Errorf("Cannot load logoimage for %s", logo)
	}

	buffer := new(bytes.Buffer)
	err := jpeg.Encode(buffer, destImage, nil)
	check(err, context)

	return buffer.Bytes()
}

func loadLogos(globpath string, context appengine.Context) map[string]image.Image {
	logoFiles, err := filepath.Glob(globpath)
	check(err, context)

	logoImages := make(map[string]image.Image)
	for _, logoFile := range logoFiles {
		logoData, err := os.Open(logoFile)
		defer logoData.Close()
		check(err, context)

		reader := bufio.NewReader(logoData)
		logoImage, err := png.Decode(reader)
		check(err, context)

		filename := filepath.Base(logoFile)
		logoImages[filename] = logoImage
	}

	return logoImages
}
