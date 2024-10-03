package main

import (
	"com.nodian.app/json"
	"com.nodian.app/markdown"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type mainApp struct {
	window         fyne.Window
	markdownEditor *markdown.MarkdownEditor
	jsonFormatter  *json.JSONFormatter
	content        *fyne.Container
}

func newMainApp(a fyne.App) *mainApp {
	win := a.NewWindow("Multi-Tool App")
	m := &mainApp{window: win}
	m.makeUI()
	return m
}

func (m *mainApp) makeUI() {
	m.markdownEditor = markdown.NewMarkdownEditor(m.window)
	m.jsonFormatter = json.NewJSONFormatter(m.window)

	// 创建左侧菜单
	menu := widget.NewList(
		func() int { return 2 },
		func() fyne.CanvasObject {
			return widget.NewIcon(theme.DocumentIcon())
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			icon := item.(*widget.Icon)
			switch id {
			case 0:
				icon.SetResource(theme.DocumentIcon())
			case 1:
				icon.SetResource(theme.ListIcon())
			}
		},
	)

	menu.OnSelected = func(id widget.ListItemID) {
		switch id {
		case 0:
			m.content.Objects[0] = m.markdownEditor.Container()
		case 1:
			m.content.Objects[0] = m.jsonFormatter.CreateUI()
		}
		m.content.Refresh()
	}

	// 创建主内容区域
	m.content = container.NewMax(m.markdownEditor.Container())

	// 创建主布局
	split := container.NewHSplit(menu, m.content)
	split.Offset = 0.1 // 设置左侧菜单宽度为10%

	m.window.SetContent(split)
}

func main() {
	a := app.New()
	mainApp := newMainApp(a)
	mainApp.window.Resize(fyne.NewSize(800, 600))
	mainApp.window.ShowAndRun()
}
