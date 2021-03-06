// 窗体布局

package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"
	//"unicode"

	"gitee.com/ying32/govcl/vcl"
	"gitee.com/ying32/govcl/vcl/rtl"
	"gitee.com/ying32/govcl/vcl/types"
	"gitee.com/ying32/govcl/vcl/win"
)

var (
	fSortOrder bool
)

const E_WARN_STR = "生成警告"
const E_ERROT_STR = "生成错误"

type TFormConv struct {
	*vcl.TForm
	icon                    *vcl.TIcon        // ICON
	MainMenu                *vcl.TMainMenu    // 主菜单栏
	FrmAbout                *vcl.TForm        // 关于
	Panel                   *vcl.TPanel       // 布局panel
	Label1, Label2, Label3  *vcl.TLabel       // xlsx路径、输出路径、翻译路径标签
	InputCbox               *vcl.TComboBox    // xlsx路径选择框
	SearchCbox              *vcl.TComboBox    // 搜索框
	OutOutEdit              *vcl.TEdit        // 输出路径框
	LangEdit                *vcl.TEdit        // 翻译路径框
	Btn1, Btn2              *vcl.TButton      // 按钮(选择路径，生成配置)
	AllChkBox, ChangeChkBox *vcl.TCheckBox    // 全选、选择有变化的
	TestChkBox              *vcl.TCheckBox    // 实验性黑科技
	HideChkBox              *vcl.TCheckBox    // 隐藏未变化的
	ListView                *vcl.TListView    // 列表
	PrgBar                  *vcl.TProgressBar // 进度条
	Statusbar               *vcl.TStatusBar   // 底部状态栏
	Pmitem                  *vcl.TPopupMenu   // 右键菜单
	Inifile                 *vcl.TIniFile     // 历史记录
	History                 []string          // 分支路径
}

// 获取程序运行路径
func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return ""
	}
	return dir
	//return strings.Replace(dir, "\\", "/", -1)
}

/*------------------------private------------------------*/
func (f *TFormConv) getInPutDir() string {
	return f.InputCbox.Text()
}

func (f *TFormConv) getOutPutDir() string {
	return f.OutOutEdit.Text()
}

func (f *TFormConv) getLangDir() string {
	return f.LangEdit.Text()
}

func (f *TFormConv) getParentDir() string {
	inputDir := f.InputCbox.Text()
	parentDir, err := filepath.Abs(inputDir + "\\..")
	if err != nil {
		f.MsgBox("路径错误", "错误")
	}
	return parentDir
}

func (f *TFormConv) MsgBox(text, caption string) {
	if len(text) > 0 {
		vcl.Application.MessageBox(text, caption, win.MB_OK+win.MB_ICONINFORMATION)
	}
}

func (f *TFormConv) updateEdit() {
	if len(f.getInPutDir()) > 0 {
		dir := f.getParentDir()
		f.OutOutEdit.SetText(dir + "\\l-xlsx")
		f.LangEdit.SetText(FindLangFolder(dir))
	}
}

func (f *TFormConv) updateProcess() {
	old := f.PrgBar.Position()
	f.PrgBar.SetPosition(old + 1)
}

func (f *TFormConv) loadIni() {
	usr, _ := user.Current()
	iniFile := vcl.NewIniFile(usr.HomeDir + "\\xlsx2lua.ini")
	f.Inifile = iniFile

	for i := 1; i <= 10; i++ {
		history := iniFile.ReadString("History", fmt.Sprintf("path%d", i), "")
		if len(history) > 0 {
			f.History = append(f.History, history)
		}
	}
}

func (f *TFormConv) saveIni() {
	for i, his := range f.History {
		f.Inifile.WriteString("History", fmt.Sprintf("path%d", i+1), his)
	}
}

