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
)

var flagImages = loadFlags("./flags/*")

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func OverlayLogo(profile image.Image, flag string) []byte {
	if flagImage, ok := flagImages[flag]; ok {
		destImage := image.NewRGBA(profile.Bounds())
		draw.Draw(destImage, destImage.Bounds(), profile, image.ZP, draw.Src)

		offset := image.Pt(300, 300)
		bounds := image.Rectangle{offset, offset.Add(flagImage.Bounds().Size())}
		draw.Draw(destImage, bounds, flagImage, image.ZP, draw.Over)

		buffer := new(bytes.Buffer)
		err := jpeg.Encode(buffer, destImage, nil)
		check(err)

		return buffer.Bytes()
	}

	return nil
}

func loadFlags(globpath string) map[string]image.Image {
	flagFiles, err := filepath.Glob(globpath)
	check(err)

	flagImages := make(map[string]image.Image)
	for _, flagFile := range flagFiles {
		flagData, err := os.Open(flagFile)
		defer flagData.Close()
		check(err)

		reader := bufio.NewReader(flagData)
		flagImage, err := png.Decode(reader)
		check(err)

		filename := filepath.Base(flagFile)
		flagImages[filename] = flagImage
	}

	return flagImages
}
