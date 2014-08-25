package main

import (
	"code.google.com/p/gordon-go/audio"
	"code.google.com/p/gordon-go/audiogui"
)

var distortion_pattern = audiogui.NewPattern([]*audio.Note{
	{0, map[string][]*audio.ControlPoint{
		"Gain": {
			{0, 8},
			{3, 1},
			{5, 8},
			{8, 1},
		},
	}},
})
