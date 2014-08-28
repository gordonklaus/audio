package main

import "code.google.com/p/gordon-go/audio"

var score = &audio.Score{[]*audio.Part{
	{"Sines", []*audio.PatternEvent{
		{0, melody_pattern},
	}},
}}
