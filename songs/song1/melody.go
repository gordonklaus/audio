package main

import (
	"code.google.com/p/gordon-go/audio"
	"code.google.com/p/gordon-go/audiogui"
)

var melody_pattern = audiogui.NewPattern([]*audio.Note{
	{0, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 6},
			{8, 6},
		},
		"Amplitude": {
			{0, 0},
		},
	}},
	{0, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 7},
		},
		"Amplitude": {
			{0, -8},
			{8, 0},
		},
	}},
	{4, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 6.584962500721156},
			{4, 6.584962500721156},
		},
		"Amplitude": {
			{0, -8},
			{4, 0},
		},
	}},
}, map[string][]*audio.ControlPoint{
	"Distortion": {
		{0, 4},
		{1, 0},
		{8, 4},
	},
	"Amplitude": {
		{0, 0},
		{8, -1},
	},
})
