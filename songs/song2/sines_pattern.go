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
			{0, -3},
		},
	}},
	{3, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8.584962500721156},
			{1, 8.584962500721156},
		},
		"Amplitude": {
			{0, -3},
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
		"Amplitude": {
			{0, -3},
		},
		"Pitch": {
			{0, 9.169925001442312},
			{1, 9.169925001442312},
		},
	}},
	{7, map[string][]*audio.ControlPoint{
		"Amplitude": {
			{0, -3},
		},
		"Pitch": {
			{0, 8.906890595608518},
			{1, 8.906890595608518},
		},
	}},
	{8, map[string][]*audio.ControlPoint{
		"Amplitude": {
			{0, -2},
		},
		"Pitch": {
			{0, 8.321928094887362},
			{1, 8.321928094887362},
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
		"Pitch": {
			{0, 7.906890595608518},
			{1, 7.906890595608518},
		},
		"Amplitude": {
			{0, -3},
		},
	}},
	{11, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8.584962500721156},
			{1, 8.584962500721156},
		},
		"Amplitude": {
			{0, -3},
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
	{12, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8},
			{1, 8},
		},
		"Amplitude": {
			{0, -2},
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
		"Amplitude": {
			{0, -2},
		},
		"Pitch": {
			{0, 8.584962500721156},
			{1, 8.584962500721156},
		},
	}},
	{14, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 9},
			{1, 9},
		},
		"Amplitude": {
			{0, -3},
		},
	}},
	{14, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 9.321928094887362},
			{1, 9.321928094887362},
		},
		"Amplitude": {
			{0, -3},
		},
	}},
	{15, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8.736965594166206},
			{1, 8.736965594166206},
		},
		"Amplitude": {
			{0, -3},
		},
	}},
	{15, map[string][]*audio.ControlPoint{
		"Amplitude": {
			{0, -3},
		},
		"Pitch": {
			{0, 9.584962500721156},
			{1, 9.584962500721156},
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
	{17, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8.169925001442312},
			{1, 8.169925001442312},
		},
		"Amplitude": {
			{0, -2},
		},
	}},
	{17, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8.754887502163468},
			{1, 8.754887502163468},
		},
		"Amplitude": {
			{0, -2},
		},
	}},
	{18, map[string][]*audio.ControlPoint{
		"Amplitude": {
			{0, -3},
		},
		"Pitch": {
			{0, 9.754887502163468},
			{1, 9.754887502163468},
		},
	}},
	{18, map[string][]*audio.ControlPoint{
		"Amplitude": {
			{0, -3},
		},
		"Pitch": {
			{0, 8.584962500721156},
			{1, 8.584962500721156},
		},
	}},
	{19, map[string][]*audio.ControlPoint{
		"Amplitude": {
			{0, -4},
		},
		"Pitch": {
			{0, 9.169925001442312},
			{1, 9.169925001442312},
		},
	}},
	{19, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 9.906890595608518},
			{1, 9.906890595608518},
		},
		"Amplitude": {
			{0, -3},
		},
	}},
	{19, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8.491853096329674},
			{1, 8.491853096329674},
		},
		"Amplitude": {
			{0, -3},
		},
	}},
	{20, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8},
			{1, 8},
		},
		"Amplitude": {
			{0, -2},
		},
	}},
	{20, map[string][]*audio.ControlPoint{
		"Amplitude": {
			{0, -3},
		},
		"Pitch": {
			{0, 8.321928094887362},
			{1, 8.321928094887362},
		},
	}},
	{20, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 9.321928094887362},
			{1, 9.321928094887362},
		},
		"Amplitude": {
			{0, -3},
		},
	}},
	{21, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 7.584962500721156},
			{1, 7.584962500721156},
		},
		"Amplitude": {
			{0, -3},
		},
	}},
	{21.018666666252102, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8.169925001442312},
			{0.981333333747898, 8.169925001442312},
		},
		"Amplitude": {
			{0, -2},
		},
	}},
	{22, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8.321928094887362},
			{1, 8.321928094887362},
		},
		"Amplitude": {
			{0, -2},
		},
	}},
	{22, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8},
			{1, 8},
		},
		"Amplitude": {
			{0, -3},
		},
	}},
	{23, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 6.584962500721156},
			{1, 6.584962500721156},
		},
		"Amplitude": {
			{0, -3},
		},
	}},
	{23, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 7.584962500721156},
			{1, 7.584962500721156},
		},
		"Amplitude": {
			{0, -3},
		},
	}},
	{23, map[string][]*audio.ControlPoint{
		"Amplitude": {
			{0, -3},
		},
		"Pitch": {
			{0, 8.169925001442312},
			{1, 8.169925001442312},
		},
	}},
	{24, map[string][]*audio.ControlPoint{
		"Amplitude": {
			{0, -2},
		},
		"Pitch": {
			{0, 7},
			{1, 7},
		},
	}},
	{24, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8.321928094887362},
			{1, 8.321928094887362},
		},
		"Amplitude": {
			{0, -3},
		},
	}},
	{25, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8.169925001442312},
			{1, 8.169925001442312},
		},
		"Amplitude": {
			{0, -3},
		},
	}},
	{25, map[string][]*audio.ControlPoint{
		"Amplitude": {
			{0, -2},
		},
		"Pitch": {
			{0, 7.584962500721156},
			{1, 7.584962500721156},
		},
	}},
	{26, map[string][]*audio.ControlPoint{
		"Amplitude": {
			{0, -2},
		},
		"Pitch": {
			{0, 8},
			{1, 8},
		},
	}},
	{26, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8.321928094887362},
			{1, 8.321928094887362},
		},
		"Amplitude": {
			{0, -2},
		},
	}},
	{27, map[string][]*audio.ControlPoint{
		"Amplitude": {
			{0, -3},
		},
		"Pitch": {
			{0, 6.584962500721156},
			{1, 6.584962500721156},
		},
	}},
	{27, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8.169925001442312},
			{1, 8.169925001442312},
		},
		"Amplitude": {
			{0, -3},
		},
	}},
	{28, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8},
			{1, 8},
		},
		"Amplitude": {
			{0, -2},
		},
	}},
	{28, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8.321928094887362},
			{1, 8.321928094887362},
		},
		"Amplitude": {
			{0, -2},
		},
	}},
	{28, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 6},
			{1, 6},
		},
		"Amplitude": {
			{0, -2},
		},
	}},
	{29, map[string][]*audio.ControlPoint{
		"Amplitude": {
			{0, -2},
		},
		"Pitch": {
			{0, 8.584962500721156},
			{1, 8.584962500721156},
		},
	}},
	{29, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 6.584962500721156},
			{1, 6.584962500721156},
		},
		"Amplitude": {
			{0, -3},
		},
	}},
	{29, map[string][]*audio.ControlPoint{
		"Amplitude": {
			{0, -3},
		},
		"Pitch": {
			{0, 8.321928094887362},
			{1, 8.321928094887362},
		},
	}},
	{30, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 9},
			{1, 9},
		},
		"Amplitude": {
			{0, -3},
		},
	}},
	{30, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 7},
			{1, 7},
		},
		"Amplitude": {
			{0, -3},
		},
	}},
	{30, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8.584962500721156},
			{1, 8.584962500721156},
		},
		"Amplitude": {
			{0, -3},
		},
	}},
	{31, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8.321928094887362},
			{1, 8.321928094887362},
		},
		"Amplitude": {
			{0, -3},
		},
	}},
	{31, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 7.321928094887362},
			{1, 7.321928094887362},
		},
		"Amplitude": {
			{0, -3},
		},
	}},
	{31, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8.584962500721156},
			{1, 8.584962500721156},
		},
		"Amplitude": {
			{0, -3},
		},
	}},
	{32, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8.754887502163468},
			{1, 8.754887502163468},
		},
		"Amplitude": {
			{0, -3},
		},
	}},
	{32, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8.169925001442312},
			{1, 8.169925001442312},
		},
		"Amplitude": {
			{0, -3},
		},
	}},
	{32, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 7.491853096329675},
			{1, 7.491853096329675},
		},
		"Amplitude": {
			{0, -2},
		},
	}},
	{33, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 7.584962500721156},
			{1, 7.584962500721156},
		},
		"Amplitude": {
			{0, -3},
		},
	}},
	{33, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8.491853096329674},
			{1, 8.491853096329674},
		},
		"Amplitude": {
			{0, -2},
		},
	}},
	{33, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8.169925001442312},
			{1, 8.169925001442312},
		},
		"Amplitude": {
			{0, -3},
		},
	}},
	{34, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 9.169925001442312},
			{1, 9.169925001442312},
		},
		"Amplitude": {
			{0, -4},
		},
	}},
	{34, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 7.754887502163468},
			{1, 7.754887502163468},
		},
		"Amplitude": {
			{0, -3},
		},
	}},
	{34, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8.754887502163468},
			{1, 8.754887502163468},
		},
		"Amplitude": {
			{0, -4},
		},
	}},
	{35, map[string][]*audio.ControlPoint{
		"Amplitude": {
			{0, -3},
		},
		"Pitch": {
			{0, 7.906890595608519},
			{1, 7.906890595608519},
		},
	}},
	{35, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8.584962500721156},
			{1, 8.584962500721156},
		},
		"Amplitude": {
			{0, -3},
		},
	}},
	{35, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8.906890595608518},
			{1, 8.906890595608518},
		},
		"Amplitude": {
			{0, -4},
		},
	}},
	{36, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8.321928094887362},
			{1, 8.321928094887362},
		},
		"Amplitude": {
			{0, -3},
		},
	}},
	{36, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 9},
			{1, 9},
		},
		"Amplitude": {
			{0, -3},
		},
	}},
	{36, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8},
			{1, 8},
		},
		"Amplitude": {
			{0, -2},
		},
	}},
}, map[string][]*audio.ControlPoint{
	"Distortion": {
		{0, -2},
	},
})
