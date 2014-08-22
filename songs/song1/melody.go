package main

import (
	"code.google.com/p/gordon-go/audio"
	"code.google.com/p/gordon-go/audiogui"
)

var melody = audiogui.RegisterPattern(&audio.Pattern{"melody", []*audio.Note{
	{0, map[string][]*audio.ControlPoint{
		"Amplitude": {
			{0, -10},
			{1, 0},
			{9, -12},
		},
		"Pitch": {
			{0, 7.2630344058337934},
			{2, 8},
			{9, 8},
		},
	}},
	{2, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8.584962500721156},
			{7, 8.584962500721156},
		},
		"Amplitude": {
			{0, 0},
			{7, -12},
		},
	}},
	{2, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 9.169925001442312},
			{7, 9.169925001442312},
		},
		"Amplitude": {
			{0, -10},
			{2, -2},
			{7, -12},
		},
	}},
	{3, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 9.584962500721156},
			{6, 9.584962500721156},
		},
		"Amplitude": {
			{0, 0},
			{1, -10},
			{6, -12},
		},
	}},
	{4, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 10.169925001442312},
			{5, 10.169925001442312},
		},
		"Amplitude": {
			{0, -10},
			{1, -4},
			{5, -12},
		},
	}},
}})
