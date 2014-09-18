package main

import "code.google.com/p/gordon-go/audio"

var score = &audio.Score{[]*audio.Part{
	{"Sines", []*audio.PatternEvent{
		{0, sines_pattern},
	}},
	{"Reverb", []*audio.PatternEvent{
		{0, reverb_pattern},
	}},
}}
