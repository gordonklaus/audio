package main

import (
	"code.google.com/p/gordon-go/audio"
	"code.google.com/p/gordon-go/audiogui"
)

var reverbControl_pattern = audiogui.NewPattern([]*audio.Note{}, map[string][]*audio.ControlPoint{
	"Wet": {
		{0, 0},
		{4, 0},
		{11, -12},
	},
	"Dry": {
		{0, -12},
	},
})
