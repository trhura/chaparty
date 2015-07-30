package app

import (
	"bufio"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
)

var flagImages = loadFlags("./flags/*")

// func main() {
// 	proFile, err := os.Open("./p.jpg")
// 	check(err)

// 	reader := bufio.NewReader(proFile)
// 	profileImage, err := jpeg.Decode(reader)
// 	check(err)

// 	OverLayFlag(profileImage, "NLD")

// }

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func OverLayFlag(profile image.Image, flag string) {
	if flagImage, ok := flagImages[flag]; ok {
		fmt.Println("yes")

		destImage := image.NewRGBA(profile.Bounds())
		draw.Draw(destImage, destImage.Bounds(), profile, image.ZP, draw.Src)

		offset := image.Pt(300, 300)
		bounds := image.Rectangle{offset, offset.Add(flagImage.Bounds().Size())}
		draw.Draw(destImage, bounds, flagImage, image.ZP, draw.Over)

		outputImage, _ := os.Create("out.png")
		defer outputImage.Close()
		png.Encode(outputImage, destImage)
	}
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
		//fmt.Println(filename)
	}

	return flagImages
}
