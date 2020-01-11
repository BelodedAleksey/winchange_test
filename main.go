package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"github.com/gonutz/w32"
	"github.com/nanitefactory/winmb"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

const (
	GW_HWNDFIRST = iota
	GW_HWNDLAST
	GW_HWNDNEXT
	HWNDPREV
	GW_OWNER
	GW_CHILD
)

const (
	MF_BYCOMMAND    = 0x00000000
	MF_BYPOSITION   = 0x00000400
	IMAGE_BITMAP    = 0
	LR_LOADFROMFILE = 0x00000010
)

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	getWindowProc        = user32.NewProc("GetWindow")
	getWindowTextProc    = user32.NewProc("GetWindowTextA")
	findWindowProc       = user32.NewProc("FindWindowA")
	getDesktopWindowProc = user32.NewProc("GetDesktopWindow")
	enumWindowsProc      = user32.NewProc("EnumWindows")
	setWindowTextProc    = user32.NewProc("SetWindowTextW")

	getMenuProc          = user32.NewProc("GetMenu")
	getMenuItemCountProc = user32.NewProc("GetMenuItemCount")
	getMenuItemIDProc    = user32.NewProc("GetMenuItemID")
	getSubMenuProc       = user32.NewProc("GetSubMenu")
	getMenuStringProc    = user32.NewProc("GetMenuStringA")

	getModuleHandleProc = kernel32.NewProc("GetModuleHandleW")

	setLayeredWindowAttributesProc = user32.NewProc("SetLayeredWindowAttributes")
)

//MenuItemInfo struct

func unicode2utf8(source string) string {
	var res = []string{""}
	sUnicode := strings.Split(source, "\\u")
	var context = ""
	for _, v := range sUnicode {
		var additional = ""
		if len(v) < 1 {
			continue
		}
		if len(v) > 4 {
			rs := []rune(v)
			v = string(rs[:4])
			additional = string(rs[4:])
		}
		temp, err := strconv.ParseInt(v, 16, 32)
		if err != nil {
			context += v
		}
		context += fmt.Sprintf("%c", temp)
		context += additional
	}
	res = append(res, context)
	return strings.Join(res, "")
}

//Смена кодировки с UTF-8 в ANSI
func utfToAnsi(str string) (string, error) {
	var windows1251 *charmap.Charmap = charmap.Windows1251
	bs := []byte(str)
	readerBs := bytes.NewReader(bs)
	readerWin := transform.NewReader(readerBs, windows1251.NewEncoder())
	bWin, err := ioutil.ReadAll(readerWin)
	if err != nil {
		return "", err
	}
	return string(bWin), nil
}

//Смена кодировки с UTF-8 в ANSI
func ansiToUTF(str string) (string, error) {
	var windows1251 *charmap.Charmap = charmap.Windows1251
	bs := []byte(str)
	readerBs := bytes.NewReader(bs)
	readerWin := transform.NewReader(readerBs, windows1251.NewDecoder())
	bWin, err := ioutil.ReadAll(readerWin)
	if err != nil {
		return "", err
	}
	return string(bWin), nil
}

//StringToUintptr func
func StringToUintptr(str string) uintptr {
	if str == "" {
		return 0
	}
	//enc := mahonia.NewEncoder("GBK")
	//str1 := enc.ConvertString(str)
	a1 := []byte(str)
	p1 := &a1[0]
	return uintptr(unsafe.Pointer(p1))
}

//FindWindow func
func FindWindow(className string, windowsName string) uintptr {
	ret, _, _ := findWindowProc.Call(
		StringToUintptr(className),
		StringToUintptr(windowsName),
	)
	return ret
}

//GetWindowText func
func GetWindowText(handler uintptr, maxCount uintptr) (string, uintptr) {
	var buf = make([]byte, maxCount)
	ret, _, _ := getWindowTextProc.Call(
		handler, uintptr(unsafe.Pointer(&buf[0])), maxCount,
	)
	out, _ := ansiToUTF(string(buf))
	return out, ret
}

//GetWindow func
func GetWindow(handler uintptr, wCmd uintptr) uintptr {
	ret, _, _ := getWindowProc.Call(
		handler, wCmd,
	)
	return ret
}

//GetDesktopWindow func
func GetDesktopWindow() uintptr {
	ret, _, _ := getDesktopWindowProc.Call()
	return ret
}

