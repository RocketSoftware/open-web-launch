package gui

//go:generate go-bindata-assetfs -pkg gui assets/...

import (
	"bytes"
	"image"
	"image/color"
	"log"
	"regexp"
	"sync"
	"sync/atomic"

	"github.com/aarzilli/nucular"
	"github.com/aarzilli/nucular/label"
	"github.com/aarzilli/nucular/style"
	"github.com/rocketsoftware/open-web-launch/utils"
	"golang.org/x/mobile/event/key"
)

const scaling = 1.2

type GUI struct {
	windowTitle string
	title       atomic.Value
	text        atomic.Value
	progress    int32
	progressMax atomic.Value
	window      nucular.MasterWindow
	ready       chan interface{} // the channel closed when the gui appears for the first time
	readyOnce   sync.Once        // protects ready channel from being closed twice
	icon        *image.RGBA
	err         error
}

// myThemeTable is modified WhiteTheme
var myThemeTable = style.ColorTable{
	ColorText:                  color.RGBA{0x3c, 0x3c, 0x3c, 255}, // modified
	ColorWindow:                color.RGBA{255, 255, 255, 255},    // modified
	ColorHeader:                color.RGBA{175, 175, 175, 255},
	ColorHeaderFocused:         color.RGBA{0xc3, 0x9a, 0x9a, 255},
	ColorBorder:                color.RGBA{0x25, 0x59, 0xA9, 255}, // modified
	ColorButton:                color.RGBA{255, 255, 255, 255},    // modified
	ColorButtonHover:           color.RGBA{255, 255, 255, 255},    // modified
	ColorButtonActive:          color.RGBA{255, 255, 255, 255},    // modified
	ColorToggle:                color.RGBA{150, 150, 150, 255},
	ColorToggleHover:           color.RGBA{120, 120, 120, 255},
	ColorToggleCursor:          color.RGBA{175, 175, 175, 255},
	ColorSelect:                color.RGBA{175, 175, 175, 255},
	ColorSelectActive:          color.RGBA{190, 190, 190, 255},
	ColorSlider:                color.RGBA{0xCE, 0xCE, 0xCE, 255}, // modified
	ColorSliderCursor:          color.RGBA{0x25, 0x59, 0xA9, 255}, // modified
	ColorSliderCursorHover:     color.RGBA{70, 70, 70, 255},
	ColorSliderCursorActive:    color.RGBA{60, 60, 60, 255},
	ColorProperty:              color.RGBA{175, 175, 175, 255},
	ColorEdit:                  color.RGBA{150, 150, 150, 255},
	ColorEditCursor:            color.RGBA{0, 0, 0, 255},
	ColorCombo:                 color.RGBA{175, 175, 175, 255},
	ColorChart:                 color.RGBA{160, 160, 160, 255},
	ColorChartColor:            color.RGBA{45, 45, 45, 255},
	ColorChartColorHighlight:   color.RGBA{255, 0, 0, 255},
	ColorScrollbar:             color.RGBA{180, 180, 180, 255},
	ColorScrollbarCursor:       color.RGBA{140, 140, 140, 255},
	ColorScrollbarCursorHover:  color.RGBA{150, 150, 150, 255},
	ColorScrollbarCursorActive: color.RGBA{160, 160, 160, 255},
	ColorTabHeader:             color.RGBA{180, 180, 180, 255},
}

func New() *GUI {
	gui := &GUI{}
	gui.ready = make(chan (interface{}))
	gui.title.Store("")
	gui.text.Store("")
	gui.progressMax.Store(0)
	return gui
}

func (gui *GUI) Start(windowTitle string) error {
	if gui == nil {
		return nil
	}
	imageBytes, err := Asset("assets/Icon64.png")
	if err != nil {
		return err
	}
	reader := bytes.NewReader(imageBytes)
	img, err := utils.LoadPngImage(reader)
	if err != nil {
		return err
	}
	gui.icon = img
	go func() {
		gui.WaitForWindow()
		if err := utils.LoadIconAndSetForWindow(windowTitle); err != nil {
			log.Printf("warning: unable to set window icon: %v", err)
		}
	}()
	window := nucular.NewMasterWindowSize(0, windowTitle, image.Point{470, 240}, gui.updateFn)
	window.SetStyle(gui.makeStyle())
	gui.window = window
	window.Main()
	return nil
}

