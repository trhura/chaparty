package app

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"appengine"
)

var THELOGOIMAGES = loadLogos(nil, "./logos", "./ribbons", "./nld")

func check(e error, context appengine.Context) {
	if e != nil {
		if context != nil {
			context.Errorf("Error: %s", e)
		} else {
			fmt.Fprintf(os.Stderr, "%s", e)
		}
	}
}

func addLogo(profilePtr *image.Image, logo string, context appengine.Context) []byte {
	profileImage := *profilePtr
	destImage := image.NewRGBA(profileImage.Bounds())
	draw.Draw(destImage, destImage.Bounds(), profileImage, image.ZP, draw.Src)

	if logoImages, ok := THELOGOIMAGES[logo]; ok {
		randi := rand.Intn(len(logoImages))
		logoImage := logoImages[randi]

		offset := image.Pt(5, 5)
		if strings.HasPrefix(logo, "NLD-") {
			offset = image.Pt(0, 0)
		}

		start := profileImage.Bounds().Size()
		start = start.Sub(offset)
		start = start.Sub(logoImage.Bounds().Size())

		bounds := image.Rectangle{start, start.Add(logoImage.Bounds().Size())}
		draw.Draw(destImage, bounds, logoImage, image.ZP, draw.Over)

	} else {
		context.Errorf("Cannot load logoimage for %s", logo)
	}

	buffer := new(bytes.Buffer)
	err := png.Encode(buffer, destImage)
	check(err, context)

	return buffer.Bytes()
}

func loadLogos(context appengine.Context, globpaths ...string) map[string][]image.Image {
	logoImagesByName := make(map[string][]image.Image)

	for _, path := range globpaths {
		logoFolders, err := filepath.Glob(path + "/*")
		check(err, context)

		for _, logoFolder := range logoFolders {
			logoFiles, err := filepath.Glob(logoFolder + "/*")
			check(err, context)

			filename := filepath.Base(logoFolder)
			logoImages := make([]image.Image, 0)

			for _, logoFile := range logoFiles {
				//fmt.Fprintf(os.Stderr, "%s\n", logoFile)
				logoData, err := os.Open(logoFile)
				defer logoData.Close()
				check(err, context)

				reader := bufio.NewReader(logoData)
				logoImage, err := png.Decode(reader)
				check(err, context)

				logoImages = append(logoImages, logoImage)
			}

			logoImagesByName[filename] = logoImages

		}
	}

	return logoImagesByName
}
