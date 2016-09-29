package image

import (
	"bytes"
	"github.com/anthonynsimon/bild/transform"
	iiifconfig "github.com/thisisaaronland/go-iiif/config"
	iiifsource "github.com/thisisaaronland/go-iiif/source"
	"image"
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

func (dims *BILDDimensions) Height() int {

	return dims.bounds.Max.X
}

func (dims *BILDDimensions) Width() int {

	return dims.bounds.Max.Y
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

	return ""
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

		box := image.Rect(rgi.X, rgi.Y, rgi.Width, rgi.Height)
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

	_, err := t.RotationInstructions(im)

	if err != nil {
		return nil
	}

	return nil
}
