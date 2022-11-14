package main

import (
	"github.com/veandco/go-sdl2/sdl"
)

// NOTE 
// Do we need c *sdl.Color here instead of a little copying?
// sdl.Color without the pointer ref would probably be better?
// I would have to write tests to verify that this is actually true!

func draw_rect_with_border(renderer *sdl.Renderer, rect *sdl.Rect, c *sdl.Color) {
	renderer.SetDrawColor((*c).R, (*c).G, (*c).B, (*c).A)
	renderer.DrawRect(rect)
}

func draw_rect_with_border_filled(renderer *sdl.Renderer, rect *sdl.Rect, c *sdl.Color) {
	renderer.SetDrawColor((*c).R, (*c).G, (*c).B, (*c).A)
	renderer.FillRect(rect)
	renderer.DrawRect(rect)
}

func draw_rect_without_border(renderer *sdl.Renderer, rect *sdl.Rect, c *sdl.Color) {
	renderer.SetDrawColor((*c).R, (*c).G, (*c).B, (*c).A)
	renderer.FillRect(rect)
}

func draw_multiple_rects_with_border(renderer *sdl.Renderer, rects []sdl.Rect, c *sdl.Color) {
	renderer.SetDrawColor((*c).R, (*c).G, (*c).B, (*c).A)
	renderer.DrawRects(rects)
}

func draw_multiple_rects_with_border_filled(renderer *sdl.Renderer, rects []sdl.Rect, c *sdl.Color) {
	renderer.SetDrawColor((*c).R, (*c).G, (*c).B, (*c).A)
	renderer.FillRects(rects)
	renderer.DrawRects(rects)
}

func draw_multiple_rects_without_border_filled(renderer *sdl.Renderer, rects []sdl.Rect, c *sdl.Color) {
	renderer.SetDrawColor((*c).R, (*c).G, (*c).B, (*c).A)
	renderer.FillRects(rects)
}

func draw_rounded_rect_with_border_filled(renderer *sdl.Renderer, rect *sdl.Rect, c *sdl.Color) {
	renderer.SetDrawColor((*c).R, (*c).G, (*c).B, (*c).A)
	renderer.FillRect(rect)
	renderer.DrawRect(rect)
	renderer.SetDrawColor(255, 255, 255, 255) // temporary
	renderer.DrawPoints([]sdl.Point{
		sdl.Point{rect.X, rect.Y},                           // top
		sdl.Point{rect.X, rect.Y + rect.H - 1},              // bottom
		sdl.Point{rect.X + rect.W - 1, rect.Y},              // top
		sdl.Point{rect.X + rect.W - 1, rect.Y + rect.H - 1}, // bottom
	})
}
