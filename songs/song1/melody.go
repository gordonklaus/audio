package main

import (
	"code.google.com/p/gordon-go/audio"
	"code.google.com/p/gordon-go/audiogui"
)

var melody_pattern = audiogui.NewPattern([]*audio.Note{
	{0, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 6},
		},
		"Amplitude": {
			{0, -8},
			{1, 0},
			{3, -8},
		},
	}},
	{0, map[string][]*audio.ControlPoint{
		"Amplitude": {
			{0, -8},
			{2, 0},
			{3, -8},
		},
		"Pitch": {
			{0, 7},
		},
	}},
	{21, map[string][]*audio.ControlPoint{
		"Amplitude": {
			{0, 0},
		},
		"Pitch": {
			{0, 8},
		},
	}},
}, map[string][]*audio.ControlPoint{
	"Distortion": {
		{0, 3},
		{12, 3},
	},
	"Amplitude": {
		{0, 0},
		{8, 0},
	},
})
