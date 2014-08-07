package main

import "code.google.com/p/gordon-go/audio"

var melody = &audio.Pattern{"melody", []*audio.Note{
	{0, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8},
			{1, 8},
		},
		"Amplitude": {
			{0, 0},
		},
	}},
	{1, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8},
		},
		"Amplitude": {
			{0, -1},
			{1, -1},
		},
	}},
	{2, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8},
		},
		"Amplitude": {
			{0, -2},
			{1, -2},
		},
	}},
	{3, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8},
		},
		"Amplitude": {
			{0, -3},
			{1, -3},
		},
	}},
	{4, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8},
		},
		"Amplitude": {
			{0, -4},
			{1, -4},
		},
	}},
	{5, map[string][]*audio.ControlPoint{
		"Amplitude": {
			{0, -5},
			{1, -5},
		},
		"Pitch": {
			{0, 8},
		},
	}},
	{6, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8},
		},
		"Amplitude": {
			{0, -6},
			{1, -6},
		},
	}},
	{7, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8},
		},
		"Amplitude": {
			{0, -7},
			{1, -7},
		},
	}},
	{8, map[string][]*audio.ControlPoint{
		"Pitch": {
			{0, 8},
		},
		"Amplitude": {
			{0, -8},
			{1, -8},
		},
	}},
}}
