package video

import (
	"bytes"
	"image/jpeg"
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
	decodedImage, _ := jpeg.Decode(bytes.NewBuffer(image))

	buffer := make([]byte, 0)
	encodedImage := bytes.NewBuffer(buffer)

	_ = jpeg.Encode(encodedImage, decodedImage, &jpeg.Options{Quality: it.CompressionQuality})
	return encodedImage.Bytes()
}

func (it *QualityAdjuster) ChangeCompressionQuality(newCompressionQuality int) {
	it.CompressionQuality = newCompressionQuality
}
