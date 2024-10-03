package main

import (
	"com.nodian.app/hash"
	"com.nodian.app/json"
	"com.nodian.app/markdown"
	"com.nodian.app/timestamp"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type mainApp struct {
	window             fyne.Window
	markdownEditor     *markdown.MarkdownEditor
	jsonFormatter      *json.JSONFormatter
	timestampConverter *timestamp.TimestampConverter
	hashTool           *hash.HashTool
	content            *fyne.Container
}

func newMainApp(a fyne.App) *mainApp {
	win := a.NewWindow("Nodian")
	m := &mainApp{window: win}
	m.makeUI()
	return m
}

func (m *mainApp) makeUI() {
	// 初始化 Markdown 编辑器
	m.markdownEditor = markdown.NewMarkdownEditor(m.window)
	err := m.markdownEditor.LoadDirectory(".") // 使用当前目录
	if err != nil {
		fyne.LogError("Failed to load directory", err)
	}

	// 初始化 JSON 格式化器
	m.jsonFormatter = json.NewJSONFormatter(m.window)

	// 初始化时间戳转换器
	m.timestampConverter = timestamp.NewTimestampConverter(m.window)

	// 初始化哈希工具
	m.hashTool = hash.NewHashTool(m.window)

	// 创建左侧菜单
	menu := widget.NewList(
		func() int { return 4 },
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
			case 2:
				icon.SetResource(theme.ComputerIcon())
			case 3:
				icon.SetResource(theme.HelpIcon())
			}
		},
	)

	menu.OnSelected = func(id widget.ListItemID) {
		switch id {
		case 0:
			m.content.Objects[0] = m.markdownEditor.Container()
		case 1:
			m.content.Objects[0] = m.jsonFormatter.CreateUI()
		case 2:
			m.content.Objects[0] = m.timestampConverter.CreateUI()
		case 3:
			m.content.Objects[0] = m.hashTool.CreateUI()
		}
		m.content.Refresh()
	}

	// 创建主内容区域
	m.content = container.NewStack(m.markdownEditor.Container())

	// 使用一个容器来固定菜单宽度
	menuContainer := container.New(&fixedWidthLayout{width: 40}, menu)

	// 使用新的布局替换之前的 split
	mainContainer := container.NewBorder(nil, nil, menuContainer, nil, m.content)

	m.window.SetContent(mainContainer)
}

// 创建一个自定义布局来固定宽度
type fixedWidthLayout struct {
	width float32
}

func (f *fixedWidthLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) == 0 {
		return
	}
	objects[0].Resize(fyne.NewSize(f.width, size.Height))
	objects[0].Move(fyne.NewPos(0, 0))
}

func (f *fixedWidthLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(f.width, 0)
	}
	minSize := objects[0].MinSize()
	return fyne.NewSize(f.width, minSize.Height)
}

func main() {
	a := app.New()
	mainApp := newMainApp(a)
	mainApp.window.Resize(fyne.NewSize(800, 600))
	mainApp.window.ShowAndRun()
}
