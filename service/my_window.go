package service

import (
	"log"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
)

type MyWindow struct {
	*walk.MainWindow
	hWnd    win.HWND
	listBox *walk.ListBox
	ni      *walk.NotifyIcon
	model   *logModel
}

func (mw *MyWindow) AddNotifyIcon() {
	var err error
	mw.ni, err = walk.NewNotifyIcon(mw)
	if err != nil {
		log.Fatal(err)
	}

	icon, err := walk.Resources.Image("img/show.ico")
	if err != nil {
		log.Fatal(err)
	}
	mw.SetIcon(icon)
	mw.ni.SetIcon(icon)
	mw.ni.SetVisible(true)

	mw.ni.MouseDown().Attach(func(x, y int, button walk.MouseButton) {
		if button == walk.LeftButton {
			mw.Show()
			win.ShowWindow(mw.Handle(), win.SW_RESTORE)
		}
	})

	//实现托盘右键退出功能
	exitAction := walk.NewAction()
	if err := exitAction.SetText("退出"); err != nil {
		log.Fatal(err)
	}
	//Exit 实现的功能
	exitAction.Triggered().Attach(func() { walk.App().Exit(0) })
	if err := mw.ni.ContextMenu().Actions().Add(exitAction); err != nil {
		log.Fatal(err)
	}

}

func (mw *MyWindow) AddLog(content string) {
	mw.Synchronize(func() {
		model := mw.model
		lb := mw.listBox
		trackLatest := lb.ItemVisible(len(model.items)-1) && len(lb.SelectedIndexes()) <= 1
		model.items = append(model.items, logEntry{time.Now(), content})
		index := len(model.items) - 1
		model.PublishItemsInserted(index, index)

		if trackLatest {
			lb.EnsureItemVisible(len(model.items) - 1)
		}
	})
}

func NewWindow(title string) (mw *MyWindow) {
	mw = new(MyWindow)
	var items []logEntry
	items = append(items, logEntry{time.Now(), "app run"})
	model := &logModel{items: items}
	mw.model = model
	styler := &Styler{
		lb:                  &mw.listBox,
		model:               model,
		dpi2StampSize:       make(map[int]walk.Size),
		widthDPI2WsPerLine:  make(map[widthDPI]int),
		textWidthDPI2Height: make(map[textWidthDPI]int),
	}
	if err := (MainWindow{
		AssignTo:   &mw.MainWindow,
		Title:      title,
		Background: SolidColorBrush{walk.RGB(100, 100, 100)},
		MinSize:    Size{200, 200},
		Size:       Size{500, 500},
		Font:       Font{Family: "Segoe UI", PointSize: 9},
		Layout:     VBox{},
		OnSizeChanged: func() {
			if win.IsIconic(mw.Handle()) {
				mw.Hide()
				mw.ni.SetVisible(true)
			}
		},
		Children: []Widget{
			Composite{
				DoubleBuffering: true,
				Layout:          VBox{},
				Children: []Widget{
					ListBox{
						AssignTo:       &mw.listBox,
						Background:     SolidColorBrush{walk.RGB(100, 100, 100)},
						MultiSelection: true,
						Model:          model,
						ItemStyler:     styler,
					},
				},
			},
		},
	}).Create(); err != nil {
		log.Fatal(err)
	}

	mw.hWnd = mw.Handle()
	mw.AddNotifyIcon()
	return
}

type logModel struct {
	walk.ReflectListModelBase
	items []logEntry
}

func (m *logModel) Items() interface{} {
	return m.items
}

type logEntry struct {
	timestamp time.Time
	message   string
}

type widthDPI struct {
	width int // in native pixels
	dpi   int
}

type textWidthDPI struct {
	text  string
	width int // in native pixels
	dpi   int
}

type Styler struct {
	lb                  **walk.ListBox
	canvas              *walk.Canvas
	model               *logModel
	font                *walk.Font
	dpi2StampSize       map[int]walk.Size
	widthDPI2WsPerLine  map[widthDPI]int
	textWidthDPI2Height map[textWidthDPI]int // in native pixels
}

func (s *Styler) ItemHeightDependsOnWidth() bool {
	return true
}

func (s *Styler) DefaultItemHeight() int {
	dpi := (*s.lb).DPI()
	marginV := walk.IntFrom96DPI(marginV96dpi, dpi)

	return s.StampSize().Height + marginV*2
}

