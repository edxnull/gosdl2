package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/math/fixed"
)

func OmitTrailingPunctuation(str string) string {
	if AllNonAlpha(str) {
		return str
	}

	end := len(str) - 1
	for IsAlpha(str[end]) == false {
		end--
	}
	return str[:end+1]
}

func OmitPrecedingPunctuation(str string) string {
	if AllNonAlpha(str) {
		return str
	}

	start := 0
	for IsAlpha(str[start]) == false {
		start++
	}
	return str[start:]
}

func HasNonAlpha(str string) bool {
	for _, c := range []byte(str) {
		if !IsAlpha(c) {
			return true
		}
	}
	return false
}

func AllNonAlpha(str string) bool {
	for _, c := range []byte(str) {
		if IsAlpha(c) {
			return false
		}
	}
	return true
}

func HasCapitalLetter(str string) bool {
	for _, c := range []byte(str) {
		if IsCapital(c) {
			return true
		}
	}
	return false
}

func IsCapital(c byte) bool {
	return (c >= byte('A')) && (c <= byte('Z'))
}

func IsAlpha(c byte) bool {
	return (c >= byte('A')) && (c <= byte('z'))
}

// It is possible that in the future the behaviour of this function will have to change.
func GetUniqueWords(s []string) DBEntry {
	mk := make(DBEntry)
	for i := 0; i < len(s); i++ {
		words := strings.Split(s[i], " ")
		for _, w := range words {
			if HasNonAlpha(w) {
				trimmed := strings.Trim(w, ",.\n\r\\/\"'-;%^$#*@(!?)_-+=:<>[]{}~|")
				if !AllNonAlpha(w) {
					if _, ok := mk[trimmed]; !ok {
						mk[trimmed] = &DBVal{
							Value: "A",
							Tags:  []string{"a", "b", "c"},
						}
					}
				}
			} else {
				// we don't want to create a new entry for ""
				if w != "" {
					mk[w] = &DBVal{
						Value: "B",
						Tags:  []string{"d", "e", "f"},
					}
				}
			}
		}
		words = nil
	}
	return mk
}

func CountSpacesBetweenWords(str string) int {
	var (
		index  = 0
		result = 0
	)
	for index < len(str) {
		if index == 0 && str[index] == ' ' {
			for str[index] == ' ' { // skip spaces
				index++
				if index == len(str) {
					break
				}
			}
		}
		if str[index] == ' ' && index+1 < len(str) && str[index+1] != ' ' {
			result += 1
			for str[index] == ' ' { // skip spaces
				index++
				if index == len(str) {
					break
				}
			}
		}
		index++
	}
	return result
}

func EaseInOutQuad(b, d, c, t float64) float64 {
	if ((t / d) / 2) < 1 {
		return c/2*(t/d)*(t/d) + b
	}
	return -c/2*((t/d)*((t/d)-2)-1) + b
}

type easingFunc func(b, d, c, t float64) float64

func EasingAnimate(v *float64, max float64, fn easingFunc, anim_t string) func() bool {
	var (
		duration float64 = 0.0
		animIn   bool
		animOut  bool
	)

	switch anim_t {
	case "animIn":
		animIn = true
	case "animOut":
		animOut = true
	}

	return func() bool {
		if animIn {
			*v = fn(*v, max, max, duration)
			duration += 0.5
			if *v >= max {
				animIn = false
				duration = 0.0
				*v = math.Round(*v)
			}
			return true
		}

		if animOut {
			*v -= fn(0, max, max, duration)
			duration += 0.5
			if *v <= max {
				animOut = false
				duration = 0.0
				*v = math.Round(*v)
			}
			return true
		}
		return false
	}
}

func WrapLines(input string, length int, font_w int) []string {
	sizeInPx := int(math.RoundToEven(float64(length/font_w))) - 1
	result := make([]string, getSliceCount(input, sizeInPx))
	pos := 0
	for _, split := range strings.Split(input, "\n") {
		slice := getSlice(split, sizeInPx)
		for s, end := slice(); ; s, end = slice() {
			if s != "" {
				result[pos] = s
				pos += 1
			}
			if end { // move this end to for(...)?
				break
			}
		}
		slice = nil
	}
	if result[len(result)-1] == "" {
		return result[:len(result)-1]
	}
	return result
}

func getSliceCount(input string, linelen int) int32 {
	var result int32
	for _, split := range strings.Split(input, "\n") {
		slice := getSlice(split, linelen)
		for _, end := slice(); ; _, end = slice() {
			result += 1
			if end {
				break
			}
		}
		slice = nil
	}
	return result
}

func getSlice(s string, end int) func() (string, bool) {
	var (
		start int
		last  int
	)
	return func() (string, bool) {
		t := s[:] // TODO: move it down to for i:=0 ....?
		if last+end > len(s) {
			end = len(s) - last
		}
		for i := 0; i < end; i++ {
			_, size := utf8.DecodeRuneInString(t)
			t = t[size:]
			last += size
		}
		t = "" // TODO: is it correct to set t to an empty string?
		if last >= len(s) {
			result := s[start:]
			return result, true
		}
		if !utf8.ValidString(s[start:last]) {
			_, size := utf8.DecodeLastRuneInString(s[start:last])
			last = last - size
		}
		last = strings.LastIndex(s[:last], " ") + 1 // +1 is to remove extra space
		result := s[start:last]
		start = last
		return result, false
	}
}