func (gui *GUI) makeStyle() *style.Style {
	style := style.FromTable(myThemeTable, scaling)
	style.Button.TextActive = myThemeTable.ColorBorder
	style.Button.TextNormal = myThemeTable.ColorBorder
	style.Button.Rounding = 0
	style.Button.Border = 1
	style.Button.Padding = image.Point{4, 4}
	style.Progress.Rounding = 0
	style.Progress.Padding = image.Point{0, 0}
	style.NormalWindow.Padding = image.Point{20, 0}
	myFont, err := utils.LoadFont("Arial", 11, scaling)
	if err != nil {
		log.Printf("warning: %v\n", err)
	}
	style.Font = myFont
	return style
}

func (gui *GUI) updateFn(w *nucular.Window) {
	gui.emitWindowReady()
	if w.Input().Keyboard.Pressed(key.CodeEscape) {
		log.Println("escape pressed, closing window...")
		gui.SendTextMessage("Cancelling...")
		go gui.cancel(w)
	}
	centralPartWidth := 420
	iconWidth := 80
	textWidth := centralPartWidth - iconWidth

	w.Row(20).Dynamic(1)
	w.Spacing(1)

	w.Row(64).Static(iconWidth, textWidth)
	w.Image(gui.icon)
	w.Label(gui.title.Load().(string), "CC")

	if gui.err != nil {
		w.Row(85).Dynamic(1)
		re := regexp.MustCompile("[^!-~\t ]")
		text := re.ReplaceAllLiteralString(gui.err.Error(), "")
		w.LabelWrap(text)

		w.Row(30).Dynamic(5)
		w.Spacing(4)
		if w.Button(label.TA("Close", "CC"), false) {
			log.Println("close button pressed, closing window...")
			go gui.cancel(w)
		}
		return
	}
	w.Row(30).Dynamic(1)
	w.Spacing(1)

	w.Row(20).Dynamic(1)
	w.Label(gui.text.Load().(string), "LC")

	w.Row(12).Dynamic(1)
	progress := int(atomic.LoadInt32(&gui.progress))
	progressMax := gui.progressMax.Load().(int)
	w.Progress(&progress, progressMax, false)

	w.Row(10).Dynamic(1)
	w.Spacing(1)

	w.Row(30).Dynamic(5)
	w.Spacing(4)
	if progress < progressMax {
		if w.Button(label.TA("Cancel", "CC"), false) {
			log.Println("cancel button pressed, closing window...")
			go gui.cancel(w)
		}
	} else {
		if w.Button(label.TA("Close", "CC"), false) {
			log.Println("close button pressed, closing window...")
			go gui.cancel(w)
		}
	}
}

func (gui *GUI) emitWindowReady() {
	gui.readyOnce.Do(func() { close(gui.ready) })
}

func (gui *GUI) WaitForWindow() {
	if gui == nil {
		return
	}
	<-gui.ready
}

func (gui *GUI) cancel(w *nucular.Window) {
	w.Master().Close()
}

func (gui *GUI) Terminate() error {
	if gui == nil {
		return nil
	}
	if gui.window != nil {
		if !gui.window.Closed() {
			go gui.window.Close()
		}
	}
	return nil
}

func (gui *GUI) SendTextMessage(text string) error {
	if gui == nil {
		return nil
	}
	gui.text.Store(text)
	gui.window.Changed()
	return nil
}

func (gui *GUI) SendErrorMessage(err error) error {
	if gui == nil {
		return nil
	}
	gui.err = err
	gui.window.Changed()
	return nil
}

func (gui *GUI) SendCloseMessage() error {
	if gui == nil {
		return nil
	}
	gui.window.Close()
	return nil
}

func (gui *GUI) SetTitle(title string) error {
	if gui == nil {
		return nil
	}
	gui.title.Store(title)
	gui.window.Changed()
	return nil
}

func (gui *GUI) SetProgressMax(val int) {
	if gui == nil {
		return
	}
	gui.progressMax.Store(val)
}

func (gui *GUI) ProgressStep() {
	if gui == nil {
		return
	}
	atomic.AddInt32(&gui.progress, 1)
	gui.window.Changed()
}

func (gui *GUI) Closed() bool {
	if gui == nil {
		return false
	}
	return gui.window.Closed()
}
