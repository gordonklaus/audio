package main

import (
	"code.google.com/p/gordon-go/audio"
	"code.google.com/p/gordon-go/audiogui"
)

var sines_pattern = audiogui.NewPattern([]*audio.Note{
	{0, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8},
			{8, 8},
		},
		"Amplitude": {
			{0, -3},
		},
	}},
	{1, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 9},
			{6, 9},
		},
		"Amplitude": {
			{0, -3},
		},
	}},
	{2, map[string][]*audio.ControlPoint{
		"Amplitude": {
			{0, -3},
		},
		"Pitch": {
			{0, 8.584962500721156},
			{4, 8.584962500721156},
		},
	}},
	{3, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8.415037499278844},
			{4, 8.415037499278844},
		},
		"Amplitude": {
			{0, -3},
		},
	}},
	{4, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8.736965594166206},
			{5, 8.736965594166206},
		},
		"Amplitude": {
			{0, -3},
		},
	}},
	{5, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8.321928094887362},
			{3, 8.321928094887362},
		},
		"Amplitude": {
			{0, -3},
		},
	}},
}, map[string][]*audio.ControlPoint{
	"Distortion": {
		{0, 0},
	},
	"Amplitude": {
		{0, 0},
	},
})
