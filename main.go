package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/golang/freetype"
	"github.com/veandco/go-sdl2/sdl"
)

var (
	cpuprof = flag.String("cpuprofile", "", "write cpu profile to 'file'")
	memprof = flag.String("memprofile", "", "write mem profile to 'file'")

	fontStr = flag.String("font", "", "usage: -font=<fname>.<ftype>")
	textStr = flag.String("text", "", "usage: -text=<fname>.<ftype>")
)

func MouseOverWords(event *sdl.MouseMotionEvent, ctx *freetype.Context, r *[]WordRects, mouseOver *[]bool) {
	var fontSize = ctx.PointToFixed(18.0).Round()
	for index := range *r {
		mx_gt_rx := int(event.X) > (*r)[index].Rect.Min.X
		mx_lt_rx_rw := int(event.X) < (*r)[index].Rect.Max.X
		my_gt_ry := int(event.Y) > (*r)[index].Rect.Min.Y-fontSize
		my_lt_ry_rh := int(event.Y) < (*r)[index].Rect.Max.Y

		if (mx_gt_rx && mx_lt_rx_rw) && (my_gt_ry && my_lt_ry_rh) {
			(*mouseOver)[index] = true
		} else {
			(*mouseOver)[index] = false
		}
	}
}

func main() {
	flag.Parse()

	if *cpuprof != "" {
		cpuf, err := os.Create(*cpuprof)
		if err != nil {
			log.Fatal("could not *create* CPU profile: ", err)
		}
		defer cpuf.Close()
		if err := pprof.StartCPUProfile(cpuf); err != nil {
			log.Fatal("could not *start* CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	runtime.LockOSThread()

	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		panic(err)
	}

	window, err := sdl.CreateWindow("", sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, 640, 480,
		sdl.WINDOW_SHOWN|sdl.WINDOW_RESIZABLE)
	if err != nil {
		panic(err)
	}

	// NOTE: I've heard that PRESENTVSYNC caps FPS
	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)
	if err != nil {
		panic(err)
	}

	ticker := time.NewTicker(time.Second / 60)

	sdl.SetHint(sdl.HINT_FRAMEBUFFER_ACCELERATION, "1")
	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1")

	renderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND)

	bgrect := sdl.Rect{X: 0, Y: 0, W: 640, H: 480}
	testTex, err := renderer.CreateTexture(uint32(sdl.PIXELFORMAT_RGBA32), sdl.TEXTUREACCESS_STREAMING, 640, 480)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer testTex.Destroy()
	testTex.SetBlendMode(sdl.BLENDMODE_BLEND)

	var fontDst string
	var textDst string

	const (
		textDir     string = "./text/"
		fontDir     string = "./fonts/"
		defaultFont string = "AnonymousPro-Regular.ttf"
		defaultText string = "HP01.txt"
	)

	if *textStr == "" {
		textDst = textDir + defaultText
	} else {
		textDst = textDir + *textStr
	}

	textData, err := ioutil.ReadFile(textDst)
	if err != nil {
		fmt.Println(err)
		return
	}

	println("[debug] got here!")

	testTokens := WrapLines(string(textData), 400, 18/2)

	println("[debug] got here!")

	if *fontStr == "" {
		fontDst = fontDir + defaultFont
	} else {
		fontDst = fontDir + *fontStr
	}

	fontBytes, err := ioutil.ReadFile(fontDst)
	if err != nil {
		fmt.Println(err)
		return
	}

	parsedFont, err := freetype.ParseFont(fontBytes)
	if err != nil {
		fmt.Println(err)
		return
	}

	fontSize := float64(18.0)

	fontBGColor := image.NewUniform(color.RGBA{255, 255, 255, 255})
	fontFGColor := image.NewUniform(color.RGBA{0, 0, 0, 255})

	bg := image.NewRGBA(image.Rect(0, 0, 640, 480))

	draw.Draw(bg, bg.Bounds(), fontBGColor, image.Point{0, 0}, draw.Src)

	ctx := freetype.NewContext()
	ctx.SetFont(parsedFont)
	ctx.SetDPI(72)
	ctx.SetFontSize(fontSize)
	ctx.SetClip(bg.Bounds())
	ctx.SetDst(bg)
	ctx.SetSrc(fontFGColor)

	textWindowOffset := 10

	pt := freetype.Pt(textWindowOffset, 20)

	var (
		zoomIn       bool
		zoomOut      bool
		moveLineUp   bool
		moveLineDown bool
		movePageUp   bool
		movePageDown bool

		mouseButtonClicked           bool
		mouseButtonClickedAndDragged bool

		running bool = true

		newFontSize float64
		startIndex  int = 0
		numLines    int = 24
	)

	// TODO(read): https://developer.apple.com/fonts/TrueType-Reference-Manual/RM02/Chap2.html#intro
	// TODO(read): https://golang.hotexamples.com/ru/examples/github.com.golang.freetype.truetype/Font/FUnitsPerEm/golang-font-funitsperem-method-examples.html
	// TODO(read): https://bit.ly/2kjbenG

	// ---- page allocs ----
	numAllocs := 0

	for i := 0; i < numLines; i++ {
		numAllocs += CountSpacesBetweenWords(testTokens[i])
	}

	numAllocs = numAllocs * 2 // alloc size subject to change

	word_rects := make([]WordRects, numAllocs)

	mouse_over := make([]bool, numAllocs)
	// ---- page allocs ----

	DrawToCtx(bg, ctx, pt, &testTokens, parsedFont, startIndex, numLines, fontSize, &word_rects)

	fmt.Printf("len of page_elem_len_x %d\n", len(word_rects))

	var anim func() bool

	testTex.Update(&bgrect, bg.Pix, bg.Stride)

	// ----- database test -----
	db := DBOpen()
	defer db.Close()

	known_word_data := GetUniqueWords(testTokens)

	// DB stuff
	if err = DBInit(db, known_word_data); err != nil {
		fmt.Printf("Something went wrong %v", err)
	}

	//if err = DBInsert(db, "hobbit"); err != nil {
	//	fmt.Printf("Something went wrong %v", err)
	//}
	// ----- database test -----

	hiStartX := int32(textWindowOffset)
	hiStartY := int32(0) // int32(ctx.PointToFixed(fontSize))

    // + 1 because we have a 0 based indexing
	hiRects := NewHiLineRects(numLines+1, hiStartX, hiStartY)

	var (
		clearScreen    bool
		word_rect_indx int = -1
		selectedLine   int32
	)

    notYetSet := true
    var startSelectionRange int

	highlighted := false
	testSelect := false

	testSelectIndexStart := 0
	testSelectIndexEnd := 0
	testSelectYCoord := 0

	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.MouseMotionEvent:
				MouseOverWords(t, ctx, &word_rects, &mouse_over)

				// check go doc sdl.MouseMotionEvent for solution
				if testSelect {
					testSelectIndexEnd = int(t.X)
					mouseButtonClickedAndDragged = true
				}
				if mouseButtonClickedAndDragged {
					testSelectYCoord = int(t.Y)
				}
			case *sdl.MouseWheelEvent:
				switch {
				case t.Y > 0:
					moveLineUp = true
				case t.Y < 0:
					moveLineDown = true
				}
			case *sdl.MouseButtonEvent:
				switch t.Type {
				case sdl.MOUSEBUTTONDOWN:
				case sdl.MOUSEBUTTONUP:
					mouseButtonClicked = true
				}

				if !testSelect && t.Type == sdl.MOUSEBUTTONDOWN && t.State == sdl.PRESSED {
					testSelect = true
					testSelectIndexStart = int(t.X)
					highlighted = false
				}

				if testSelect && t.Type == sdl.MOUSEBUTTONUP && t.State == sdl.RELEASED {
					testSelect = false
					highlighted = true
				}
			case *sdl.KeyboardEvent:
				if t.Keysym.Sym == sdl.K_ESCAPE {
					running = false
				}

				switch t.Type {
				case sdl.KEYDOWN:
				case sdl.KEYUP:
					switch t.Keysym.Sym {
					case sdl.K_f:
						if !zoomIn {
							zoomIn = true
							newFontSize = fontSize + 2.0
						}
					case sdl.K_b:
						if !zoomOut {
							zoomOut = true
							newFontSize = fontSize - 2.0
						}
					case sdl.K_UP:
						moveLineUp = true
					case sdl.K_DOWN:
						moveLineDown = true
					case sdl.K_LEFT:
						movePageUp = true
					case sdl.K_RIGHT:
						movePageDown = true
					}
				}
			default:
				continue
			}
		}
		renderer.SetDrawColor(255, 255, 255, 255)
		renderer.Clear()

		renderer.Copy(testTex, nil, &bgrect)

		if mouseButtonClicked {
			for i := 0; i < len(mouse_over); i++ {
				if mouse_over[i] == true && word_rect_indx != i {
					if word_rect_indx < 0 { // guard against -1 index
						word_rect_indx = 0
					}

					// clear
					colorG := image.NewUniform(color.RGBA{0, 255, 0, 108})
					draw.Draw(bg, word_rects[word_rect_indx].Rect, colorG, image.Point{0, 0}, draw.Src)
					testTex.Update(&bgrect, bg.Pix, bg.Stride)

					// draw
					colorB := image.NewUniform(color.RGBA{0, 0, 244, 108})
					draw.Draw(bg, word_rects[i].Rect, colorB, image.Point{0, 0}, draw.Src)
					testTex.Update(&bgrect, bg.Pix, bg.Stride)

					word_rect_indx = i
				}

				if mouse_over[i] == true && word_rect_indx == i {
					w := GetWord(&testTokens, &word_rects, word_rect_indx)
					exists, err := DBView(db, w)
					if err != nil {
						fmt.Println(err)
					}

					const msg = "'%s' exists in the database = %q\n"
					fmt.Printf(msg, w, exists)

					clearScreen = false
					break
				}

				if word_rect_indx >= 0 {
					clearScreen = true
				}
			}
			mouseButtonClicked = false
		}

		if mouseButtonClickedAndDragged {
			start := GetSelectedCharLen(testSelectIndexStart)
			if start > 0 {
				start -= 1
			}
			end := GetSelectedCharLen(testSelectIndexEnd)

			lineHeight := ctx.PointToFixed(fontSize).Round()
            // hiRects.rects[selectedLine].H = int32(lineHeight)

			selectedLine = int32(YCoordToNumLines(testSelectYCoord, lineHeight))

			markLast := false
			wentDown := false

			if hiRects.currln < selectedLine {
				hiRects.currln = selectedLine
				wentDown = true
			} else if hiRects.currln > selectedLine {
				markLast = true
			}

			var lineY int
			lineY = lineHeight * (int(selectedLine) + 1)

			// used for selecting testToken string after movePageDown/Up moveLineUp/Down
			startIndex32 := int32(startIndex)

			if selectedLine <= int32(numLines) {
                println(selectedLine)
				hiRects.rects[selectedLine].Y = int32(lineY - textWindowOffset)
				// hiRects.rects[selectedLine].H = int32(lineHeight)

				// TODO
				// we are already calling WidthOfString all over our codebase, so wouldn't
				// it be better to somehow memoize the results so that we don't have to call it as much?
				maxWidth := int(WidthOfString(parsedFont, fontSize, testTokens[selectedLine+startIndex32]))

				// we need to track testSelectIndexStart
				// (!) we can probably solve our selection problems by using testSelectIndexStart
				if testSelectIndexStart >= textWindowOffset && testSelectIndexStart <= maxWidth && notYetSet {
					// fmt.Printf("ok... we got our position x: %d\n", testSelectIndexStart)
                    hiRects.rects[selectedLine].X = int32(testSelectIndexStart)
                    notYetSet = false
                    startSelectionRange = int(selectedLine)
				}

				if testSelectIndexEnd < maxWidth && !wentDown {
					hiRects.rects[selectedLine].W = int32(testSelectIndexEnd) - hiRects.rects[selectedLine].X
				} else if testSelectIndexEnd < maxWidth && wentDown || testSelectIndexEnd >= maxWidth && wentDown {
					prevSelectedLine := selectedLine
					if prevSelectedLine > 0 {
						prevSelectedLine -= 1
					}
					maxWidth = int(WidthOfString(parsedFont, fontSize, testTokens[prevSelectedLine+startIndex32]))
					hiRects.rects[prevSelectedLine].W = int32(maxWidth) - hiRects.rects[selectedLine-1].X

                    // this is why we couldn't properly render maxWidth when we start at a latter X
                    // it's because we need to select the prevLine instead of selectedLine! 
                    // prevLine == selectedLine-1
                    // println(maxWidth, int32(maxWidth) - hiRects.rects[selectedLine-1].X, hiRects.rects[selectedLine-1].X)
				} else if testSelectIndexEnd > maxWidth {
					hiRects.rects[selectedLine].W = int32(maxWidth)
				}

				if markLast {
					hiRects.UnShowRangeFrom(int(selectedLine))
					hiRects.currln = selectedLine
					markLast = false
				}

                // loop through and select range from start up to selectedLine
                // so that we catch any lines that were missed
				hiRects.Show(startSelectionRange, int(selectedLine))

                for i := int32(0); i < selectedLine; i++ {
                    if hiRects.IsShown(int(i)) {
                        // we don't set rects[i].W on selection start
                        if i == int32(startSelectionRange) {
                            draw_rect_without_border(renderer, &hiRects.rects[i], &sdl.Color{R: 200, G: 100, B: 80, A: 100})
                            continue
                        }
                        // the following two lines are needed to set the Y properly
                        lineY = lineHeight * (int(i) + 1)
                        hiRects.rects[i].Y = int32(lineY - textWindowOffset)

                        hiRects.rects[i].W = int32(WidthOfString(parsedFont, fontSize, testTokens[i+startIndex32]))
                        draw_rect_without_border(renderer, &hiRects.rects[i], &sdl.Color{R: 200, G: 100, B: 80, A: 100})
                    }
                }
			}

			if highlighted {
				max_str_len := len(testTokens[selectedLine+startIndex32])

                // completely broken!
				// this branch is not guaranteed to be bug free (IT HAS TO BE TESTED!)
				// switch {case:, case:, ...} also has to be tested!
				if start < max_str_len && end < max_str_len {
					switch {
					case end > 0:
						println("case: end > 0")
						fmt.Println(testTokens[selectedLine+startIndex32][start:end])
					case end == 0:
						println("case: end == 0")
						fmt.Println(testTokens[selectedLine+startIndex32][start : start+1])
					case end < start:
						println("case: end < start")
						fmt.Println(testTokens[selectedLine+startIndex32][start : start+end])
					case end+start >= max_str_len:
						println("case: end + start >= max_str_len")
						fmt.Println(testTokens[selectedLine+startIndex32][start:])
					// (!) not sure about this one
					case end == start:
						println("case: end == start")
						fmt.Println(testTokens[selectedLine+startIndex32][start : start+1])
					// (!) not sure about this one
					case end == 0 && start == 0:
						println("case: end == 0 && start == 0")
						break
					}
				}

                // reset
                notYetSet = true
                startSelectionRange = 0
				mouseButtonClickedAndDragged = false
				hiRects.UnShowAllAndReset(hiStartX)
			}
		}

		if clearScreen {
			colorG := image.NewUniform(color.RGBA{0, 255, 0, 108})
			draw.Draw(bg, word_rects[word_rect_indx].Rect, colorG, image.Point{0, 0}, draw.Src)
			testTex.Update(&bgrect, bg.Pix, bg.Stride)
			word_rect_indx = -1
			clearScreen = false
		}

		if movePageUp {
			movePageUp = false
			if startIndex-numLines < 0 {
				startIndex = 0
			} else if startIndex > 0 {
				startIndex -= numLines
			}

			// clear
			draw.Draw(bg, bg.Bounds(), fontBGColor, image.Point{0, 0}, draw.Src)

			// redraw with increased fontSize and update
			ctx.SetFontSize(fontSize)

			// do we need to call freetype.Pt() here? Can't we just pt.X, pt.Y = ?, ?
			pt = freetype.Pt(textWindowOffset, 20)

			DrawToCtx(bg, ctx, pt, &testTokens, parsedFont, startIndex, numLines, fontSize, &word_rects)

			testTex.Update(&bgrect, bg.Pix, bg.Stride)
		}

		if moveLineUp {
			moveLineUp = false
			if startIndex > 0 {
				startIndex -= 1
			}

			// clear
			draw.Draw(bg, bg.Bounds(), fontBGColor, image.Point{0, 0}, draw.Src)

			// redraw with increased fontSize and update
			ctx.SetFontSize(fontSize)

			// do we need to call freetype.Pt() here? Can't we just pt.X, pt.Y = ?, ?
			pt = freetype.Pt(textWindowOffset, 20)

			DrawToCtx(bg, ctx, pt, &testTokens, parsedFont, startIndex, numLines, fontSize, &word_rects)

			testTex.Update(&bgrect, bg.Pix, bg.Stride)
		}

		if movePageDown {
			movePageDown = false

			if startIndex >= 0 {
				startIndex += numLines
			}

			// clear
			draw.Draw(bg, bg.Bounds(), fontBGColor, image.Point{0, 0}, draw.Src)

			// redraw with increased fontSize and update
			ctx.SetFontSize(fontSize)

			// do we need to call freetype.Pt() here? Can't we just pt.X, pt.Y = ?, ?
			pt = freetype.Pt(textWindowOffset, 20)

			DrawToCtx(bg, ctx, pt, &testTokens, parsedFont, startIndex, numLines, fontSize, &word_rects)

			testTex.Update(&bgrect, bg.Pix, bg.Stride)
		}

		if moveLineDown {
			moveLineDown = false
			startIndex += 1

			// clear
			draw.Draw(bg, bg.Bounds(), fontBGColor, image.Point{0, 0}, draw.Src)

			// redraw with increased fontSize and update
			ctx.SetFontSize(fontSize)

			// do we need to call freetype.Pt() here? Can't we just pt.X, pt.Y = ?, ?
			pt = freetype.Pt(textWindowOffset, 20)

			DrawToCtx(bg, ctx, pt, &testTokens, parsedFont, startIndex, numLines, fontSize, &word_rects)

			testTex.Update(&bgrect, bg.Pix, bg.Stride)
		}

		if zoomIn {
			if anim == nil {
				anim = EasingAnimate(&fontSize, newFontSize, EaseInOutQuad, "animIn")
			}

			fmt.Println(fontSize)
			if !anim() {
				anim = nil
				zoomIn = false
				println("END of animIn")
			}

			// clear
			draw.Draw(bg, bg.Bounds(), fontBGColor, image.Point{0, 0}, draw.Src)

			// redraw with increased fontSize and update
			ctx.SetFontSize(fontSize)

			// do we need to call freetype.Pt() here? Can't we just pt.X, pt.Y = ?, ?
			pt = freetype.Pt(textWindowOffset, 20)

			DrawToCtx(bg, ctx, pt, &testTokens, parsedFont, startIndex, numLines, fontSize, &word_rects)

			testTex.Update(&bgrect, bg.Pix, bg.Stride)
		}

		if zoomOut {
			if anim == nil {
				anim = EasingAnimate(&fontSize, newFontSize, EaseInOutQuad, "animOut")
			}

			fmt.Println(fontSize)
			if !anim() {
				anim = nil
				zoomOut = false
				println("END of animOut")
			}

			// clear
			draw.Draw(bg, bg.Bounds(), fontBGColor, image.Point{0, 0}, draw.Src)

			// redraw with increased fontSize and update
			ctx.SetFontSize(fontSize)

			// do we need to call freetype.Pt() here? Can't we just pt.X, pt.Y = ?, ?
			pt = freetype.Pt(textWindowOffset, 20)

			DrawToCtx(bg, ctx, pt, &testTokens, parsedFont, startIndex, numLines, fontSize, &word_rects)

			testTex.Update(&bgrect, bg.Pix, bg.Stride)
		}

		// we don't have to do this on every frame
		// figure out a way to make it more efficient.
		// use t := time.Now()
		// 1) https://gameprogrammingpatterns.com/game-loop.html
		// 2) https://bell0bytes.eu/the-game-loop//
		// ...

		renderer.Present()
		<-ticker.C
	}

    // why aren't we defer'ring these?
	sdl.Quit()
	ticker.Stop()
	renderer.Destroy()
	window.Destroy()
	runtime.UnlockOSThread()

	if *memprof != "" {
		memf, err := os.Create(*memprof)
		if err != nil {
			log.Fatal("could not *create* MEM profile: ", err)
		}
		defer memf.Close()
		runtime.GC()
		if err := pprof.WriteHeapProfile(memf); err != nil {
			log.Fatal("could not *start*  MEM profile: ", err)
		}
	}
}
