package main

import "code.google.com/p/gordon-go/audio"

var melody = &audio.Pattern{"melody", []*audio.Note{
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
			{0, -0.4},
			{1, 0.20000000000000007},
		},
	}},
	{2, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 0},
			{1, 1},
		},
		"Amplitude": {
			{0, 0},
			{0.25, 1},
			{1, 0},
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
}}
