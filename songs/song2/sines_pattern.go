package main

import (
	"code.google.com/p/gordon-go/audio"
	"code.google.com/p/gordon-go/audiogui"
)

var sines_pattern = audiogui.NewPattern([]*audio.Note{
	{0, map[string][]*audio.ControlPoint{
		"Amplitude": {
			{0, -2},
		},
		"Pitch": {
			{0, 8},
			{1, 8},
		},
	}},
	{1, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8.584962500721156},
			{1, 8.584962500721156},
		},
		"Amplitude": {
			{0, -2},
		},
	}},
	{2, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 9},
			{1, 9},
		},
		"Amplitude": {
			{0, -2},
		},
	}},
	{3, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8.584962500721156},
			{1, 8.584962500721156},
		},
		"Amplitude": {
			{0, -2},
		},
	}},
	{4, map[string][]*audio.ControlPoint{
		"Amplitude": {
			{0, -2},
		},
		"Pitch": {
			{0, 8.169925001442312},
			{1, 8.169925001442312},
		},
	}},
	{5, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8.491853096329674},
			{1, 8.491853096329674},
		},
		"Amplitude": {
			{0, -2},
		},
	}},
	{6, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 9.169925001442312},
			{1, 9.169925001442312},
		},
		"Amplitude": {
			{0, -2},
		},
	}},
	{7, map[string][]*audio.ControlPoint{
		"Amplitude": {
			{0, -2},
		},
		"Pitch": {
			{0, 8.906890595608518},
			{1, 8.906890595608518},
		},
	}},
	{8, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8.321928094887362},
			{1, 8.321928094887362},
		},
		"Amplitude": {
			{0, -2},
		},
	}},
	{9, map[string][]*audio.ControlPoint{
		"Amplitude": {
			{0, -2},
		},
		"Pitch": {
			{0, 8.584962500721156},
			{1, 8.584962500721156},
		},
	}},
	{10, map[string][]*audio.ControlPoint{
		"Amplitude": {
			{0, -2},
		},
		"Pitch": {
			{0, 7.906890595608518},
			{1, 7.906890595608518},
		},
	}},
	{11, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8.584962500721156},
			{1, 8.584962500721156},
		},
		"Amplitude": {
			{0, -2},
		},
	}},
	{12, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8},
			{1, 8},
		},
		"Amplitude": {
			{0, -2},
		},
	}},
	{12, map[string][]*audio.ControlPoint{
		"Amplitude": {
			{0, -2},
		},
		"Pitch": {
			{0, 8.321928094887362},
			{1, 8.321928094887362},
		},
	}},
	{13, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8.321928094887362},
			{1, 8.321928094887362},
		},
		"Amplitude": {
			{0, -2},
		},
	}},
	{13, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8.584962500721156},
			{1, 8.584962500721156},
		},
		"Amplitude": {
			{0, -2},
		},
	}},
	{14, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 9},
			{1, 9},
		},
		"Amplitude": {
			{0, -2},
		},
	}},
	{14, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 9.321928094887362},
			{1, 9.321928094887362},
		},
		"Amplitude": {
			{0, -2},
		},
	}},
	{15, map[string][]*audio.ControlPoint{
		"Amplitude": {
			{0, -2},
		},
		"Pitch": {
			{0, 9.584962500721156},
			{1, 9.584962500721156},
		},
	}},
	{15, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8.736965594166206},
			{1, 8.736965594166206},
		},
		"Amplitude": {
			{0, -2},
		},
	}},
	{16, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 9.169925001442312},
			{1, 9.169925001442312},
		},
		"Amplitude": {
			{0, -2},
		},
	}},
	{16, map[string][]*audio.ControlPoint{
		"Amplitude": {
			{0, -2},
		},
		"Pitch": {
			{0, 8.491853096329674},
			{1, 8.491853096329674},
		},
	}},
}, map[string][]*audio.ControlPoint{
	"Distortion": {
		{0, -2},
	},
})