// 主菜单
func (f *TFormConv) initFormMenu() {
	mainForm := f.TForm
	mainMenu := vcl.NewMainMenu(f)
	f.MainMenu = mainMenu

	// 不自动生成热键
	mainMenu.SetAutoHotkeys(types.MaManual)
	// 一级菜单
	item := vcl.NewMenuItem(mainForm)
	item.SetCaption("文件(&F)")

	subMenu := vcl.NewMenuItem(mainForm)
	subMenu.SetCaption("新建(&N)")
	subMenu.SetShortCutFromString("Ctrl+N")
	subMenu.SetOnClick(func(vcl.IObject) {
		//fmt.Println("单击了新建")
	})
	item.Add(subMenu)

	subMenu = vcl.NewMenuItem(mainForm)
	subMenu.SetCaption("打开(&O)")
	subMenu.SetShortCutFromString("Ctrl+O")
	item.Add(subMenu)

	subMenu = vcl.NewMenuItem(mainForm)
	subMenu.SetCaption("保存(&S)")
	subMenu.SetShortCutFromString("Ctrl+S")
	item.Add(subMenu)

	// 分割线
	subMenu = vcl.NewMenuItem(mainForm)
	subMenu.SetCaption("-")
	item.Add(subMenu)

	subMenu = vcl.NewMenuItem(mainForm)
	subMenu.SetCaption("退出(&Q)")
	subMenu.SetShortCutFromString("Ctrl+Q")
	subMenu.SetOnClick(func(vcl.IObject) {
		mainForm.Close()
	})
	item.Add(subMenu)

	mainMenu.Items().Add(item)

	item = vcl.NewMenuItem(mainForm)
	item.SetCaption("帮助(&H)")

	subMenu = vcl.NewMenuItem(mainForm)
	subMenu.SetCaption("关于(&A)")
	item.Add(subMenu)
	mainMenu.Items().Add(item)
	subMenu.SetOnClick(func(vcl.IObject) {
		f.FrmAbout.ShowModal()
	})

	// 状态栏
	statusbar := vcl.NewStatusBar(mainForm)
	statusbar.SetParent(mainForm)
	statusbar.SetName("statusbar")
	statusbar.SetSizeGrip(false) // 右下角出现可调整窗口三角形，默认显示
	pnl := statusbar.Panels().Add()
	pnl.SetText("文件数量:0")
	pnl.SetWidth(100)
	pn2 := statusbar.Panels().Add()
	pn2.SetText("有变化的数量:0")
	pn2.SetWidth(200)
	pn3 := statusbar.Panels().Add()
	pn3.SetText("总耗时(ms):0")
	pn3.SetWidth(100)
	f.Statusbar = statusbar
}

func (f *TFormConv) initfrmAbout() {
	frmAbout := vcl.Application.CreateForm()
	frmAbout.ScreenCenter()
	frmAbout.SetCaption("关于")
	frmAbout.SetBorderStyle(types.BsSingle)
	frmAbout.EnabledMaximize(false)
	frmAbout.EnabledMinimize(false)
	frmAbout.SetWidth(405)
	frmAbout.SetHeight(210)
	f.FrmAbout = frmAbout

	about := vcl.NewLabel(frmAbout)
	about.SetParent(frmAbout)
	about.SetAlign(types.AlClient)
	//about.SetTop(frmAbout.ClientHeight() / 2)
	about.SetAutoSize(false)
	about.SetAlignment(types.TaCenter)
	about.SetLayout(types.TlCenter)
	//about.SetStyleElements(types.AkRight)
	about.SetCaption("这是一个奇怪的工具\r\ndomi © 2018")

	//	btn := vcl.NewButton(frmAbout)
	//	btn.SetParent(frmAbout)
	//	btn.SetCaption("OK")
	//	btn.SetModalResult(types.MbOK)
	//	btn.SetLeft(frmAbout.ClientWidth() - btn.Width() - 10)
	//	btn.SetTop(frmAbout.ClientHeight() - btn.Height() - 10)
}