//SetWindowText func
func SetWindowText(hwnd uintptr, text string) bool {
	lpText := syscall.StringToUTF16Ptr(text)
	ret, _, _ := setWindowTextProc.Call(
		hwnd,
		uintptr(unsafe.Pointer(lpText)),
	)
	return ret != 0
}

//EnumWindowsByTitle func
func EnumWindowsByTitle(title string) []uintptr {
	firstWindow := GetDesktopWindow()
	handler := GetWindow(firstWindow, GW_CHILD)
	maxCount := 256
	var arr = make([]uintptr, 0)
	for handler != 0 {
		str, _ := GetWindowText(handler, uintptr(maxCount))
		if strings.Contains(str, title) {
			fmt.Println(str)
			arr = append(arr, handler)
		}
		handler = GetWindow(handler, GW_HWNDNEXT)
	}
	return arr
}

//GetMenu func
func GetMenu(hwnd uintptr) uintptr {
	ret, _, _ := getMenuProc.Call(hwnd)
	return ret
}

//GetMenuItemCount func
func GetMenuItemCount(hmenu uintptr) int {
	c, _, _ := getMenuItemCountProc.Call(hmenu)
	return int(c)
}

//GetMenuItemID func
func GetMenuItemID(hmenu uintptr, pos int) int32 {
	id, _, _ := getMenuItemIDProc.Call(
		hmenu,
		uintptr(pos),
	)
	return int32(id)
}

//GetSubMenu func
func GetSubMenu(hmenu uintptr, pos int) uintptr {
	m, _, _ := getSubMenuProc.Call(
		hmenu,
		uintptr(pos),
	)
	return m
}

//GetMenuString func
func GetMenuString(hmenu uintptr, uIDItem int32, fByPosition bool, maxCount uintptr) (string, int32) {
	var flags uint32
	if fByPosition {
		flags = MF_BYPOSITION
	} else {
		flags = MF_BYCOMMAND
	}
	var buf = make([]byte, maxCount)
	ret, _, _ := getMenuStringProc.Call(
		hmenu,
		uintptr(uIDItem),
		uintptr(unsafe.Pointer(&buf[0])),
		maxCount,
		uintptr(flags),
	)
	out, _ := ansiToUTF(string(buf))
	return out, int32(ret)
}

//GetModuleHandle func
func GetModuleHandle() uintptr {
	ret, _, _ := getModuleHandleProc.Call()
	return ret
}

//SetLayeredWindowAttributes func
func SetLayeredWindowAttributes(hwnd uintptr, crKey, bAlpha, dwFlags int32) bool {
	ret, _, _ := setLayeredWindowAttributesProc.Call(
		hwnd,
		uintptr(crKey),
		uintptr(bAlpha),
		uintptr(dwFlags),
	)
	return ret != 0
}

