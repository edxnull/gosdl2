package main

import (
	"image"
)

// There are probably ways how to optimize this data structure.
// One way that I can think of of the top of my head would be
// something where we have a fixed number, or even a floating point number
// to represent LineNr.fraction where fraction is the word count.
// Also, should I change LineNr to uint later? Or maybe even uint16?
// Cuz, how many lines are we gonna have? int64 is an overkill!

// TODO: add LineWidth (as per main.go 375)
type WordRects struct {
	Rect   image.Rectangle
	LineNr int
}

// DB Types
// change Tags to something better (like an enum lookup map or smth)
type DBEntry map[string]*DBVal

type DBVal struct {
	Value string
	Tags  []string
}
