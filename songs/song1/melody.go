package main

import "code.google.com/p/gordon-go/audio"

var melody = &audio.Pattern{"melody", []*audio.Note{
	{0, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 0},
			{1, 0},
			{2, 1},
			{3, 1},
			{4, 0},
			{5, 0},
		},
		"Amplitude": {
			{0, 1},
		},
	}},
	{0, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 0},
			{5, 0},
		},
		"Amplitude": {
			{0, 1},
		},
	}},
}}