// 初始化panel内的布局
func (f *TFormConv) initPanel() {
	mainForm := f.TForm
	pnl := vcl.NewPanel(mainForm)
	pnl.SetParent(mainForm)
	pnl.SetHeight(130)
	pnl.SetAlign(types.AlTop)
	f.Panel = pnl

	_createLabel := func(caption string, left, top int32) *vcl.TLabel {
		label := vcl.NewLabel(mainForm)
		label.SetLeft(left)
		label.SetTop(top)
		label.SetCaption(caption)
		label.SetParent(pnl)
		return label
	}
	_createEdit := func(caption string, left, top int32) *vcl.TEdit {
		edit := vcl.NewEdit(mainForm)
		edit.SetLeft(left)
		edit.SetTop(top)
		edit.SetText(caption)
		edit.SetWidth(300)
		edit.SetReadOnly(true)
		edit.SetParent(pnl)
		return edit
	}
	_createBtn := func(caption string, left, top int32) *vcl.TButton {
		btn := vcl.NewButton(mainForm)
		btn.SetParent(pnl)
		btn.SetLeft(left)
		btn.SetTop(top)
		btn.SetWidth(100)
		btn.SetCaption(caption)
		return btn
	}
	_createChkBox := func(caption string, left, top int32) *vcl.TCheckBox {
		chkBox := vcl.NewCheckBox(mainForm)
		chkBox.SetParent(pnl)
		chkBox.SetLeft(left)
		chkBox.SetTop(top)
		chkBox.SetCaption(caption)
		return chkBox
	}

	// 第1行
	{
		left, top := int32(10), int32(20)
		f.Label1 = _createLabel("配置路径：", left, top)

		// 路径input
		left += f.Label1.Width() + 5
		cbox := vcl.NewComboBox(mainForm)
		cbox.SetParent(pnl)
		cbox.SetLeft(left)
		cbox.SetTop(top)
		cbox.SetWidth(300)
		cbox.SetStyle(types.CsOwnerDrawFixed)
		for _, his := range f.History {
			cbox.Items().Add(his)
		}
		cbox.SetItemIndex(0)
		f.InputCbox = cbox

		// btn
		top -= 5
		left += cbox.Width() + 10
		f.Btn1 = _createBtn("选择路径", left, top)

		left += f.Btn1.Width() + 10
		f.Btn2 = _createBtn("生成配置", left, top)
		f.Btn2.SetHint("生成配置过程中，其他操作将无法进行")

		left += f.Btn2.Width() + 10
		refreshBtn := _createBtn("重新载入", left, top)
		refreshBtn.SetHint("重新载入配置")
		refreshBtn.SetOnClick(func(vcl.IObject) {
			f.LoadXlxs()
		})
	}

	// 第2行
	{
		left, top := int32(10), f.Label1.Top()+f.Label1.Height()+20
		f.Label2 = _createLabel("输出路径：", left, top)
		left += f.Label2.Width() + 5
		f.OutOutEdit = _createEdit("", left, top)

		left += f.OutOutEdit.Width() + 10
		prgLable := _createLabel("生成进度：", left, top+5)
		left += prgLable.Width() + 5
		prgbar := vcl.NewProgressBar(mainForm)
		prgbar.SetParent(mainForm)
		prgbar.SetBounds(left, top, 255, f.OutOutEdit.Height())
		prgbar.SetMin(0)
		prgbar.SetPosition(0)
		prgbar.SetOrientation(types.PbHorizontal)
		f.PrgBar = prgbar
	}

	// 第3行
	{
		left, top := int32(10), f.Label2.Top()+f.Label2.Height()+20
		f.Label3 = _createLabel("翻译路径：", left, top)
		left += f.Label3.Width() + 5
		f.LangEdit = _createEdit("", left, top)

		left += f.LangEdit.Width() + 10
		cbox := vcl.NewComboBox(mainForm)
		cbox.SetParent(pnl)
		cbox.SetLeft(left)
		cbox.SetTop(top)
		cbox.SetWidth(320)
		f.SearchCbox = cbox
		cbox.SetOnChange(func(sender vcl.IObject) {
			str := cbox.Text()
			if cbox.Items().IndexOf(str) != -1 {
				var i int32
				lv := f.ListView
				lvCount := lv.Items().Count()
				for i = 0; i < lvCount; i++ {
					item := lv.Items().Item(i)
					if item.Caption() == str {
						lv.SetSelected(item)
						item.MakeVisible(true)
						break
					}
				}
				return
			}

			res := FuzzyMatching.ClosestN(str, 5)
			cbox.Items().Clear()
			cbox.SetSelStart(int32(len(str)))
			for _, his := range res {
				cbox.Items().Add(his)
			}
			cbox.SetDroppedDown(true)
		})
	}

	// 第4行
	{
		left, top := int32(10), f.Label3.Top()+f.Label3.Height()+10
		f.AllChkBox = _createChkBox("全选", left, top)
		left = left + f.AllChkBox.Width() + 10

		f.ChangeChkBox = _createChkBox("选择有变化的", left, top)
		f.ChangeChkBox.SetChecked(true)
		left = left + f.ChangeChkBox.Width() + 10

		f.TestChkBox = _createChkBox("实验性特性", left, top)
		f.TestChkBox.SetHint("针对翻译，能够检查翻译和源配置的英文字符是否一一对应")

		left = left + f.TestChkBox.Width() + 10
		f.HideChkBox = _createChkBox("隐藏未变化文件", left, top)
		f.HideChkBox.SetWidth(150)
	}
	f.updateEdit()
}

