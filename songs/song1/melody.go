package main

import "code.google.com/p/gordon-go/audio"

var melody = []*audio.Note{
	{0, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 1},
			{1, 0},
		},
		"Amplitude": {
			{0, 0},
			{0.25, 1},
			{1, 0},
		},
	}},
	{1, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 0},
		},
		"Amplitude": {
			{0, 0},
			{1, 0.6000000000000001},
		},
	}},
	{2, map[string][]*audio.ControlPoint{
		"Amplitude": {
			{0, 0},
			{0.25, 1},
			{1, 0},
		},
		"Pitch": {
			{0, 0},
			{1, 1},
		},
	}},
	{3, map[string][]*audio.ControlPoint{
		"Amplitude": {
			{0, 1},
			{1, -0.8},
		},
		"Pitch": {
			{0, 0.2},
		},
	}},
}