// stolen from golang's stdlib
func WidthOfString(font *truetype.Font, size float64, s string) float64 {
	scale := size / float64(font.FUnitsPerEm()) // scale converts truetype.FUnit to float64

	width := 0
	prev, hasPrev := truetype.Index(0), false
	for _, rune := range s {
		index := font.Index(rune)
		if hasPrev {
			width += int(font.Kern(fixed.Int26_6(font.FUnitsPerEm()), prev, index))
		}
		width += int(font.HMetric(fixed.Int26_6(font.FUnitsPerEm()), index).AdvanceWidth)
		prev, hasPrev = index, true
	}

	return math.Round(float64(width) * scale)
}

func CharWidths(font *truetype.Font, size float64, s string) []float64 {
	scale := size / float64(font.FUnitsPerEm()) // scale converts truetype.FUnit to float64

	widths := make([]float64, len(s))
	prev, hasPrev := truetype.Index(0), false
	for i, rune := range s {
		var tempVal int
		index := font.Index(rune)
		if hasPrev {
			tempVal += int(font.Kern(fixed.Int26_6(font.FUnitsPerEm()), prev, index))
		}
		tempVal += int(font.HMetric(fixed.Int26_6(font.FUnitsPerEm()), index).AdvanceWidth)
		prev, hasPrev = index, true
		widths[i] = math.Round(float64(tempVal) * scale)
	}

	return widths
}

// refactor!
// hardcoded values! (10)
// potential bugs! when width is not the same accross chars
func GetSelectedCharLen(w int) int {
	if w == 0 {
		return 0
	}

	var index int
	for ; w >= 10; w -= 10 {
		index += 1
	}
	return index
}

// zero based indexing is important because this function is used
// in mouse highlighting scenario, where it's being passed to []testTokens.
func YCoordToNumLines(y int, lineHeight int) int {
	if y == 0 {
		return 0
	}

	var numLines int
	for h := int(lineHeight); y >= h; y -= h {
		numLines += 1
	}
	return numLines
}

func DrawToCtx(bg *image.RGBA, ctx *freetype.Context, pt fixed.Point26_6,
	tokens *[]string, font *truetype.Font,
	startIndex, numLines int,
	fontSize float64, rects *[]WordRects) {
	const windowOffset = 10

	rectIndex := 0
	lineIndex := 0

	lineHeight := ctx.PointToFixed(fontSize)
	spaceWidth := int(WidthOfString(font, fontSize, " "))
	colorGreen := image.NewUniform(color.RGBA{0, 255, 0, 108})

	// clear everything back to 0
	for i := range *rects {
		// maybe it would be beter to set .Rect through
		// Min.X, Min.Y, Max.X, Max.Y instead of image.Rect(...)?
		(*rects)[i].Rect = image.Rect(0, 0, 0, 0)
		(*rects)[i].LineNr = 0
	}

	for n := startIndex; n < numLines+startIndex; n++ {
		if n >= len(*tokens) {
			break
		}

		_, err := ctx.DrawString((*tokens)[n], pt)
		if err != nil {
			fmt.Println(err)
			return
		}

		prevWidth := 0
		currWidth := 0

		var rect image.Rectangle

		for i, word := range strings.Split((*tokens)[n], " ") {
			if word == "\n" || word == "\r" {
				break
			}

			if i == 0 {
				currWidth = windowOffset + int(WidthOfString(font, fontSize, word))

				rect = image.Rect(windowOffset, pt.Y.Round()+2, currWidth, pt.Y.Round()+5)

				(*rects)[rectIndex].Rect = rect
				rectIndex += 1

				draw.Draw(bg, rect, colorGreen, image.Point{0, 0}, draw.Src)

				prevWidth += currWidth
				currWidth = prevWidth + spaceWidth
			} else {
				rect = image.Rect(currWidth, pt.Y.Round()+2,
					currWidth+int(WidthOfString(font, fontSize, word)), pt.Y.Round()+5)

				(*rects)[rectIndex].Rect = rect
				rectIndex += 1

				draw.Draw(bg, rect, colorGreen, image.Point{0, 0}, draw.Src)

				prevWidth = currWidth + int(WidthOfString(font, fontSize, word))
				currWidth = prevWidth + spaceWidth
			}
			(*rects)[lineIndex].LineNr = n
			lineIndex += 1
		}
		pt.Y += lineHeight
	}
}

func GetWord(tokens *[]string, rects *[]WordRects, index int) string {
	wordCount := 0
	LineNr := (*rects)[index].LineNr
	for i := index; (*rects)[i].LineNr == LineNr; i-- {
		wordCount++
		if i == 0 {
			break
		}
	}
	// TODO(optimize): such that we don't use strings.Split() anymore
	desiredWord := strings.Split((*tokens)[LineNr], " ")[wordCount-1]
	noPrecedingPunc := OmitPrecedingPunctuation(desiredWord)

	return OmitTrailingPunctuation(noPrecedingPunc)
}
