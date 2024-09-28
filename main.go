package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"com.nodian.app/markdown"
)

func main() {
	a := app.New()
	w := a.NewWindow("Nodian")
	w.Resize(fyne.NewSize(1200, 800))

	// 创建左侧菜单
	menuItems := []struct {
		icon fyne.Resource
		name string
	}{
		{theme.DocumentIcon(), "Markdown"},
		{theme.FileIcon(), "日程"},
		{theme.DocumentIcon(), "JSON"},
		{theme.HomeIcon(), "时间转换"},
		{theme.InfoIcon(), "HASH"},
		{theme.ContentCopyIcon(), "剪贴板"},
	}

	var menuButtons []fyne.CanvasObject
	for _, item := range menuItems {
		button := widget.NewButtonWithIcon("", item.icon, nil)
		button.Importance = widget.LowImportance
		menuButtons = append(menuButtons, button)
	}

	menu := container.NewVBox(menuButtons...)

	// 创建一个固定宽度的容器来包裹菜单
	fixedWidthMenu := container.NewBorder(nil, nil, nil, nil, menu)
	fixedWidthMenu.Resize(fyne.NewSize(40, 0))

	// 创建 Markdown 编辑器
	markdownEditor := markdown.NewMarkdownEditor(w)
	err := markdownEditor.LoadDirectory(".") // 使用当前目录
	if err != nil {
		fyne.LogError("Failed to load directory", err)
	}

	// 创建内容区
	content := markdownEditor.Container()

	// 创建主布局
	split := container.NewBorder(nil, nil, fixedWidthMenu, nil, content)

	w.SetContent(split)
	w.ShowAndRun()
}
