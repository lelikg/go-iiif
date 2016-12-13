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

type CustomResponse struct {
     IsGIF bool
}

func CustomTransform (im Image, t *Transformation, config *iiifconfig.Config) (*CustomResponse, error) { 

	rsp := CustomResponse{
	    IsGIF: false,
	}

	fi, err := t.FormatInstructions(im)

	if err != nil {
		return nil, err
	}

	if t.Quality == "dither" {

		err := DitherImage(im)

		if err != nil {
			return nil, err
		}

	} else if strings.HasPrefix(t.Quality, "primitive:") {

		parts := strings.Split(t.Quality, ":")
		parts = strings.Split(parts[1], ",")

		mode, err := strconv.Atoi(parts[0])

		if err != nil {
			return nil, err
		}

		iters, err := strconv.Atoi(parts[1])

		if err != nil {
			return nil, err
		}

		max_iters := config.Primitive.MaxIterations

		if max_iters > 0 && iters > max_iters {
			err = errors.New("Invalid primitive iterations")
			return nil, err
		}

		alpha, err := strconv.Atoi(parts[2])

		if err != nil {
			return nil, err
		}

		if alpha > 255 {
			err = errors.New("Invalid primitive alpha")
			return nil, err
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
			return nil, err
		}

		if fi.Format == "gif" {
			rsp.IsGIF = true
		}
	}

	return &rsp, nil
}