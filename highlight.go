package main

import (
	"github.com/veandco/go-sdl2/sdl"
)

type HiLineRects struct {
	rects  []sdl.Rect
	length []int // unused atm
	show   []bool
	currln int32
}

func NewHiLineRects(numLines int, x, y int32) *HiLineRects {
	hilines := HiLineRects{
		rects:  make([]sdl.Rect, numLines),
		length: make([]int, numLines),
		show:   make([]bool, numLines),
		// currln is set to 0 by default
	}

	MAGIC_HEIGHT := int32(18)

	for index := range hilines.rects {
		hilines.rects[index].X = x
		hilines.rects[index].Y = y
		hilines.rects[index].H = MAGIC_HEIGHT
	}
	return &hilines
}

func (hlr *HiLineRects) IsShown(current int) bool {
	if hlr.show[current] {
		return true
	}
	return false
}

func (hlr *HiLineRects) Show(start, end int) {
	for i := start; i < hlr.Len(); i++ {
		if !hlr.IsShown(i) && i <= end {
			hlr.show[i] = true
		}
	}
}

func (hlr *HiLineRects) UnShow(current int) {
	if hlr.show[current] {
		hlr.show[current] = false
	}
}

// maybe we should use textbox offset instead of a 0?
func (hlr *HiLineRects) UnShowAllAndReset(x int32) {
	for index := 0; index < hlr.Len(); index++ {
		if hlr.show[index] {
			hlr.show[index] = false
			hlr.rects[index].X = x
			hlr.length[index] = 0
		}
	}
	hlr.currln = 0
}

func (hlr *HiLineRects) UnShowRangeFrom(rng int) {
	for index := rng; index < hlr.Len(); index++ {
		if hlr.show[index] {
			hlr.show[index] = false
		}
	}
}

func (hlr *HiLineRects) Len() int {
	return len(hlr.rects)
}
