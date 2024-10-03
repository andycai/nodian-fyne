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
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type MarkdownEditor struct {
	container     *fyne.Container
	treeView      *widget.Tree
	contentSplit  *container.Split
	tabs          *container.DocTabs
	rootPath      string
	window        fyne.Window
	selectedNode  widget.TreeNodeID
	openFiles     map[string]*widget.Entry // 新增：用于跟踪打开的文件
	isCreatingNew bool
	newItemEntry  *widget.Entry
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

	// 部工具栏
	toolbar := container.NewHBox(
		widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), func() { m.startCreatingNew(false) }),
		widget.NewButtonWithIcon("", theme.FolderNewIcon(), func() { m.startCreatingNew(true) }),
		widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), m.refreshTree),
		widget.NewButtonWithIcon("", theme.VisibilityIcon(), m.toggleTreeExpansion),
		widget.NewButtonWithIcon("", theme.DocumentSaveIcon(), m.saveCurrentFile),
		widget.NewButtonWithIcon("", theme.ContentCutIcon(), m.renameSelected), // 新增重命名按钮
		widget.NewButtonWithIcon("", theme.DeleteIcon(), m.deleteSelected),     // 新增���除按钮
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
	// 		return // 忽略拖事件
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
		if os.IsNotExist(err) {
			// 如果文件或目录不存在，我们假设它不是分支
			return false
		}
		fyne.LogError("Failed to get file info", err)
		return false
	}
	return info.IsDir()
}

func (m *MarkdownEditor) createNode(branch bool) fyne.CanvasObject {
	var icon fyne.Resource
	if branch {
		icon = theme.FolderIcon()
	} else {
		icon = theme.DocumentIcon()
	}
	return container.New(layout.NewHBoxLayout(),
		widget.NewIcon(icon),
		widget.NewLabel(""))
}

func (m *MarkdownEditor) updateNode(uid widget.TreeNodeID, branch bool, node fyne.CanvasObject) {
	if m.isCreatingNew && strings.HasSuffix(string(uid), "new_item") {
		// 如果是正在创建新项目，不做任何修改
		return
	}

	container, ok := node.(*fyne.Container)
	if !ok {
		// 如果不是容器，创建一个新的容器
		container = fyne.NewContainerWithLayout(layout.NewHBoxLayout(),
			widget.NewIcon(theme.DocumentIcon()),
			widget.NewLabel(""))
		m.treeView.UpdateNode(uid, branch, container)
		return
	}

	icon := container.Objects[0].(*widget.Icon)
	label, isLabel := container.Objects[1].(*widget.Label)

	if branch {
		icon.SetResource(theme.FolderIcon())
	} else {
		icon.SetResource(theme.DocumentIcon())
	}

	if isLabel {
		label.SetText(filepath.Base(m.uidToPath(uid)))
	}
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

	// 监听标签页闭事件
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
	// fyne.LogError("Markdown content", errors.New(content))
}

func (m *MarkdownEditor) saveCurrentFile() {
	currentTab := m.tabs.Selected()
	if currentTab == nil {
		return // 没有选中的标签页
	}

	var editor *widget.Entry
	var path string

	// 查找当前正在编辑文件
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

	// 更新标签页标题（移星号）
	currentTab.Text = strings.TrimPrefix(currentTab.Text, "*")
	m.tabs.Refresh()

	// 移除成功保存的弹窗
	// dialog.ShowInformation("保存成功", "文件已成功保存", m.window)
}

func (m *MarkdownEditor) refreshTree() {
	m.treeView.Refresh()
}

func (m *MarkdownEditor) toggleTreeExpansion() {
	if m.treeView.IsBranchOpen(m.treeView.Root) {
		m.treeView.CloseAllBranches()
	} else {
		m.openAllBranches(m.treeView.Root)
	}
	m.treeView.Refresh()
}

