package image

// package all custom (go-iiif) quality transformations here so that they can be shared
// by the various engine-specific packages (vips, bild, etc.) – doesn't work yet
// (20161202/thisisaaronland)

import (
       "errors"
	iiifconfig "github.com/thisisaaronland/go-iiif/config"
       "strconv"
       "strings"       
)

func Foo (im Image, t *Transformation, config *iiifconfig.Config) (error) { 

	fi, err := t.FormatInstructions(im)

	if err != nil {
		return nil
	}

	if t.Quality == "dither" {

		err := DitherImage(im)

		if err != nil {
			return err
		}

	} else if strings.HasPrefix(t.Quality, "primitive:") {

		parts := strings.Split(t.Quality, ":")
		parts = strings.Split(parts[1], ",")

		mode, err := strconv.Atoi(parts[0])

		if err != nil {
			return err
		}

		iters, err := strconv.Atoi(parts[1])

		if err != nil {
			return err
		}

		max_iters := config.Primitive.MaxIterations

		if max_iters > 0 && iters > max_iters {
			return errors.New("Invalid primitive iterations")
		}

		alpha, err := strconv.Atoi(parts[2])

		if err != nil {
			return err
		}

		if alpha > 255 {
			return errors.New("Invalid primitive alpha")
		}

		animated := false

		if fi.Format == "gif" {
			animated = true
		}

		opts := PrimitiveOptions{
			Alpha:      alpha,
			Mode:       mode,
			Iterations: iters,
			Size:       0,
			Animated:   animated,
		}

		err = PrimitiveImage(im, opts)

		if err != nil {
			return err
		}

		/*
		if fi.Format == "gif" {
			im.isgif = true
		}
		*/
	}

	return nil
}