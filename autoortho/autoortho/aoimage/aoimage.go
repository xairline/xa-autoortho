package main

/*
#cgo CFLAGS: -I/opt/homebrew/Cellar/jpeg-turbo/3.0.4/include
#cgo LDFLAGS: -L/opt/homebrew/Cellar/jpeg-turbo/3.0.4/lib -lturbojpeg

#include <stdlib.h>
#include <stdint.h>
#include <turbojpeg.h>
//#include "aoimage.h"

// Include the C source code directly
// Be cautious with including C source files directly as it may cause multiple definitions.
// For small projects or testing purposes, it can be acceptable.
#include "aoimage.c"
*/
import "C"

import (
	"fmt"
	"unsafe"
)

// AoImage struct wraps the C aoimage_t struct
type AoImage struct {
	cImage *C.aoimage_t
}

// NewAoImage initializes a new AoImage instance
func NewAoImage() *AoImage {
	// Allocate memory for C aoimage_t
	cImg := (*C.aoimage_t)(C.malloc(C.size_t(unsafe.Sizeof(C.aoimage_t{}))))
	C.memset(unsafe.Pointer(cImg), 0, C.size_t(unsafe.Sizeof(C.aoimage_t{})))
	return &AoImage{cImage: cImg}
}

// Close frees the resources associated with the AoImage
func (img *AoImage) Close() {
	if img.cImage != nil {
		C.aoimage_delete(img.cImage)
		C.free(unsafe.Pointer(img.cImage))
		img.cImage = nil
	}
}

// Convert converts the image to the specified mode (only RGBA supported)
func (img *AoImage) Convert(mode string) (*AoImage, error) {
	if mode != "RGBA" {
		return nil, fmt.Errorf("only conversion to RGBA supported")
	}
	newImg := NewAoImage()
	res := C.aoimage_2_rgba(img.cImage, newImg.cImage)
	if res == 0 {
		errMsg := C.GoString(&newImg.cImage.errmsg[0])
		newImg.Close()
		return nil, fmt.Errorf("AoImage.Convert error: %s", errMsg)
	}
	return newImg, nil
}

// Reduce2 reduces the image by a factor of 2, 'steps' number of times
func (img *AoImage) Reduce2(steps int) (*AoImage, error) {
	if steps < 1 {
		return nil, fmt.Errorf("useless Reduce2")
	}
	half := img
	for steps >= 1 {
		orig := half
		halfImg := NewAoImage()
		res := C.aoimage_reduce_2(orig.cImage, halfImg.cImage)
		if res == 0 {
			errMsg := C.GoString(&halfImg.cImage.errmsg[0])
			halfImg.Close()
			return nil, fmt.Errorf("AoImage.Reduce2 error: %s", errMsg)
		}
		half = halfImg
		steps--
	}
	return half, nil
}

// Scale scales the image by the specified factor
func (img *AoImage) Scale(factor uint32) (*AoImage, error) {
	scaledImg := NewAoImage()
	res := C.aoimage_scale(img.cImage, scaledImg.cImage, C.uint32_t(factor))
	if res == 0 {
		errMsg := C.GoString(&scaledImg.cImage.errmsg[0])
		scaledImg.Close()
		return nil, fmt.Errorf("AoImage.Scale error: %s", errMsg)
	}
	return scaledImg, nil
}

// WriteJPG writes the image to a JPEG file with the specified quality
func (img *AoImage) WriteJPG(filename string, quality int32) error {
	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))
	res := C.aoimage_write_jpg(cFilename, img.cImage, C.int32_t(quality))
	if res == 0 {
		errMsg := C.GoString(&img.cImage.errmsg[0])
		return fmt.Errorf("AoImage.WriteJPG error: %s", errMsg)
	}
	return nil
}

// ToBytes converts the image to a byte slice
func (img *AoImage) ToBytes() ([]byte, error) {
	size := int(img.cImage.width * img.cImage.height * img.cImage.channels)
	buf := C.malloc(C.size_t(size))
	defer C.free(buf)
	C.aoimage_tobytes(img.cImage, (*C.uint8_t)(buf))
	data := C.GoBytes(buf, C.int(size))
	return data, nil
}

// DataPtr returns the pointer to the image data
func (img *AoImage) DataPtr() uintptr {
	return uintptr(unsafe.Pointer(img.cImage.ptr))
}

