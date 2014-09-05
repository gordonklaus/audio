package main

import (
	"code.google.com/p/gordon-go/audio"
	"code.google.com/p/gordon-go/audiogui"
)

var reverbControl_pattern = audiogui.NewPattern([]*audio.Note{
}, map[string][]*audio.ControlPoint{
	"Dry": {
		{0, -12},
		{14, -12},
	},
	"Wet": {
		{0, 0},
		{173, 0},
		{184, -16},
	},
})
