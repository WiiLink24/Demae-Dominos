package dominos

import (
	"bufio"
	"bytes"
	"fmt"
	"golang.org/x/image/draw"
	"image"
	"image/jpeg"
	"io/ioutil"
	"net/http"
	"os"
)

func (d *Dominos) DownloadAndReturnImage(filename string) []byte {
	// If the file exists, serve it
	file, err := os.ReadFile(fmt.Sprintf("./images/%s/%s", d.country, filename))
	if err != nil {
		if !os.IsNotExist(err) {
			return nil
		}
	} else {
		return file
	}

	respChan := make(chan *http.Response)
	go d.sendAsyncGET(fmt.Sprintf("%s/%s", d.imageURL, filename), respChan)

	imageResponse := <-respChan
	defer imageResponse.Body.Close()
	jpg, err := jpeg.Decode(imageResponse.Body)
	if err != nil {
		return nil
	}

	output := resize(jpg)
	var outputImgWriter bytes.Buffer
	err = jpeg.Encode(bufio.NewWriter(&outputImgWriter), output, nil)
	if err != nil {
		return nil
	}

	ioutil.WriteFile(fmt.Sprintf("./images/%s/%s", d.country, filename), outputImgWriter.Bytes(), 0666)
	return outputImgWriter.Bytes()
}

func resize(origImage image.Image) image.Image {
	newImage := image.NewRGBA(image.Rect(0, 0, 160, 160))
	draw.BiLinear.Scale(newImage, newImage.Bounds(), origImage, origImage.Bounds(), draw.Over, nil)
	return newImage
}
