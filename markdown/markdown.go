package markdown

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type MarkdownEditor struct {
	container    *fyne.Container
	treeView     *widget.Tree
	contentSplit *container.Split
	tabs         *container.DocTabs
	rootPath     string
	window       fyne.Window
	selectedNode widget.TreeNodeID
	openFiles    map[string]*widget.Entry // 新增：用于跟踪打开的文件
}

func NewMarkdownEditor(window fyne.Window) *MarkdownEditor {
	m := &MarkdownEditor{
		window:    window,
		openFiles: make(map[string]*widget.Entry), // 初始化 openFiles
	}
	m.initUI()
	return m
}

func (m *MarkdownEditor) initUI() {
	// 创建目录树
	m.treeView = widget.NewTree(
		m.childUIDs,
		m.isBranch,
		m.createNode,
		m.updateNode,
	)

	m.treeView.OnSelected = m.onNodeSelected

	// 创建顶部工具栏
	toolbar := container.NewHBox(
		widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), func() { m.newFile(m.selectedNode) }),
		widget.NewButtonWithIcon("", theme.FolderNewIcon(), func() { m.newFolder(m.selectedNode) }),
		widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), m.refreshTree),
		widget.NewButtonWithIcon("", theme.VisibilityIcon(), m.toggleTreeExpansion),
	)

	// 创建文件标签和内容区
	m.tabs = container.NewDocTabs()
	m.contentSplit = container.NewHSplit(
		container.NewBorder(toolbar, nil, nil, nil, m.treeView),
		m.tabs,
	)
	m.contentSplit.Offset = 0.2 // 将目录树的宽度设置为内容区域的 20%

	m.container = container.NewStack(m.contentSplit)

	// 修改快捷键支持
	ctrlS := &desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: fyne.KeyModifierControl}
	superS := &desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: fyne.KeyModifierSuper}

	m.window.Canvas().AddShortcut(ctrlS, func(shortcut fyne.Shortcut) {
		m.saveCurrentFile()
	})
	m.window.Canvas().AddShortcut(superS, func(shortcut fyne.Shortcut) {
		m.saveCurrentFile()
	})

	// 添加右键菜单
	// m.treeView.OnTapped = func(e *fyne.PointEvent) {
	// 	if e.Position.X < 0 {
	// 		return // 忽略拖动事件
	// 	}
	// 	if e.Button == fyne.RightMouseButton {
	// 		menu := m.createContextMenu(m.treeView.SelectedID())
	// 		widget.ShowPopUpMenuAtPosition(menu, fyne.CurrentApp().Driver().CanvasForObject(m.treeView), e.Position)
	// 	}
	// }
}

func (m *MarkdownEditor) LoadDirectory(path string) error {
	// 确保 nodian 目录存在
	m.rootPath = filepath.Join(path, "nodian")
	if _, err := os.Stat(m.rootPath); os.IsNotExist(err) {
		err = os.Mkdir(m.rootPath, 0755)
		if err != nil {
			return err
		}
	}

	m.treeView.Root = "" // 将根设置为空字符串
	m.treeView.OpenAllBranches()
	m.treeView.Refresh()

	return nil
}

func (m *MarkdownEditor) Container() fyne.CanvasObject {
	return m.container
}

func (m *MarkdownEditor) childUIDs(uid widget.TreeNodeID) []widget.TreeNodeID {
	path := m.uidToPath(uid)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		fyne.LogError("Failed to read directory", err)
		return nil
	}

	var children []widget.TreeNodeID
	for _, file := range files {
		childUID := file.Name()
		if uid != "" {
			childUID = filepath.Join(uid, file.Name())
		}
		children = append(children, childUID)
	}

	sort.Slice(children, func(i, j int) bool {
		iPath := m.uidToPath(children[i])
		jPath := m.uidToPath(children[j])
		iInfo, _ := os.Stat(iPath)
		jInfo, _ := os.Stat(jPath)
		if iInfo.IsDir() && !jInfo.IsDir() {
			return true
		}
		if !iInfo.IsDir() && jInfo.IsDir() {
			return false
		}
		return strings.ToLower(iInfo.Name()) < strings.ToLower(jInfo.Name())
	})

	return children
}

func (m *MarkdownEditor) isBranch(uid widget.TreeNodeID) bool {
	path := m.uidToPath(uid)
	info, err := os.Stat(path)
	if err != nil {
		fyne.LogError("Failed to get file info", err)
		return false
	}
	return info.IsDir()
}

func (m *MarkdownEditor) createNode(branch bool) fyne.CanvasObject {
	return widget.NewLabel("Template Object")
}

func (m *MarkdownEditor) updateNode(uid widget.TreeNodeID, branch bool, node fyne.CanvasObject) {
	label := node.(*widget.Label)
	path := m.uidToPath(uid)
	label.SetText(filepath.Base(path))
}