func main() {
	handlers := EnumWindowsByTitle("Блокнот")
	var targetHandle uintptr
	for _, val := range handlers {
		fmt.Println("Window: ", val)
		str, _ := GetWindowText(val, 256)
		fmt.Println("Text IFIX: ", str)
		targetHandle = val
	}
	//handlerFix := FindWindow("", "Безымянный — Блокнот")
	fmt.Println("Handler IFIX: ", targetHandle)
	str, _ := GetWindowText(targetHandle, 256)
	fmt.Println("Text IFIX: ", str)
	//SetWindowText(handlerFix, "ИДИ на хУЙ")
	handleMenu := GetMenu(targetHandle)
	fmt.Println("Handler Menu: ", handleMenu)
	//numItems := GetMenuItemCount(handleMenu)
	hinstance := GetModuleHandle()
	fmt.Println("HInstance: ", hinstance)

	txt, _ := GetMenuString(handleMenu, 0, true, 256)
	fmt.Println("Menu: ", txt)

	s := "lool"
	var info w32.MENUITEMINFO
	info.Mask = w32.MIIM_STRING
	info.TypeData = uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(s)))
	//w32.GetMenuItemInfo(w32.HMENU(handleMenu), 0, true, &info)
	w32.SetMenuItemInfo(w32.HMENU(handleMenu), 0, true, &info)
	w32.InsertMenuItem(w32.HMENU(handleMenu), 7, true, &info)
	w32.DrawMenuBar(w32.HWND(targetHandle))

	//Draw Transparent Window over another window
	var wc w32.WNDCLASSEX
	wc.Size = uint32(unsafe.Sizeof(wc))
	wc.Instance = 0
	wc.Style = w32.CS_HREDRAW | w32.CS_VREDRAW
	curs := uint16(w32.IDC_ARROW)
	wc.Cursor = w32.LoadCursor(0, &curs)
	wc.ClassName = syscall.StringToUTF16Ptr("MyTrans")
	wc.WndProc = syscall.NewCallback(
		func(hwnd w32.HWND, msg uint32, wParam, lParam uintptr) uintptr {
			switch msg {
			case w32.WM_PAINT:
				winmb.MessageBoxPlain("PAINT", "PAINT")
			case w32.WM_NCPAINT:
				winmb.MessageBoxPlain("PAINT", "PAINT")
			case w32.WM_SIZE:
				winmb.MessageBoxPlain("SIZE", "SIZE")
			default:
				w32.DefWindowProc(hwnd, msg, wParam, lParam)
			}
			return 0
			//return w32.CallWindowProc(origProc, hwnd, msg, wParam, lParam)
		})
	w32.RegisterClassEx(&wc)
	transHandle := w32.CreateWindowEx(
		w32.WS_EX_LAYERED,
		wc.ClassName,
		syscall.StringToUTF16Ptr(""),
		w32.WS_POPUP,
		200,
		200,
		800,
		600,
		0,
		0,
		0,
		unsafe.Pointer(uintptr(0)),
	)
	SetLayeredWindowAttributes(uintptr(transHandle), 0, 255, 1)
	w32.ShowWindow(transHandle, w32.SW_SHOWNORMAL)

	var msg w32.MSG
	for w32.GetMessage(&msg, 0, 0, 0) != 0 {
		w32.TranslateMessage(&msg)
		w32.DispatchMessage(&msg)
	}

	//Painting by hdc but it redraw by system
	/*prev := w32.POINT{}
	hdc := w32.GetDC(w32.HWND(targetHandle))
	w32.MoveToEx(hdc, 20, 20, &prev)
	w32.LineTo(hdc, 50, 50)
	w32.MoveToEx(hdc, int(prev.X), int(prev.Y), &prev)
	w32.ReleaseDC(w32.HWND(targetHandle), hdc)*/

	//Painting by catch WM_PAINT
	//origProc := w32.GetWindowLongPtr(w32.HWND(targetHandle), w32.GWLP_WNDPROC)

	//wndProc for WM_PAINT Not Working from another process need dll injection
	/*wndProc := func(hwnd w32.HWND, msg uint32, wParam, lParam uintptr) uintptr {
		switch msg {
		case w32.WM_PAINT:
			winmb.MessageBoxPlain("PAINT", "PAINT")
		case w32.WM_NCPAINT:
			winmb.MessageBoxPlain("PAINT", "PAINT")
		case w32.WM_SIZE:
			winmb.MessageBoxPlain("SIZE", "SIZE")
		}
		return w32.DefWindowProc(hwnd, msg, wParam, lParam)
		//return w32.CallWindowProc(origProc, hwnd, msg, wParam, lParam)
	}
	ret := w32.SetWindowLongPtr(
		w32.HWND(targetHandle), w32.GWLP_WNDPROC, syscall.NewCallback(wndProc),
	)
	fmt.Println("RET setwindowlongptr: ", ret)*/

	/*fileName, _ := syscall.UTF16PtrFromString("image.bmp")
	handleImage := w32.LoadImage(
		0, fileName, w32.IMAGE_ICON, 0, 0, w32.LR_LOADFROMFILE,
	)
	w32.SetMenuItemBitmaps(
		w32.HMENU(handleMenu),
		0,
		MF_BYPOSITION,
		w32.HBITMAP(handleImage),
		w32.HBITMAP(handleImage),
	)
	w32.DrawMenuBar(w32.HWND(targetHandle))*/

	/*for i := 0; i < numItems; i++ {
		id := GetMenuItemID(handleMenu, i)
		fmt.Println("ID: ", id)
		if id == -1 {
			subMenu := GetSubMenu(handleMenu, 0)
			fmt.Println("SubMenu: ", subMenu)
			numSubItems := GetMenuItemCount(subMenu)
			for j := 0; j < numSubItems; j++ {
				subID := GetMenuItemID(subMenu, j)
				fmt.Println("subID: ", subID)
				str, ret := GetMenuString(subMenu, subID, 256)
				fmt.Println("Text: ", str, "RetCode: ", ret)
			}
		}
	}*/
}
