package main

import (
	"code.google.com/p/gordon-go/audio"
	"code.google.com/p/gordon-go/audiogui"
)

var distortion_pattern = audiogui.NewPattern([]*audio.Note{
	{0, map[string][]*audio.ControlPoint{
		"Gain": {
			{0, 6.125000000000001},
			{2, 1.3819444444444444},
			{5, 4.208333333333334},
			{8, 1.0156249999999962},
		},
	}},
})
