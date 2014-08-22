package main

import "code.google.com/p/gordon-go/audio"

var score = &audio.Score{[]*audio.Part{
	{"Sines", []*audio.PatternEvent{
		{0, melody},
	}},
	{"Sines2", []*audio.PatternEvent{
		{2, melody},
	}},
}}