// Paste pastes another image onto this image at the specified position
func (img *AoImage) Paste(pImg *AoImage, x, y uint32) error {
	res := C.aoimage_paste(img.cImage, pImg.cImage, C.uint32_t(x), C.uint32_t(y))
	if res == 0 {
		errMsg := C.GoString(&img.cImage.errmsg[0])
		return fmt.Errorf("AoImage.Paste error: %s", errMsg)
	}
	return nil
}

// Copy creates a copy of the image
func (img *AoImage) Copy(heightOnly uint32) (*AoImage, error) {
	newImg := NewAoImage()
	res := C.aoimage_copy(img.cImage, newImg.cImage, C.uint32_t(heightOnly))
	if res == 0 {
		errMsg := C.GoString(&newImg.cImage.errmsg[0])
		newImg.Close()
		return nil, fmt.Errorf("AoImage.Copy error: %s", errMsg)
	}
	return newImg, nil
}

// Crop crops the image at the specified position
func (img *AoImage) Crop(x, y uint32) (*AoImage, error) {
	cImg := NewAoImage()
	res := C.aoimage_crop(img.cImage, cImg.cImage, C.uint32_t(x), C.uint32_t(y))
	if res == 0 {
		errMsg := C.GoString(&cImg.cImage.errmsg[0])
		cImg.Close()
		return nil, fmt.Errorf("AoImage.Crop error: %s", errMsg)
	}
	return cImg, nil
}

// Desaturate desaturates the image by the specified saturation factor
func (img *AoImage) Desaturate(saturation float32) error {
	if saturation < 0.0 || saturation > 1.0 {
		return fmt.Errorf("invalid saturation value")
	}
	res := C.aoimage_desaturate(img.cImage, C.float(saturation))
	if res == 0 {
		errMsg := C.GoString(&img.cImage.errmsg[0])
		return fmt.Errorf("AoImage.Desaturate error: %s", errMsg)
	}
	return nil
}

// Size returns the width and height of the image
func (img *AoImage) Size() (uint32, uint32) {
	return uint32(img.cImage.width), uint32(img.cImage.height)
}

// Factory functions

// New creates a new image with the specified mode, dimensions, and color
func New(mode string, width, height uint32, color [3]uint32) (*AoImage, error) {
	if mode != "RGBA" {
		return nil, fmt.Errorf("only 'RGBA' mode is supported")
	}
	img := NewAoImage()
	res := C.aoimage_create(img.cImage, C.uint32_t(width), C.uint32_t(height), C.uint32_t(color[0]), C.uint32_t(color[1]), C.uint32_t(color[2]))
	if res == 0 {
		errMsg := C.GoString(&img.cImage.errmsg[0])
		img.Close()
		return nil, fmt.Errorf("AoImage.New error: %s", errMsg)
	}
	return img, nil
}

// LoadFromMemory loads an image from a byte slice
func LoadFromMemory(mem []byte) (*AoImage, error) {
	img := NewAoImage()
	res := C.aoimage_from_memory(img.cImage, (*C.uint8_t)(unsafe.Pointer(&mem[0])), C.uint32_t(len(mem)))
	if res == 0 {
		errMsg := C.GoString(&img.cImage.errmsg[0])
		img.Close()
		return nil, fmt.Errorf("AoImage.LoadFromMemory error: %s", errMsg)
	}
	return img, nil
}

// Open opens an image from a file
func Open(filename string) (*AoImage, error) {
	img := NewAoImage()
	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))
	res := C.aoimage_read_jpg(cFilename, img.cImage)
	if res == 0 {
		errMsg := C.GoString(&img.cImage.errmsg[0])
		img.Close()
		return nil, fmt.Errorf("AoImage.Open error: %s", errMsg)
	}
	return img, nil
}

func main() {
	// Example usage
	width := uint32(16)
	height := uint32(16)

	black, err := New("RGBA", 256*width, 256*height, [3]uint32{0, 0, 0})
	if err != nil {
		fmt.Printf("Error creating black image: %v\n", err)
		return
	}
	w, h := black.Size()
	fmt.Printf("Black image size: %d x %d\n", w, h)
	if err := black.WriteJPG("black.jpg", 90); err != nil {
		fmt.Printf("Error writing black image: %v\n", err)
	}
	black.Close()
	fmt.Println("Black image done")

	// Continue with other operations as needed
}