func (m *MarkdownEditor) onNodeSelected(uid widget.TreeNodeID) {
	m.selectedNode = uid
	path := m.uidToPath(uid)
	info, err := os.Stat(path)
	if err != nil {
		fyne.LogError("Failed to get file info", err)
		return
	}

	if !info.IsDir() {
		m.openFile(path)
	}
}

func (m *MarkdownEditor) openFile(path string) {
	// 检查文件是否已经打开
	for _, tab := range m.tabs.Items {
		if tab.Text == filepath.Base(path) {
			m.tabs.Select(tab)
			return
		}
	}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		dialog.ShowError(err, m.window)
		return
	}

	editor := widget.NewMultiLineEntry()
	editor.SetText(string(content))

	preview := NewCustomRichText() // 使用自定义的 RichText
	preview.Wrapping = fyne.TextWrapWord

	split := container.NewHSplit(editor, container.NewScroll(preview))
	split.Offset = 0.5

	tab := container.NewTabItem(filepath.Base(path), split)
	m.tabs.Append(tab)
	m.tabs.Select(tab)

	m.openFiles[path] = editor // 将打开的文件添加到 map 中

	// 立即更新预览
	m.updatePreview(preview, editor.Text)

	// 强制重新布局整个分割视图
	split.Refresh()

	editor.OnChanged = func(content string) {
		if !strings.HasPrefix(tab.Text, "*") {
			tab.Text = "*" + tab.Text
			m.tabs.Refresh()
		}
		m.updatePreview(preview, content)
	}

	// 监听标签页关闭事件
	m.tabs.OnClosed = func(item *container.TabItem) {
		for p, e := range m.openFiles {
			if e == editor {
				delete(m.openFiles, p)
				break
			}
		}
	}
}

// CustomRichText 是一个自定义的 RichText 组件
type CustomRichText struct {
	widget.RichText
	lineSpacing float32
}

// 创建新的 CustomRichText
func NewCustomRichText() *CustomRichText {
	rt := &CustomRichText{}
	rt.ExtendBaseWidget(rt)
	rt.lineSpacing = 5.5 // 设置行间距为 1.5 倍
	return rt
}

// MinSize 重写 MinSize 方法以考虑行间距
func (rt *CustomRichText) MinSize() fyne.Size {
	size := rt.RichText.MinSize()
	size.Height = float32(float64(size.Height) * float64(rt.lineSpacing))
	return size
}

// CreateRenderer 重写 CreateRenderer 方法以自定义渲染
func (rt *CustomRichText) CreateRenderer() fyne.WidgetRenderer {
	return &customRichTextRenderer{
		richText:     rt,
		baseRenderer: rt.RichText.CreateRenderer(),
	}
}

// customRichTextRenderer 是 CustomRichText 的自定义渲染器
type customRichTextRenderer struct {
	richText     *CustomRichText
	baseRenderer fyne.WidgetRenderer
}

func (r *customRichTextRenderer) Destroy() {
	r.baseRenderer.Destroy()
}

func (r *customRichTextRenderer) Layout(size fyne.Size) {
	r.baseRenderer.Layout(size)
	r.applyLineSpacing()
}

func (r *customRichTextRenderer) MinSize() fyne.Size {
	baseSize := r.baseRenderer.MinSize()
	return fyne.NewSize(baseSize.Width, baseSize.Height*r.richText.lineSpacing)
}

func (r *customRichTextRenderer) Objects() []fyne.CanvasObject {
	return r.baseRenderer.Objects()
}

func (r *customRichTextRenderer) Refresh() {
	r.baseRenderer.Refresh()
	r.applyLineSpacing()
}

func (r *customRichTextRenderer) applyLineSpacing() {
	y := float32(0)
	for _, o := range r.Objects() {
		if text, ok := o.(*canvas.Text); ok {
			text.Move(fyne.NewPos(text.Position().X, y))
			y += text.MinSize().Height * r.richText.lineSpacing
		}
	}
}

func (m *MarkdownEditor) updatePreview(preview *CustomRichText, content string) {
	preview.ParseMarkdown(content)
	preview.Refresh()

	// 强制重新布局
	if split, ok := m.tabs.Selected().Content.(*container.Split); ok {
		if scroll, ok := split.Trailing.(*container.Scroll); ok {
			scroll.Refresh()
		}
	}

	// 添加日志以检查 Markdown 内容
	fyne.LogError("Markdown content", errors.New(content))
}