func (f *TFormConv) initListView() {
	// TPopupMenu
	mainForm := f.TForm
	pm := vcl.NewPopupMenu(mainForm)
	pmitem := vcl.NewMenuItem(mainForm)
	pmitem.SetCaption("打开文件")
	pm.Items().Add(pmitem)

	pmLang := vcl.NewMenuItem(mainForm)
	pmLang.SetCaption("打开翻译文件")
	pm.Items().Add(pmLang)

	line := vcl.NewMenuItem(mainForm)
	line.SetCaption("-")
	pm.Items().Add(line)

	pmitem1 := vcl.NewMenuItem(mainForm)
	pmitem1.SetCaption("打开文件所在目录")
	pm.Items().Add(pmitem1)

	pmitem2 := vcl.NewMenuItem(mainForm)
	pmitem2.SetCaption("打开输出目录")
	pm.Items().Add(pmitem2)

	pmitem3 := vcl.NewMenuItem(mainForm)
	pmitem3.SetCaption("打开翻译目录")
	pm.Items().Add(pmitem3)

	pmitem4 := vcl.NewMenuItem(mainForm)
	pmitem4.SetCaption("打开Config目录")
	pm.Items().Add(pmitem4)

	line = vcl.NewMenuItem(mainForm)
	line.SetCaption("-")
	pm.Items().Add(line)

	pmErr := vcl.NewMenuItem(mainForm)
	pmErr.SetCaption("显示错误")
	pm.Items().Add(pmErr)
	f.Pmitem = pm

	line = vcl.NewMenuItem(mainForm)
	line.SetCaption("-")
	pm.Items().Add(line)

	svnUpPmitem := vcl.NewMenuItem(mainForm)
	svnUpPmitem.SetCaption("SVN更新")
	pm.Items().Add(svnUpPmitem)
	svnCiPmitem := vcl.NewMenuItem(mainForm)
	svnCiPmitem.SetCaption("SVN提交")
	pm.Items().Add(svnCiPmitem)

	// 生成结果列表
	imgList := vcl.NewImageList(mainForm)
	//imgList.SetHeight(100)
	imgList.SetWidth(1)
	lv1 := vcl.NewListView(mainForm)
	lv1.SetParent(mainForm)
	lv1.SetWidth(mainForm.ClientWidth())
	lv1.SetAlign(types.AlClient)
	//lv1.SetClientWidth(300)
	lv1.SetSmallImages(imgList)
	lv1.SetRowSelect(true)
	lv1.SetReadOnly(true)
	lv1.SetGridLines(true)
	lv1.SetViewStyle(types.VsReport)
	lv1.Font().SetName("微软雅黑")
	lv1.Font().SetSize(10)
	lv1.SetCheckboxes(true)
	lv1.SetPopupMenu(pm)
	f.ListView = lv1

	addCol := func(caption string, width int32, autosize bool) {
		col := lv1.Columns().Add()
		col.SetCaption(caption)
		col.SetWidth(width)
		col.SetAutoSize(autosize)
		if !autosize {
			col.SetMaxWidth(width)
			col.SetMinWidth(width)
		}
		col.SetAlignment(types.TaLeftJustify)
	}
	addCol("文件名", lv1.ClientWidth()-317, true)
	addCol("文件状态", 100, false)
	addCol("生成结果", 200, false)

	// 右键菜单相应
	pmitem.SetOnClick(func(vcl.IObject) {
		item := f.ListView.Selected()
		if item.IsValid() {
			rtl.SysOpen(item.Caption())
		}
		//cmdStr := exec.Command("cmd", "/C start "+item.Caption())
		//go cmdStr.Run()
	})
	pmLang.SetOnClick(func(vcl.IObject) {
		item := f.ListView.Selected()
		if !item.IsValid() {
			return
		}
		idx := int(item.Data())
		if idx >= len(Convs) {
			return
		}
		c := Convs[idx]
		dir := f.getLangDir()
		if len(dir) > 0 {
			fname := dir + "\\" + strings.Replace("\\"+c.RelPath, "\\", "$", -1)
			rtl.SysOpen(fname)
		}
	})
	pmitem1.SetOnClick(func(vcl.IObject) {
		item := f.ListView.Selected()
		if item.IsValid() {
			rtl.SysOpen(rtl.ExtractFilePath(item.Caption()))
		}
	})
	pmitem2.SetOnClick(func(vcl.IObject) {
		item := f.ListView.Selected()
		if item.IsValid() {
			idx := int(item.Data())
			dir := f.getOutPutDir() + "\\" + Convs[idx].FolderName
			rtl.SysOpen(dir)
		}
	})
	pmitem3.SetOnClick(func(vcl.IObject) {
		dir := f.getLangDir()
		if len(dir) > 0 {
			rtl.SysOpen(dir)
		}
	})
	pmitem4.SetOnClick(func(vcl.IObject) {
		dir := f.getParentDir()
		if len(dir) > 0 {
			rtl.SysOpen(dir)
		}
	})
	pmErr.SetOnClick(func(vcl.IObject) {
		item := f.ListView.Selected()
		if item.IsValid() {
			idx := int(item.Data())
			f.MsgBox(Convs[idx].formatErr(), "生成结果")
		}
	})
	svnUpPmitem.SetOnClick(func(vcl.IObject) {
		if len(f.getInPutDir()) == 0 {
			f.MsgBox("请选择配置路径(xlsx文件夹)", "错误")
			return
		}
		_, err := exec.LookPath("TortoiseProc")
		if err == nil {
			command := fmt.Sprintf(`/command:update /path:%s /closeonend:0`, f.getParentDir())
			cmdStr := exec.Command("TortoiseProc", command)
			err = cmdStr.Run()
			if err != nil {
				f.MsgBox("SVN更新错误", "错误")
			} else {
				f.LoadXlxs()
			}
		} else {
			f.MsgBox("请先安装TortoiseProc", "错误")
		}
	})
	svnCiPmitem.SetOnClick(func(vcl.IObject) {
		if len(f.getInPutDir()) == 0 {
			f.MsgBox("请选择配置路径(xlsx文件夹)", "错误")
			return
		}
		_, err := exec.LookPath("TortoiseProc")
		if err == nil {
			command := fmt.Sprintf(`/command:commit /path:%s\ /closeonend:0`, f.getParentDir())
			cmdStr := exec.Command("TortoiseProc", command)
			err = cmdStr.Run()
			if err != nil {
				f.MsgBox("SVN提交错误", "错误")
			}
		} else {
			f.MsgBox("请先安装TortoiseProc", "错误")
		}
	})
}

