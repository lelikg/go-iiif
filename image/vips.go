package image

// https://github.com/h2non/bimg
// https://github.com/jcupitt/libvips

import (
	"bytes"
	"errors"
	"fmt"
	iiifconfig "github.com/thisisaaronland/go-iiif/config"
	iiifsource "github.com/thisisaaronland/go-iiif/source"
	"gopkg.in/h2non/bimg.v1"
	"image"
	"image/gif"
	_ "log"
)

type VIPSImage struct {
	Image
	config    *iiifconfig.Config
	source    iiifsource.Source
	source_id string
	id        string
	bimg      *bimg.Image
	isgif     bool
}

type VIPSDimensions struct {
	Dimensions
	imagesize bimg.ImageSize
}

func (d *VIPSDimensions) Height() int {
	return d.imagesize.Height
}

func (d *VIPSDimensions) Width() int {
	return d.imagesize.Width
}

/*

See notes in NewVIPSImageFromConfigWithSource - basically getting an image's
dimensions after the we've done the GIF conversion (just see the notes...)
will make bimg/libvips sad so account for that in Dimensions() and create a
pure Go implementation of the Dimensions interface (20160922/thisisaaronland)

*/

type GolangImageDimensions struct {
	Dimensions
	image image.Image
}

func (dims *GolangImageDimensions) Width() int {
	bounds := dims.image.Bounds()
	return bounds.Max.X
}

func (dims *GolangImageDimensions) Height() int {
	bounds := dims.image.Bounds()
	return bounds.Max.Y
}

func NewVIPSImageFromConfigWithSource(config *iiifconfig.Config, src iiifsource.Source, id string) (*VIPSImage, error) {

	body, err := src.Read(id)

	if err != nil {
		return nil, err
	}

	bimg := bimg.NewImage(body)

	im := VIPSImage{
		config:    config,
		source:    src,
		source_id: id,
		id:        id,
		bimg:      bimg,
		isgif:     false,
	}

	/*

		Hey look - see the 'isgif' flag? We're going to hijack the fact that
		bimg doesn't handle GIF files and if someone requests them then we
		will do the conversion after the final call to im.bimg.Process and
		after we do handle any custom features. We are relying on the fact
		that both bimg.NewImage and bimg.Image() expect and return raw bytes
		and we are ignoring whatever bimg thinks in the Format() function.
		So basically you should not try to any processing in bimg/libvips
		after the -> GIF transformation. (20160922/thisisaaronland)

		See also: https://github.com/h2non/bimg/issues/41
	*/

	return &im, nil
}

func (im *VIPSImage) Update(body []byte) error {

	bimg := bimg.NewImage(body)
	im.bimg = bimg

	return nil
}

func (im *VIPSImage) Body() []byte {

	return im.bimg.Image()
}

func (im *VIPSImage) Format() string {

	return im.bimg.Type()
}

func (im *VIPSImage) ContentType() string {

	format := im.Format()
	content_type, _ := ImageFormatToContentType(format)

	return content_type
}

func (im *VIPSImage) Identifier() string {
	return im.id
}

func (im *VIPSImage) Rename(id string) error {
	im.id = id
	return nil
}

func (im *VIPSImage) Dimensions() (Dimensions, error) {

	// see notes in NewVIPSImageFromConfigWithSource
	// ideally this never gets triggered but just in case...

	if im.isgif {

		buf := bytes.NewBuffer(im.Body())
		goimg, err := gif.Decode(buf)

		if err != nil {
			return nil, err
		}

		d := GolangImageDimensions{
			image: goimg,
		}

		return &d, nil
	}

	sz, err := im.bimg.Size()

	if err != nil {
		return nil, err
	}

	d := VIPSDimensions{
		imagesize: sz,
	}

	return &d, nil
}

// https://godoc.org/github.com/h2non/bimg#Options

func (im *VIPSImage) Transform(t *Transformation) error {

	// http://iiif.io/api/image/2.1/#order-of-implementation
	var opts bimg.Options

	if t.Region != "full" {

		rgi, err := t.RegionInstructions(im)

		if err != nil {
			return err
		}

		opts = bimg.Options{
			AreaWidth:  rgi.Width,
			AreaHeight: rgi.Height,
			Left:       rgi.X,
			Top:        rgi.Y,
		}

		/*

					We need to do this or libvips will freak out and think it's trying to save
			   		an SVG file which it can't do (20160929/thisisaaronland)

		*/

		if im.ContentType() == "image/svg+xml" {
			opts.Type = bimg.PNG

		}

		/*
		   So here's a thing that we need to do because... computers?
		   (20160910/thisisaaronland)
		*/

		if opts.Top == 0 && opts.Left == 0 {
			opts.Top = -1
		}

		_, err = im.bimg.Process(opts)

		if err != nil {
			return err
		}

	}

	dims, err := im.Dimensions()

	if err != nil {
		return err
	}

	opts = bimg.Options{
		Width:  dims.Width(),  // opts.AreaWidth,
		Height: dims.Height(), // opts.AreaHeight,
	}

	if t.Size != "max" && t.Size != "full" {

		si, err := t.SizeInstructions(im)

		if err != nil {
			return err
		}

		opts.Height = si.Height
		opts.Width = si.Width
		opts.Enlarge = si.Enlarge
		opts.Force = si.Force
	}

	ri, err := t.RotationInstructions(im)

	if err != nil {
		return nil
	}

	opts.Flip = ri.Flip
	opts.Rotate = bimg.Angle(int(ri.Angle) % 360)

	if t.Quality == "color" || t.Quality == "default" {
		// do nothing.
	} else if t.Quality == "gray" {
		opts.Interpretation = bimg.InterpretationBW
	} else if t.Quality == "bitonal" {
		opts.Interpretation = bimg.InterpretationBW
	} else {
		// this should be trapped above
	}

	fi, err := t.FormatInstructions(im)

	if err != nil {
		return nil
	}

	if fi.Format == "jpg" {
		opts.Type = bimg.JPEG
	} else if fi.Format == "png" {
		opts.Type = bimg.PNG
	} else if fi.Format == "webp" {
		opts.Type = bimg.WEBP
	} else if fi.Format == "tif" {
		opts.Type = bimg.TIFF
	} else if fi.Format == "gif" {
		opts.Type = bimg.PNG // see this - we're just going to trick libvips until the very last minute...
	} else {
		msg := fmt.Sprintf("Unsupported image format '%s'", fi.Format)
		return errors.New(msg)
	}

	_, err = im.bimg.Process(opts)

	if err != nil {
		return err
	}

	rsp, err := CustomTransform(im, t, im.config)

	if err != nil {
		return err
	}

	// see notes in NewVIPSImageFromConfigWithSource

	if fi.Format == "gif" && !rsp.IsGIF {

		goimg, err := IIIFImageToGolangImage(im)

		if err != nil {
			return err
		}

		err = GolangImageToIIIFImage(goimg, im)

		if err != nil {
			return err
		}

	}

	return nil
}