func (m *MarkdownEditor) saveCurrentFile() {
	currentTab := m.tabs.Selected()
	if currentTab == nil {
		return // 没有选中的标签页
	}

	var editor *widget.Entry
	var path string

	// 查找当前正在编辑的文件
	if split, ok := currentTab.Content.(*container.Split); ok {
		if entry, ok := split.Leading.(*widget.Entry); ok {
			editor = entry
			for p, e := range m.openFiles {
				if e == entry {
					path = p
					break
				}
			}
		}
	}

	if editor == nil || path == "" {
		dialog.ShowError(errors.New("无法找到当前编辑的文件"), m.window)
		return
	}

	// 获取当前编辑器中的文本内容
	content := editor.Text

	// 保存文件
	err := ioutil.WriteFile(path, []byte(content), 0644)
	if err != nil {
		dialog.ShowError(err, m.window)
		return
	}

	// 更新标签页标题（移除星号）
	currentTab.Text = strings.TrimPrefix(currentTab.Text, "*")
	m.tabs.Refresh()

	dialog.ShowInformation("保存成功", "文件已成功保存", m.window)
}

func (m *MarkdownEditor) refreshTree() {
	m.treeView.Refresh()
}

func (m *MarkdownEditor) toggleTreeExpansion() {
	if m.treeView.IsBranchOpen(m.treeView.Root) {
		m.treeView.CloseAllBranches()
	} else {
		m.treeView.OpenAllBranches()
	}
}

func (m *MarkdownEditor) createContextMenu(uid widget.TreeNodeID) *fyne.Menu {
	return fyne.NewMenu("",
		fyne.NewMenuItem("New File", func() { m.newFile(uid) }),
		fyne.NewMenuItem("New Folder", func() { m.newFolder(uid) }),
		fyne.NewMenuItem("Rename", func() { m.rename(uid) }),
		fyne.NewMenuItem("Delete", func() { m.delete(uid) }),
	)
}

func (m *MarkdownEditor) newFile(uid widget.TreeNodeID) {
	selectedPath := m.uidToPath(uid)
	if uid == "" {
		selectedPath = m.rootPath
	} else if !m.isBranch(uid) {
		selectedPath = filepath.Dir(selectedPath)
	}

	entry := widget.NewEntry()
	entry.SetPlaceHolder("Enter file name")
	dialog.ShowForm("New File", "Create", "Cancel", []*widget.FormItem{
		widget.NewFormItem("Name", entry),
	}, func(ok bool) {
		if ok {
			fileName := entry.Text
			if !strings.HasSuffix(fileName, ".md") {
				fileName += ".md"
			}
			newPath := filepath.Join(selectedPath, fileName)
			err := ioutil.WriteFile(newPath, []byte(""), 0644)
			if err != nil {
				dialog.ShowError(err, m.window)
				return
			}
			m.treeView.Refresh()
			m.openFile(newPath)
		}
	}, m.window)
}

func (m *MarkdownEditor) newFolder(uid widget.TreeNodeID) {
	selectedPath := m.uidToPath(uid)
	if uid == "" {
		selectedPath = m.rootPath
	} else if !m.isBranch(uid) {
		selectedPath = filepath.Dir(selectedPath)
	}

	entry := widget.NewEntry()
	entry.SetPlaceHolder("Enter folder name")
	dialog.ShowForm("New Folder", "Create", "Cancel", []*widget.FormItem{
		widget.NewFormItem("Name", entry),
	}, func(ok bool) {
		if ok {
			newPath := filepath.Join(selectedPath, entry.Text)
			err := os.Mkdir(newPath, 0755)
			if err != nil {
				dialog.ShowError(err, m.window)
				return
			}
			m.treeView.Refresh()
		}
	}, m.window)
}

func (m *MarkdownEditor) rename(uid widget.TreeNodeID) {
	oldPath := m.uidToPath(uid)
	entry := widget.NewEntry()
	entry.SetText(filepath.Base(oldPath))
	dialog.ShowForm("Rename", "Rename", "Cancel", []*widget.FormItem{
		widget.NewFormItem("New Name", entry),
	}, func(ok bool) {
		if ok {
			newPath := filepath.Join(filepath.Dir(oldPath), entry.Text)
			err := os.Rename(oldPath, newPath)
			if err != nil {
				dialog.ShowError(err, m.window)
				return
			}
			m.treeView.Refresh()
		}
	}, m.window)
}

func (m *MarkdownEditor) delete(uid widget.TreeNodeID) {
	path := m.uidToPath(uid)
	dialog.ShowConfirm("Delete", "Are you sure you want to delete this item?", func(ok bool) {
		if ok {
			err := os.RemoveAll(path)
			if err != nil {
				dialog.ShowError(err, m.window)
				return
			}
			m.treeView.Refresh()
		}
	}, m.window)
}

func (m *MarkdownEditor) uidToPath(uid widget.TreeNodeID) string {
	return filepath.Join(m.rootPath, uid)
}