// 设置控件的事件
func (f *TFormConv) setEvent() {
	cbox := f.InputCbox
	lv1 := f.ListView
	btn1, btn2 := f.Btn1, f.Btn2
	allChkBox := f.AllChkBox
	cbox.SetOnChange(func(vcl.IObject) {
		if cbox.ItemIndex() != -1 {
			f.updateEdit()
			f.LoadXlxs()
			f.SearchCbox.Clear()
		}
	})

	// listview 排序
	lv1.SetOnCompare(lvTraiCompare)
	lv1.SetOnColumnClick(func(sender vcl.IObject, column *vcl.TListColumn) {
		// 按柱头索引排序, lcl兼容版第二个参数永远为 column
		fSortOrder = !fSortOrder
		lv1.CustomSort(0, int(column.Index()))
	})
	lv1.SetOnDblClick(func(sender vcl.IObject) {
		item := f.ListView.Selected()
		item.SetChecked(!item.Checked())
	})
	lv1.SetOnAdvancedCustomDrawItem(func(sender *vcl.TListView, item *vcl.TListItem, state types.TCustomDrawState, Stage types.TCustomDrawStage, defaultDraw *bool) {
		canvas := sender.Canvas()
		//font := canvas.Font()
		i := int(item.Index())
		if i%2 == 0 {
			canvas.Brush().SetColor(0x02F0EEF7)
		}

		resStr := item.SubItems().Strings(1)
		if resStr == E_ERROT_STR {
			canvas.Brush().SetColor(types.ClRed)
			//font.SetColor(types.ClSilver)
		} else if resStr == E_WARN_STR {
			canvas.Brush().SetColor(types.ClYellow)
			//font.SetColor(types.ClSilver)
		}
	})

	// button
	btn1.SetOnClick(func(vcl.IObject) {
		options := types.TSelectDirExtOpts(rtl.Include(0, types.SdNewFolder, types.SdShowEdit, types.SdNewUI))
		if ok, dir := vcl.SelectDirectory2("选择配置路径", "C:/", options, nil); ok {
			f.History = append(f.History, dir)
			cbox.SetText(dir)
			idx := cbox.Items().Add(dir)
			cbox.SetItemIndex(idx)

			f.updateEdit()
			f.LoadXlxs()
		}
	})
	btn2.SetOnClick(func(vcl.IObject) {
		count := lv1.Items().Count()
		idxs := make(map[int]bool, count)
		var i int32
		for i = 0; i < count; i++ {
			item := lv1.Items().Item(i)
			if item.Checked() {
				idxs[int(item.Data())] = true
			}
		}
		if len(idxs) > 0 {
			f.Panel.SetEnabled(false)
			f.ListView.SetEnabled(false)
			f.PrgBar.SetMax(int32(len(idxs)))

			go startConv(idxs)
		} else {
			f.MsgBox("请选择配置", "通知")
		}
	})

	// check box
	allChkBox.SetOnClick(func(vcl.IObject) {
		var i int32
		for i = 0; i < lv1.Items().Count(); i++ {
			lv1.Items().Item(i).SetChecked(allChkBox.Checked())
		}
	})
	f.ChangeChkBox.SetOnClick(func(vcl.IObject) {
		listView := f.ListView
		var i int32
		count := listView.Items().Count()
		for i = 0; i < count; i++ {
			item := listView.Items().Item(i)
			idx := int(item.Data())
			if Convs[idx].hasChanged() {
				item.SetChecked(f.ChangeChkBox.Checked())
			}
		}
	})
	f.HideChkBox.SetOnClick(func(vcl.IObject) {
		f.updateListView()
	})
}

