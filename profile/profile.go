package profile

import (
	"fmt"
	iiifimage "github.com/thisisaaronland/go-iiif/image"
	iiiflevel "github.com/thisisaaronland/go-iiif/level"
)

type Profile struct {
	Context  string        `json:"@context"`
	Id       string        `json:"@id"`
	Type     string        `json:"@type"` // Optional or iiif:Image
	Protocol string        `json:"protocol"`
	Width    int           `json:"width"`
	Height   int           `json:"height"`
	Profile  []interface{} `json:"profile"`
	//	Sizes    []string `json:"sizes"` // Optional, existing/supported sizes.
	//	Tiles    []string `json:"tiles"` // Optional
}

func NewProfile(endpoint string, image iiifimage.Image, level iiiflevel.Level) (*Profile, error) {

	dims, err := image.Dimensions()

	if err != nil {
		return nil, err
	}

	p := Profile{
		Context:  "http://iiif.io/api/image/2/context.json",
		Id:       fmt.Sprintf("%s/%s", endpoint, image.Identifier()),
		Type:     "iiif:Image",
		Protocol: "http://iiif.io/api/image",
		Width:    dims.Width(),
		Height:   dims.Height(),
		Profile: []interface{}{
			"http://iiif.io/api/image/2/level2.json",
			level,
		},
		//Sizes: []ProfileSize{},
		//Tiles: []ProfileTile{},
	}

	return &p, nil
}
