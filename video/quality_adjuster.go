package video

import (
	"bytes"
	"image/jpeg"
	"log"
)

type QualityAdjuster struct {
	CompressionQuality int
}

func NewQualityAdjuster() *QualityAdjuster {
	return &QualityAdjuster{
		CompressionQuality: jpeg.DefaultQuality,
	}
}

func (it *QualityAdjuster) Compress(image []byte) []byte {
	decodedImage, err := jpeg.Decode(bytes.NewBuffer(image))
	if err != nil {
		log.Fatalln("[ERRPR] cannot decode frame from jpeg:", err)
	}

	buffer := make([]byte, 0)
	encodedImage := bytes.NewBuffer(buffer)

	err = jpeg.Encode(encodedImage, decodedImage, &jpeg.Options{Quality: it.CompressionQuality})
	if err != nil {
		log.Fatalln("[ERRPR] cannot encode frame to jpeg:", err)
	}
	return encodedImage.Bytes()
}

func (it *QualityAdjuster) ChangeCompressionQuality(newCompressionQuality int) {
	it.CompressionQuality = newCompressionQuality
}