// 排序
func lvTraiCompare(sender vcl.IObject, item1, item2 *vcl.TListItem, data int32, compare *int32) {
	var s1, s2 string
	if data != 0 {
		s1 = item1.SubItems().Strings(data - 1)
		s2 = item2.SubItems().Strings(data - 1)
	} else {
		s1 = item1.Caption()
		s2 = item2.Caption()
	}
	if fSortOrder {
		*compare = int32(strings.Compare(s1, s2))
	} else {
		*compare = -int32(strings.Compare(s1, s2))
	}
}

func (f *TFormConv) updateListView() {
	listView := f.ListView
	listView.Items().Clear()
	listView.Items().BeginUpdate()

	changeCount := 0
	selectChange := f.ChangeChkBox.Checked()
	hideNochange := f.HideChkBox.Checked()
	for i, conv := range Convs {
		isChange := conv.hasChanged()
		if hideNochange {
			if isChange {
				changeCount += 1
				item := listView.Items().Add()
				item.SetCaption(conv.AbsPath) // 第一列为Caption属性所管理
				item.SetChecked(selectChange)
				item.SubItems().Add("配置有变化")
				item.SubItems().Add("-")
				item.SetData(uintptr(i))
			}
		} else {
			item := listView.Items().Add()
			item.SetCaption(conv.AbsPath) // 第一列为Caption属性所管理
			if isChange {
				changeCount += 1
				item.SetChecked(selectChange)
				item.SubItems().Add("配置有变化")
			} else {
				item.SubItems().Add("-")
			}
			item.SubItems().Add("-")
			item.SetData(uintptr(i))
		}
	}
	listView.Items().EndUpdate()
	listView.CustomSort(0, int(1)) // 按是否变化排序列表

	f.PrgBar.SetPosition(0)
	f.Statusbar.Panels().Items(0).SetText(fmt.Sprintf("文件数量：%d", int32(len(Convs))))
	f.Statusbar.Panels().Items(1).SetText(fmt.Sprintf("有变化的数量：%d", changeCount))
}

