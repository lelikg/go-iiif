package image

// https://github.com/anthonynsimon/bild/

import (
	"bytes"
	"fmt"
	"github.com/anthonynsimon/bild/effect"
	"github.com/anthonynsimon/bild/segment"
	"github.com/anthonynsimon/bild/transform"
	iiifconfig "github.com/thisisaaronland/go-iiif/config"
	iiifsource "github.com/thisisaaronland/go-iiif/source"
	"image"
	_ "image/gif"
	_ "log"
)

type BILDImage struct {
	Image
	config    *iiifconfig.Config
	source    iiifsource.Source
	format    string
	source_id string
	id        string
	image     image.Image
}

type BILDDimensions struct {
	Dimensions
	bounds image.Rectangle
}

func (dims *BILDDimensions) String() string {

	return fmt.Sprintf("%d x %d", dims.Width(), dims.Height())
}

func (dims *BILDDimensions) Height() int {

	return dims.bounds.Max.Y
}

func (dims *BILDDimensions) Width() int {

	return dims.bounds.Max.X
}

func NewBILDImageFromConfigWithSource(config *iiifconfig.Config, src iiifsource.Source, id string) (*BILDImage, error) {

	body, err := src.Read(id)

	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(body)
	img, format, err := image.Decode(buf)

	if err != nil {
		return nil, err
	}

	bim := BILDImage{
		id:        id,
		source_id: id,
		image:     img,
		format:    format,
		config:    config,
	}

	return &bim, nil
}

func (im *BILDImage) Update(body []byte) error {

	buf := bytes.NewBuffer(body)

	img, format, err := image.Decode(buf)

	/*

		TO DO: ALLOW body TO BE DECODED AS AN ANIMATED GIF
		github.com/thisisaaronland/go-iiif/image

		src/github.com/thisisaaronland/go-iiif/image/bild.go:87: cannot use img (type *gif.GIF) as type image.Image in assignment:
								 *gif.GIF does not implement image.Image (missing At method)
		src/github.com/thisisaaronland/go-iiif/image/bild.go:88: undefined: format

		gifs, err := gif.DecodeAll(buf)

		img := gifs.Image	// returns an image.Paletted
		format := "gif"

	*/

	if err != nil {
		return err
	}

	im.image = img
	im.format = format
	return nil
}

func (im *BILDImage) Body() []byte {

	body, _ := GolangImageToBytes(im.image, im.ContentType())
	return body
}

func (im *BILDImage) Format() string {

	return im.format
}

func (im *BILDImage) ContentType() string {

	format := im.Format()
	content_type, _ := ImageFormatToContentType(format)

	return content_type
}

func (im *BILDImage) Identifier() string {

	return im.id
}

func (im *BILDImage) Rename(id string) error {

	im.id = id
	return nil
}

func (im *BILDImage) Dimensions() (Dimensions, error) {

	bounds := im.image.Bounds()

	dims := BILDDimensions{
		bounds: bounds,
	}

	return &dims, nil
}

func (im *BILDImage) Transform(t *Transformation) error {

	// http://iiif.io/api/image/2.1/#order-of-implementation

	// seems to work (20161006/thisisaaronland)

	if t.Region != "full" {

		rgi, err := t.RegionInstructions(im)

		if err != nil {
			return err
		}

		box := image.Rect(rgi.X, rgi.Y, rgi.X+rgi.Width, rgi.Y+rgi.Height)
		crop := transform.Crop(im.image, box)

		im.image = crop
	}

	// seems to work (20161002/thisisaaronland)

	if t.Size != "max" && t.Size != "full" {

		si, err := t.SizeInstructions(im)

		if err != nil {
			return err
		}

		resized := transform.Resize(im.image, si.Width, si.Height, transform.Linear)
		im.image = resized
	}

	// seems to work (20161004/thisisaaronland)

	ri, err := t.RotationInstructions(im)

	if err != nil {
		return err
	}

	if ri.Flip {
		flipped := transform.FlipH(im.image)
		im.image = flipped
	}

	if ri.Angle != 0 {

		opts := &transform.RotationOptions{ResizeBounds: true, Pivot: nil}
		angle := float64(ri.Angle)

		rotated := transform.Rotate(im.image, angle, opts)
		im.image = rotated
	}

	// seems to work - not sure how best to define the threshold value
	// for bitonal images... (20161004/thisisaaronland)

	if t.Quality == "color" || t.Quality == "default" {

		// do nothing.

	} else if t.Quality == "gray" {

		grey := effect.Grayscale(im.image)
		im.image = grey

	} else if t.Quality == "bitonal" {

		bw := segment.Threshold(im.image, 160)
		im.image = bw

	} else {
		// this should be trapped above
	}

	// this seems to work (20161005/thisisaaronland)

	fi, err := t.FormatInstructions(im)

	if err != nil {
		return err
	}

	content_type, err := ImageFormatToContentType(fi.Format)

	if err != nil {
		return err
	}

	if content_type != im.ContentType() {

		converted, format, err := GolangImageToGolangImage(im.image, content_type)

		if err != nil {
			return err
		}

		im.image = converted
		im.format = format
	}

	// rsp, err := CustomTransform(im, t, im.config)
	_, err = CustomTransform(im, t, im.config)

	if err != nil {
		return err
	}

	return nil
}
