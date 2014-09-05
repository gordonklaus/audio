package main

import (
	"code.google.com/p/gordon-go/audio"
	"code.google.com/p/gordon-go/audiogui"
)

var melody_pattern = audiogui.NewPattern([]*audio.Note{
	{0, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 7},
		},
		"Amplitude": {
			{0, -12},
			{2, 0},
			{4, -12},
		},
	}},
}, map[string][]*audio.ControlPoint{
	"Amplitude": {
		{0, 0},
	},
	"Distortion": {
		{0, 0},
	},
})