/*------------------------public------------------------*/
func CreateMainForm() *TFormConv {
	form := new(TFormConv)

	// icon
	icon := vcl.NewIcon()
	icon.LoadFromResourceID(rtl.MainInstance(), 3)
	vcl.Application.Initialize()
	vcl.Application.SetMainFormOnTaskBar(true)
	vcl.Application.SetIcon(icon)

	mainForm := vcl.Application.CreateForm()
	mainForm.SetCaption("xlsxconv")
	mainForm.ScreenCenter()
	mainForm.SetPosition(types.PoScreenCenter)
	mainForm.EnabledMaximize(false)
	//mainForm.SetBorderStyle(types.BsSingle)
	mainForm.SetWidth(1024)
	mainForm.SetHeight(800)
	mainForm.SetDoubleBuffered(true)
	mainForm.SetShowHint(true)

	form.icon = icon
	form.TForm = mainForm
	form.History = make([]string, 0, 10)
	return form
}

// 创建窗体内的控件
func (f *TFormConv) CreateControl() {
	f.loadIni()
	f.initFormMenu()
	f.initfrmAbout()
	f.initPanel()
	f.initListView()
	f.setEvent()
}

func (f *TFormConv) LoadXlxs() {
	dir := f.InputCbox.Text()
	if len(dir) > 0 {
		err := WalkXlsx(dir)
		f.updateListView()
		if err == nil {
			f.saveIni()
		} else {
			f.MsgBox(err.Error(), "加载配置错误")
		}
	}
}

func (f *TFormConv) ConvResult(idxs map[int]bool, startTime time.Time) {
	f.ListView.SetEnabled(true)
	listView := f.ListView

	var i int32
	var c *XlsxConv
	var okCount, errCount, warnCount int
	count := listView.Items().Count()
	for i = 0; i < count; i++ {
		item := listView.Items().Item(i)
		idx := int(item.Data())
		if _, ok := idxs[idx]; ok {
			c = Convs[idx]
			//item.SubItems().BeginUpdate()
			if c.hasError(E_ERROR) {
				errCount++
				item.SubItems().SetStrings(1, E_ERROT_STR)
			} else if c.hasError(E_WARN) {
				warnCount++
				item.SubItems().SetStrings(0, "-")
				item.SubItems().SetStrings(1, E_WARN_STR)
			} else {
				okCount++
				item.SubItems().SetStrings(0, "-")
				item.SubItems().SetStrings(1, fmt.Sprintf("耗时(ms):%d", c.Msec))
			}
			//item.SubItems().EndUpdate()
		}
	}
	f.Panel.SetEnabled(true)
	f.PrgBar.SetPosition(0)
	f.Statusbar.Panels().Items(2).SetText(fmt.Sprintf("总耗时(ms)：%d", int(time.Now().Sub(startTime).Nanoseconds()/1e6)))
	f.MsgBox(fmt.Sprintf("错误：%d条，警告：%d条，成功：%d条", errCount, warnCount, okCount), "生成结果")
}