const (
	marginH96dpi int = 6
	marginV96dpi int = 2
	lineW96dpi   int = 1
)

func (s *Styler) ItemHeight(index, width int) int {
	dpi := (*s.lb).DPI()
	marginH := walk.IntFrom96DPI(marginH96dpi, dpi)
	marginV := walk.IntFrom96DPI(marginV96dpi, dpi)
	lineW := walk.IntFrom96DPI(lineW96dpi, dpi)

	msg := s.model.items[index].message

	twd := textWidthDPI{msg, width, dpi}

	if height, ok := s.textWidthDPI2Height[twd]; ok {
		return height + marginV*2
	}

	canvas, err := s.Canvas()
	if err != nil {
		return 0
	}

	stampSize := s.StampSize()

	wd := widthDPI{width, dpi}
	wsPerLine, ok := s.widthDPI2WsPerLine[wd]
	if !ok {
		bounds, _, err := canvas.MeasureTextPixels("W", (*s.lb).Font(), walk.Rectangle{Width: 9999999}, walk.TextCalcRect)
		if err != nil {
			return 0
		}
		wsPerLine = (width - marginH*4 - lineW - stampSize.Width) / bounds.Width
		s.widthDPI2WsPerLine[wd] = wsPerLine
	}

	if len(msg) <= wsPerLine {
		s.textWidthDPI2Height[twd] = stampSize.Height
		return stampSize.Height + marginV*2
	}

	bounds, _, err := canvas.MeasureTextPixels(msg, (*s.lb).Font(), walk.Rectangle{Width: width - marginH*4 - lineW - stampSize.Width, Height: 255}, walk.TextEditControl|walk.TextWordbreak|walk.TextEndEllipsis)
	if err != nil {
		return 0
	}

	s.textWidthDPI2Height[twd] = bounds.Height

	return bounds.Height + marginV*2
}

func (s *Styler) StyleItem(style *walk.ListItemStyle) {
	if canvas := style.Canvas(); canvas != nil {
		if style.Index()%2 == 1 && style.BackgroundColor == walk.Color(win.GetSysColor(win.COLOR_WINDOW)) {
			style.BackgroundColor = walk.Color(win.GetSysColor(win.COLOR_BTNFACE))
			if err := style.DrawBackground(); err != nil {
				return
			}
		}

		pen, err := walk.NewCosmeticPen(walk.PenSolid, style.LineColor)
		if err != nil {
			return
		}
		defer pen.Dispose()

		dpi := (*s.lb).DPI()
		marginH := walk.IntFrom96DPI(marginH96dpi, dpi)
		marginV := walk.IntFrom96DPI(marginV96dpi, dpi)
		lineW := walk.IntFrom96DPI(lineW96dpi, dpi)

		b := style.BoundsPixels()
		b.X += marginH
		b.Y += marginV

		item := s.model.items[style.Index()]

		style.DrawText(item.timestamp.Format(time.Stamp), b, walk.TextEditControl|walk.TextWordbreak)

		stampSize := s.StampSize()

		x := b.X + stampSize.Width + marginH + lineW
		canvas.DrawLinePixels(pen, walk.Point{x, b.Y - marginV}, walk.Point{x, b.Y - marginV + b.Height})

		b.X += stampSize.Width + marginH*2 + lineW
		b.Width -= stampSize.Width + marginH*4 + lineW

		style.DrawText(item.message, b, walk.TextEditControl|walk.TextWordbreak|walk.TextEndEllipsis)
	}
}

func (s *Styler) StampSize() walk.Size {
	dpi := (*s.lb).DPI()

	stampSize, ok := s.dpi2StampSize[dpi]
	if !ok {
		canvas, err := s.Canvas()
		if err != nil {
			return walk.Size{}
		}

		bounds, _, err := canvas.MeasureTextPixels("Jan _2 20:04:05 ", (*s.lb).Font(), walk.Rectangle{Width: 9999999}, walk.TextCalcRect)
		if err != nil {
			return walk.Size{}
		}

		stampSize = bounds.Size()
		s.dpi2StampSize[dpi] = stampSize
	}

	return stampSize
}

func (s *Styler) Canvas() (*walk.Canvas, error) {
	if s.canvas != nil {
		return s.canvas, nil
	}

	canvas, err := (*s.lb).CreateCanvas()
	if err != nil {
		return nil, err
	}
	s.canvas = canvas
	(*s.lb).AddDisposable(canvas)

	return canvas, nil
}