func (m *MarkdownEditor) openAllBranches(uid widget.TreeNodeID) {
	m.treeView.OpenBranch(uid)
	for _, childUID := range m.treeView.ChildUIDs(uid) {
		if m.treeView.IsBranch(childUID) {
			m.openAllBranches(childUID)
		}
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

func (m *MarkdownEditor) startCreatingNew(isFolder bool) {
	parentNode := m.selectedNode
	if parentNode == "" || !m.isBranch(parentNode) {
		parentNode = m.treeView.Root
	}

	entry := widget.NewEntry()
	entry.SetPlaceHolder("Enter name...")

	var title string
	if isFolder {
		title = "New Folder"
	} else {
		title = "New File"
	}

	m.showCustomFormDialog(title, "Create", "Cancel", []*widget.FormItem{
		widget.NewFormItem("Name", entry),
	}, func(ok bool) {
		if ok {
			name := entry.Text
			m.finishCreatingNew(name, isFolder)
		}
	}, m.window)
}

func (m *MarkdownEditor) finishCreatingNew(name string, isFolder bool) {
	if name == "" {
		return
	}

	parentNode := m.selectedNode
	if parentNode == "" || !m.isBranch(parentNode) {
		parentNode = m.treeView.Root
	}

	parentPath := m.uidToPath(parentNode)
	newPath := filepath.Join(parentPath, name)
	var err error

	if isFolder {
		err = os.Mkdir(newPath, 0755)
	} else {
		if !strings.HasSuffix(name, ".md") {
			name += ".md"
			newPath += ".md"
		}
		err = os.WriteFile(newPath, []byte(""), 0644)
	}

	if err != nil {
		dialog.ShowError(err, m.window)
		return
	}

	// 更新树形视图
	m.treeView.OpenBranch(parentNode)
	m.treeView.Refresh()

	if !isFolder {
		m.openFile(newPath)
	}
}

func (m *MarkdownEditor) cancelCreatingNew() {
	m.isCreatingNew = false
	m.newItemEntry = nil
	m.window.Canvas().SetOnTypedKey(nil)
	m.treeView.Refresh()
}

func (m *MarkdownEditor) renameSelected() {
	if m.selectedNode == "" {
		dialog.ShowInformation("提示", "请先选择一个文件或文件夹", m.window)
		return
	}
	m.rename(m.selectedNode)
}

func (m *MarkdownEditor) deleteSelected() {
	if m.selectedNode == "" {
		dialog.ShowInformation("提示", "请先选择一个文件或文件夹", m.window)
		return
	}
	m.delete(m.selectedNode)
}

func (m *MarkdownEditor) showCustomFormDialog(title, confirm, dismiss string, items []*widget.FormItem, callback func(bool), window fyne.Window) {
	content := container.NewVBox()
	for _, item := range items {
		formItemContainer := container.NewBorder(nil, nil, widget.NewLabel(item.Text), nil, item.Widget)
		content.Add(formItemContainer)
	}

	// 创建按钮
	confirmButton := widget.NewButton(confirm, func() {})
	dismissButton := widget.NewButton(dismiss, func() {})

	buttons := container.NewHBox(layout.NewSpacer(), dismissButton, confirmButton)
	content.Add(buttons)

	// 创建自定义对话框，不使用默认按钮
	customDialog := dialog.NewCustomWithoutButtons(title, content, window)
	customDialog.Resize(fyne.NewSize(400, customDialog.MinSize().Height))

	// 设置按钮动作
	confirmButton.OnTapped = func() {
		customDialog.Hide()
		callback(true)
	}
	dismissButton.OnTapped = func() {
		customDialog.Hide()
		callback(false)
	}

	customDialog.Show()
}

func (m *MarkdownEditor) rename(uid widget.TreeNodeID) {
	oldPath := m.uidToPath(uid)
	entry := widget.NewEntry()
	entry.SetText(filepath.Base(oldPath))

	// 创建一个容器来包装输入框，并设置其大小
	entryContainer := container.New(layout.NewMaxLayout(), entry)
	entryContainer.Resize(fyne.NewSize(300, entry.MinSize().Height))

	m.showCustomFormDialog("Rename", "Rename", "Cancel", []*widget.FormItem{
		widget.NewFormItem("New Name", entryContainer),
	}, func(ok bool) {
		if ok {
			newName := entry.Text
			newPath := filepath.Join(filepath.Dir(oldPath), newName)
			err := os.Rename(oldPath, newPath)
			if err != nil {
				dialog.ShowError(err, m.window)
				return
			}

			// 更新树形视图
			m.treeView.Refresh()

			// 更新选中的节点
			m.selectedNode = widget.TreeNodeID(filepath.Join(filepath.Dir(string(uid)), newName))

			// 更新编辑区域的文件标签
			for i, tab := range m.tabs.Items {
				if strings.TrimPrefix(tab.Text, "*") == filepath.Base(oldPath) {
					m.tabs.Items[i].Text = newName
					if strings.HasPrefix(tab.Text, "*") {
						m.tabs.Items[i].Text = "*" + newName
					}

					// 更新 openFiles 映射
					if entry, ok := m.openFiles[oldPath]; ok {
						delete(m.openFiles, oldPath)
						m.openFiles[newPath] = entry
					}

					break
				}
			}
			m.tabs.Refresh()
		}
	}, m.window)
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
	m.showCustomFormDialog("New File", "Create", "Cancel", []*widget.FormItem{
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
	m.showCustomFormDialog("New Folder", "Create", "Cancel", []*widget.FormItem{
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

func (m *MarkdownEditor) delete(uid widget.TreeNodeID) {
	path := m.uidToPath(uid)
	dialog.ShowConfirm("Delete", "Are you sure you want to delete this item?", func(ok bool) {
		if ok {
			// 首先关闭文件（如果它在编辑区域中打开）
			fileName := filepath.Base(path)
			for _, tab := range m.tabs.Items {
				if strings.TrimPrefix(tab.Text, "*") == fileName {
					m.tabs.Remove(tab)
					delete(m.openFiles, path)
					break
				}
			}

			// 然后删除文件
			err := os.RemoveAll(path)
			if err != nil {
				dialog.ShowError(err, m.window)
				return
			}

			// 更新树形视图
			m.treeView.Refresh()

			// 清除选中的节点
			m.selectedNode = ""
		}
	}, m.window)
}

func (m *MarkdownEditor) uidToPath(uid widget.TreeNodeID) string {
	return filepath.Join(m.rootPath, uid)
}
