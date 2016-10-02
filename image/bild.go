package image

// https://github.com/anthonynsimon/bild/

import (
	"bytes"
	"fmt"
	"github.com/anthonynsimon/bild/effect"
	"github.com/anthonynsimon/bild/transform"
	iiifconfig "github.com/thisisaaronland/go-iiif/config"
	iiifsource "github.com/thisisaaronland/go-iiif/source"
	"image"
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
	}

	return &bim, nil
}

func (im *BILDImage) Update(body []byte) error {

	buf := bytes.NewBuffer(body)
	img, format, err := image.Decode(buf)

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

	if t.Region != "full" {

		rgi, err := t.RegionInstructions(im)

		if err != nil {
			return err
		}

		box := image.Rect(rgi.X, rgi.Y, rgi.X+rgi.Width, rgi.Y+rgi.Height)
		crop := transform.Crop(im.image, box)

		im.image = crop
	}

	if t.Size != "max" && t.Size != "full" {

		si, err := t.SizeInstructions(im)

		if err != nil {
			return err
		}

		resized := transform.Resize(im.image, si.Width, si.Height, transform.Linear)
		im.image = resized
	}

	ri, err := t.RotationInstructions(im)

	if err != nil {
		return nil
	}

	if ri.Angle != 0 {

		opts := &transform.RotationOptions{ResizeBounds: false, Pivot: nil}
		angle := float64(ri.Angle)

		rotated := transform.Rotate(im.image, angle, opts)
		im.image = rotated
	}

	if t.Quality == "color" || t.Quality == "default" {
		// do nothing.
	} else if t.Quality == "gray" {

		grey := effect.Grayscale(im.image)
		im.image = grey

	} else if t.Quality == "bitonal" {

		// how to do this in bild? (20160930/thisisaaronland)

	} else {
		// this should be trapped above
	}

	fi, err := t.FormatInstructions(im)

	if err != nil {
		return nil
	}

	ct, _ := ImageFormatToContentType(fi.Format)

	if ct != im.ContentType() {

		// this doesn't work yet (20161001/thisisaaronland)

		source_content_type := im.ContentType()
		dest_content_type, _ := ImageFormatToContentType(fi.Format)

		converted, err := GolangImageToGolangImage(im.image, source_content_type, dest_content_type)

		if err != nil {
			return nil
		}

		im.image = converted
	}

	return nil
}
