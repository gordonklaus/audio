package main

import (
	"code.google.com/p/gordon-go/audio"
	"code.google.com/p/gordon-go/audiogui"
)

var reverb_pattern = audiogui.NewPattern([]*audio.Note{
}, map[string][]*audio.ControlPoint{
	"Sustain": {
		{0, -1},
	},
	"Dry": {
		{0, -12},
	},
	"Wet": {
		{0, 0},
		{173, 0},
		{184, -16},
	},
	"Decay": {
		{0, 1},
	},
})
