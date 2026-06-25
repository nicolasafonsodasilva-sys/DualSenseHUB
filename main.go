// SPDX-License-Identifier: MIT
//go:build windows

package main

import (
	"embed"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"image"
	"image/draw"
	_ "image/png"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"
)

//go:embed assets/*.png
var overlayAssets embed.FS

const (
	appName               = "DualSenseHUB"
	windowClassName       = "DualSenseHUBOverlayWindow"
	overlayWidth          = 212
	overlayHeight         = 84
	overlayRightMargin    = 64
	overlayTopMargin      = 68
	overlayVisibleTime    = 3 * time.Second
	lowBatteryVisibleTime = 5 * time.Second
	lowBatteryThreshold   = 5
	psHoldTime            = 3 * time.Second
	controllerTimeout     = 750 * time.Millisecond

	WM_DESTROY             = 0x0002
	WM_CLOSE               = 0x0010
	WM_SETICON             = 0x0080
	WM_DISPLAYCHANGE       = 0x007E
	WM_INPUT_DEVICE_CHANGE = 0x00FE
	WM_INPUT               = 0x00FF
	WM_TIMER               = 0x0113
	WM_APP_BATTERY         = 0x8001

	GIDC_ARRIVAL = 1
	GIDC_REMOVAL = 2

	RIM_TYPEHID     = 2
	RID_INPUT       = 0x10000003
	RIDI_DEVICEINFO = 0x2000000B
	RIDEV_PAGEONLY  = 0x00000020
	RIDEV_INPUTSINK = 0x00000100
	RIDEV_DEVNOTIFY = 0x00002000

	RIDI_DEVICENAME = 0x20000007

	WS_POPUP = 0x80000000

	WS_EX_TRANSPARENT = 0x00000020
	WS_EX_TOOLWINDOW  = 0x00000080
	WS_EX_TOPMOST     = 0x00000008
	WS_EX_LAYERED     = 0x00080000
	WS_EX_NOACTIVATE  = 0x08000000

	CS_HREDRAW = 0x0002
	CS_VREDRAW = 0x0001

	SW_HIDE           = 0
	SW_SHOWNOACTIVATE = 4

	IMAGE_ICON  = 1
	ICON_SMALL  = 0
	ICON_BIG    = 1
	ICON_SMALL2 = 2

	SM_CXSCREEN = 0
	SM_CYSCREEN = 1

	DIB_RGB_COLORS = 0
	BI_RGB         = 0

	ULW_ALPHA    = 0x00000002
	AC_SRC_OVER  = 0
	AC_SRC_ALPHA = 1

	REG_SZ        = 1
	KEY_SET_VALUE = 0x0002

	ERROR_ALREADY_EXISTS = 183

	CREATE_NO_WINDOW = 0x08000000

	PROCESS_SYNCHRONIZE = 0x00100000
	WAIT_OBJECT_0       = 0x00000000
	WAIT_TIMEOUT        = 0x00000102

	MOVEFILE_REPLACE_EXISTING = 0x00000001
	MOVEFILE_WRITE_THROUGH    = 0x00000008

	MB_OK        = 0x00000000
	MB_ICONERROR = 0x00000010
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	advapi32 = syscall.NewLazyDLL("advapi32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")

	procRegisterClassExW          = user32.NewProc("RegisterClassExW")
	procCreateWindowExW           = user32.NewProc("CreateWindowExW")
	procDefWindowProcW            = user32.NewProc("DefWindowProcW")
	procFindWindowW               = user32.NewProc("FindWindowW")
	procPostMessageW              = user32.NewProc("PostMessageW")
	procGetWindowThreadProcessId  = user32.NewProc("GetWindowThreadProcessId")
	procMessageBoxW               = user32.NewProc("MessageBoxW")
	procLoadImageW                = user32.NewProc("LoadImageW")
	procSendMessageW              = user32.NewProc("SendMessageW")
	procGetMessageW               = user32.NewProc("GetMessageW")
	procTranslateMessage          = user32.NewProc("TranslateMessage")
	procDispatchMessageW          = user32.NewProc("DispatchMessageW")
	procPostQuitMessage           = user32.NewProc("PostQuitMessage")
	procRegisterRawInputDevices   = user32.NewProc("RegisterRawInputDevices")
	procGetRawInputData           = user32.NewProc("GetRawInputData")
	procGetRawInputDeviceList     = user32.NewProc("GetRawInputDeviceList")
	procGetRawInputDeviceInfoW    = user32.NewProc("GetRawInputDeviceInfoW")
	procSetTimer                  = user32.NewProc("SetTimer")
	procShowWindow                = user32.NewProc("ShowWindow")
	procGetSystemMetrics          = user32.NewProc("GetSystemMetrics")
	procUpdateLayeredWindow       = user32.NewProc("UpdateLayeredWindow")
	procGetDC                     = user32.NewProc("GetDC")
	procReleaseDC                 = user32.NewProc("ReleaseDC")
	procSetProcessDPIAwareContext = user32.NewProc("SetProcessDpiAwarenessContext")

	procCreateCompatibleDC = gdi32.NewProc("CreateCompatibleDC")
	procCreateDIBSection   = gdi32.NewProc("CreateDIBSection")
	procSelectObject       = gdi32.NewProc("SelectObject")
	procDeleteObject       = gdi32.NewProc("DeleteObject")
	procDeleteDC           = gdi32.NewProc("DeleteDC")

	procGetModuleHandleW    = kernel32.NewProc("GetModuleHandleW")
	procCreateMutexW        = kernel32.NewProc("CreateMutexW")
	procGetLastError        = kernel32.NewProc("GetLastError")
	procOpenProcess         = kernel32.NewProc("OpenProcess")
	procWaitForSingleObject = kernel32.NewProc("WaitForSingleObject")
	procCloseHandle         = kernel32.NewProc("CloseHandle")
	procMoveFileExW         = kernel32.NewProc("MoveFileExW")

	procRegCreateKeyExW = advapi32.NewProc("RegCreateKeyExW")
	procRegSetValueExW  = advapi32.NewProc("RegSetValueExW")
	procRegDeleteValueW = advapi32.NewProc("RegDeleteValueW")
	procRegCloseKey     = advapi32.NewProc("RegCloseKey")

	procSHChangeNotify = shell32.NewProc("SHChangeNotify")
)

type point struct {
	X int32
	Y int32
}

type size struct {
	CX int32
	CY int32
}

type msg struct {
	Hwnd     uintptr
	Message  uint32
	_        uint32
	WParam   uintptr
	LParam   uintptr
	Time     uint32
	Pt       point
	LPrivate uint32
}

type wndClassEx struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       uintptr
}

type rawInputDevice struct {
	UsagePage uint16
	Usage     uint16
	Flags     uint32
	Target    uintptr
}

type rawInputDeviceList struct {
	Device uintptr
	Type   uint32
}

type reportDiagnosticState struct {
	count       uint64
	lastLogAt   time.Time
	lastReport  []byte
	lastDecoded dualSenseReport
}

type rawInputHeader struct {
	Type   uint32
	Size   uint32
	Device uintptr
	WParam uintptr
}

type bitmapInfoHeader struct {
	Size          uint32
	Width         int32
	Height        int32
	Planes        uint16
	BitCount      uint16
	Compression   uint32
	SizeImage     uint32
	XPelsPerMeter int32
	YPelsPerMeter int32
	ClrUsed       uint32
	ClrImportant  uint32
}

type bitmapInfo struct {
	Header bitmapInfoHeader
	Colors [1]uint32
}

type blendFunction struct {
	BlendOp             byte
	BlendFlags          byte
	SourceConstantAlpha byte
	AlphaFormat         byte
}

type overlayFrame struct {
	width  int
	height int
	pixels []byte // premultiplied BGRA, top-down
}

type controllerState struct {
	batteryKnown      bool
	batteryPercent    int
	charging          bool
	psDown            bool
	psDownAt          time.Time
	lastReportAt      time.Time
	shutdownTriggered bool
	overlayVisible    bool
	overlayLowBattery bool
	overlayHideAt     time.Time
	lowBatteryAlerted bool
	lastDevice        uintptr
	simpleMode        bool
	lastBatteryPollAt time.Time
	lastFullReportAt  time.Time
}

var (
	mainWindow          uintptr
	state               controllerState
	frameCache          = make(map[string]*overlayFrame)
	deviceCache         = make(map[uintptr]bool)
	batteryPollInFlight atomic.Bool
	reportDiagnosticMu  sync.Mutex
	reportDiagnostics   = make(map[uintptr]*reportDiagnosticState)
)

func utf16Ptr(s string) *uint16 {
	p, _ := syscall.UTF16PtrFromString(s)
	return p
}

func loword(v uintptr) uint16 { return uint16(v & 0xFFFF) }

func main() {
	initDiagnosticLog()
	defer closeDiagnosticLog()
	diagnosticLogf("main entered")
	runtime.LockOSThread()
	diagnosticLogf("OS thread locked")

	// Per-monitor DPI awareness keeps the overlay at the exact pixel size from the reference image.
	r, _, dpiErr := procSetProcessDPIAwareContext.Call(^uintptr(3)) // DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2 == -4
	diagnosticLogf("SetProcessDpiAwarenessContext result=%d err=%v", r, dpiErr)

	// A downloaded copy acts only as an installer/updater. It copies itself to LocalAppData,
	// waits for any old instance to fully exit, starts the installed copy, and then exits.
	// This avoids the update race that could close the old app before the new one was ready.
	if installAndRelaunchIfNeeded() {
		diagnosticLogf("installer/updater instance finished")
		return
	}
	diagnosticLogf("running from installed path")

	if !acquireSingleInstance() {
		diagnosticLogf("single-instance mutex already exists; exiting")
		return
	}
	diagnosticLogf("single-instance mutex acquired")

	removeLegacyStartupEntry()
	removeLegacyInstalledCopy()
	installStartupEntry()
	installStartMenuShortcut()

	hInstance, _, _ := procGetModuleHandleW.Call(0)
	className := utf16Ptr(windowClassName)
	windowProc := syscall.NewCallback(wndProc)

	// Load the embedded icon explicitly for the hidden window class.
	// Task Manager may use the window/class icon instead of extracting the file icon,
	// so setting both prevents it from falling back to an old/default cached icon.
	hIconBig, _, bigIconErr := procLoadImageW.Call(hInstance, 1, IMAGE_ICON, 32, 32, 0)
	hIconSmall, _, smallIconErr := procLoadImageW.Call(hInstance, 1, IMAGE_ICON, 16, 16, 0)
	diagnosticLogf("LoadImageW icons big=0x%X err=%v small=0x%X err=%v", hIconBig, bigIconErr, hIconSmall, smallIconErr)

	wc := wndClassEx{
		CbSize:        uint32(unsafe.Sizeof(wndClassEx{})),
		Style:         CS_HREDRAW | CS_VREDRAW,
		LpfnWndProc:   windowProc,
		HInstance:     hInstance,
		HIcon:         hIconBig,
		LpszClassName: className,
		HIconSm:       hIconSmall,
	}
	if r, _, err := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc))); r == 0 {
		diagnosticLogf("RegisterClassExW FAILED err=%v", err)
		return
	} else {
		diagnosticLogf("RegisterClassExW ok atom=%d", r)
	}

	x, y := overlayPosition(overlayWidth)
	exStyle := uintptr(WS_EX_LAYERED | WS_EX_TRANSPARENT | WS_EX_TOOLWINDOW | WS_EX_TOPMOST | WS_EX_NOACTIVATE)
	mainWindow, _, _ = procCreateWindowExW.Call(
		exStyle,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(utf16Ptr(appName))),
		WS_POPUP,
		uintptr(x), uintptr(y), overlayWidth, overlayHeight,
		0, 0, hInstance, 0,
	)
	if mainWindow == 0 {
		diagnosticLogf("CreateWindowExW FAILED")
		return
	}
	diagnosticLogf("overlay window created hwnd=0x%X position=(%d,%d) size=%dx%d", mainWindow, x, y, overlayWidth, overlayHeight)
	if hIconBig != 0 {
		procSendMessageW.Call(mainWindow, WM_SETICON, ICON_BIG, hIconBig)
	}
	if hIconSmall != 0 {
		procSendMessageW.Call(mainWindow, WM_SETICON, ICON_SMALL, hIconSmall)
		procSendMessageW.Call(mainWindow, WM_SETICON, ICON_SMALL2, hIconSmall)
	}

	if !registerControllerRawInput(mainWindow) {
		diagnosticLogf("RegisterRawInputDevices FAILED")
		showFatalError("O Windows não permitiu acessar o controle. Reinicie o computador e execute o DualSenseHUB novamente.")
		return
	}
	diagnosticLogf("RegisterRawInputDevices succeeded")
	enumerateRawInputDevicesForLog()
	if timer, _, err := procSetTimer.Call(mainWindow, 1, 50, 0); timer == 0 {
		diagnosticLogf("SetTimer FAILED err=%v", err)
	} else {
		diagnosticLogf("SetTimer ok id=%d", timer)
	}

	var m msg
	for {
		r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) <= 0 {
			diagnosticLogf("GetMessageW ended result=%d", int32(r))
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
}

func installedExecutablePath() (string, error) {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		var err error
		base, err = os.UserConfigDir()
		if err != nil {
			return "", err
		}
	}
	return filepath.Join(base, appName, appName+".exe"), nil
}

func samePath(a, b string) bool {
	aa, errA := filepath.Abs(a)
	bb, errB := filepath.Abs(b)
	if errA != nil || errB != nil {
		return false
	}
	return strings.EqualFold(filepath.Clean(aa), filepath.Clean(bb))
}

func installAndRelaunchIfNeeded() bool {
	current, err := os.Executable()
	if err != nil {
		diagnosticLogf("os.Executable FAILED: %v", err)
		showFatalError("Não foi possível localizar o executável atual.")
		return true
	}
	target, err := installedExecutablePath()
	if err != nil {
		diagnosticLogf("installedExecutablePath FAILED: %v", err)
		showFatalError("Não foi possível localizar a pasta de instalação.")
		return true
	}
	diagnosticLogf("executable current=%s target=%s", current, target)
	if samePath(current, target) {
		diagnosticLogf("already running from installed target")
		return false
	}
	diagnosticLogf("running as installer/updater")

	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		diagnosticLogf("MkdirAll FAILED: %v", err)
		showFatalError("Não foi possível criar a pasta de instalação do DualSenseHUB.")
		return true
	}

	// Copy first. The currently installed executable may still be running, but a separate
	// .new file can be written safely while we ask the old process to shut down.
	temp := target + ".new"
	_ = os.Remove(temp)
	diagnosticLogf("copying installer to temp=%s", temp)
	if err := copyFile(current, temp); err != nil {
		diagnosticLogf("copyFile FAILED: %v", err)
		_ = os.Remove(temp)
		showFatalError("Não foi possível copiar a atualização para a pasta de instalação.")
		return true
	}

	diagnosticLogf("requesting existing instances to close")
	closeExistingInstancesAndWait(8 * time.Second)

	if err := replaceInstalledExecutable(temp, target, 8*time.Second); err != nil {
		diagnosticLogf("replaceInstalledExecutable FAILED: %v", err)
		_ = os.Remove(temp)
		showFatalError("Não foi possível atualizar o DualSenseHUB. Feche o programa no Gerenciador de Tarefas e execute este arquivo novamente.")
		return true
	}

	diagnosticLogf("installed executable replaced successfully")
	command := exec.Command(target)
	command.Dir = filepath.Dir(target)
	command.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: CREATE_NO_WINDOW,
	}
	if err := command.Start(); err != nil {
		diagnosticLogf("starting installed executable FAILED: %v", err)
		showFatalError("A instalação foi concluída, mas o DualSenseHUB não pôde ser iniciado.")
		return true
	}
	diagnosticLogf("started installed executable pid=%d", command.Process.Pid)
	return true
}

func replaceInstalledExecutable(temp, target string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for {
		r, _, callErr := procMoveFileExW.Call(
			uintptr(unsafe.Pointer(utf16Ptr(temp))),
			uintptr(unsafe.Pointer(utf16Ptr(target))),
			MOVEFILE_REPLACE_EXISTING|MOVEFILE_WRITE_THROUGH,
		)
		if r != 0 {
			return nil
		}
		lastErr = callErr
		if time.Now().After(deadline) {
			if lastErr == nil {
				lastErr = fmt.Errorf("MoveFileExW falhou")
			}
			return lastErr
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func showFatalError(message string) {
	procMessageBoxW.Call(
		0,
		uintptr(unsafe.Pointer(utf16Ptr(message))),
		uintptr(unsafe.Pointer(utf16Ptr(appName))),
		MB_OK|MB_ICONERROR,
	)
}

func copyFile(source, destination string) error {
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	syncErr := out.Sync()
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	if syncErr != nil {
		return syncErr
	}
	return closeErr
}

func closeExistingInstancesAndWait(timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	seen := make(map[uint32]struct{})
	var processHandles []uintptr

	// Close both this version and the previous DualSensePower version during migration.
	for _, class := range []string{windowClassName, "DualSensePowerOverlayWindow"} {
		className := utf16Ptr(class)
		for attempt := 0; attempt < 40; attempt++ {
			hwnd, _, _ := procFindWindowW.Call(uintptr(unsafe.Pointer(className)), 0)
			if hwnd == 0 {
				break
			}
			diagnosticLogf("found existing window class=%s hwnd=0x%X", class, hwnd)

			var pid uint32
			procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
			if pid != 0 {
				if _, exists := seen[pid]; !exists {
					seen[pid] = struct{}{}
					handle, _, _ := procOpenProcess.Call(PROCESS_SYNCHRONIZE, 0, uintptr(pid))
					if handle != 0 {
						processHandles = append(processHandles, handle)
					}
				}
			}

			diagnosticLogf("posting WM_CLOSE pid=%d hwnd=0x%X", pid, hwnd)
			procPostMessageW.Call(hwnd, WM_CLOSE, 0, 0)
			time.Sleep(50 * time.Millisecond)
		}
	}

	for _, handle := range processHandles {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			procCloseHandle.Call(handle)
			continue
		}
		milliseconds := uint32(remaining / time.Millisecond)
		if milliseconds == 0 {
			milliseconds = 1
		}
		result, _, _ := procWaitForSingleObject.Call(handle, uintptr(milliseconds))
		_ = result == WAIT_OBJECT_0 || result == WAIT_TIMEOUT
		procCloseHandle.Call(handle)
	}
}

func acquireSingleInstance() bool {
	name := utf16Ptr("Local\\DualSenseHUB.SingleInstance")
	h, _, _ := procCreateMutexW.Call(0, 1, uintptr(unsafe.Pointer(name)))
	if h == 0 {
		return false
	}
	last, _, _ := procGetLastError.Call()
	return uint32(last) != ERROR_ALREADY_EXISTS
}

func removeLegacyStartupEntry() {
	subKey := utf16Ptr(`Software\Microsoft\Windows\CurrentVersion\Run`)
	var key uintptr
	result, _, _ := procRegCreateKeyExW.Call(
		0x80000001, // HKEY_CURRENT_USER
		uintptr(unsafe.Pointer(subKey)),
		0, 0, 0,
		KEY_SET_VALUE,
		0,
		uintptr(unsafe.Pointer(&key)),
		0,
	)
	if result != 0 || key == 0 {
		return
	}
	defer procRegCloseKey.Call(key)
	legacyName := utf16Ptr("DualSensePower")
	procRegDeleteValueW.Call(key, uintptr(unsafe.Pointer(legacyName)))
}

func removeLegacyInstalledCopy() {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		return
	}
	legacyDir := filepath.Join(base, "DualSensePower")
	legacyExe := filepath.Join(legacyDir, "DualSensePower.exe")
	_ = os.Remove(legacyExe)
	_ = os.Remove(legacyDir)
}

func installStartupEntry() {
	exe, err := os.Executable()
	if err != nil {
		diagnosticLogf("installStartupEntry os.Executable FAILED: %v", err)
		return
	}
	exe, err = filepath.Abs(exe)
	if err != nil {
		return
	}

	subKey := utf16Ptr(`Software\Microsoft\Windows\CurrentVersion\Run`)
	var key uintptr
	result, _, _ := procRegCreateKeyExW.Call(
		0x80000001, // HKEY_CURRENT_USER
		uintptr(unsafe.Pointer(subKey)),
		0, 0, 0,
		KEY_SET_VALUE,
		0,
		uintptr(unsafe.Pointer(&key)),
		0,
	)
	if result != 0 || key == 0 {
		return
	}
	defer procRegCloseKey.Call(key)

	valueName := utf16Ptr(appName)
	command := `"` + exe + `"`
	data, _ := syscall.UTF16FromString(command)
	procRegSetValueExW.Call(
		key,
		uintptr(unsafe.Pointer(valueName)),
		0,
		REG_SZ,
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(len(data)*2),
	)
}

func powershellSingleQuoted(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func installStartMenuShortcut() {
	exe, err := os.Executable()
	if err != nil {
		diagnosticLogf("installStartMenuShortcut os.Executable FAILED: %v", err)
		return
	}
	exe, err = filepath.Abs(exe)
	if err != nil {
		diagnosticLogf("installStartMenuShortcut filepath.Abs FAILED: %v", err)
		return
	}

	appData := os.Getenv("APPDATA")
	if appData == "" {
		diagnosticLogf("installStartMenuShortcut APPDATA is empty")
		return
	}

	programsDir := filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs")
	if err := os.MkdirAll(programsDir, 0o755); err != nil {
		diagnosticLogf("installStartMenuShortcut MkdirAll FAILED: %v", err)
		return
	}
	shortcutPath := filepath.Join(programsDir, appName+".lnk")

	// WScript.Shell is available on supported Windows versions and creates a real
	// Start Menu .lnk that keeps working after the downloaded updater is deleted.
	script := strings.Join([]string{
		"$ws = New-Object -ComObject WScript.Shell",
		"$sc = $ws.CreateShortcut(" + powershellSingleQuoted(shortcutPath) + ")",
		"$sc.TargetPath = " + powershellSingleQuoted(exe),
		"$sc.WorkingDirectory = " + powershellSingleQuoted(filepath.Dir(exe)),
		"$sc.IconLocation = " + powershellSingleQuoted(exe+",0"),
		"$sc.Description = 'DualSenseHUB - bateria e desligamento do DualSense'",
		"$sc.Save()",
	}, "; ")

	command := exec.Command(
		"powershell.exe",
		"-NoLogo",
		"-NoProfile",
		"-NonInteractive",
		"-ExecutionPolicy", "Bypass",
		"-WindowStyle", "Hidden",
		"-Command", script,
	)
	command.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: CREATE_NO_WINDOW,
	}
	if output, err := command.CombinedOutput(); err != nil {
		diagnosticLogf("installStartMenuShortcut PowerShell FAILED: %v output=%q", err, strings.TrimSpace(string(output)))
		return
	}

	// Tell Explorer and Task Manager that icon-bearing shell items changed. This
	// prevents them from reusing the old black-background icon cached for the path.
	const (
		SHCNE_ASSOCCHANGED = 0x08000000
		SHCNF_IDLIST       = 0x0000
	)
	procSHChangeNotify.Call(SHCNE_ASSOCCHANGED, SHCNF_IDLIST, 0, 0)
	diagnosticLogf("Start Menu shortcut installed: %s", shortcutPath)
}

func registerControllerRawInput(hwnd uintptr) bool {
	// Diagnostic mode registers the whole Generic Desktop usage page and then
	// filters by Sony VID/PID. This catches controllers whose Bluetooth top-level
	// collection is not exposed specifically as joystick (0x04) or gamepad (0x05).
	devices := []rawInputDevice{
		{UsagePage: 0x01, Usage: 0x00, Flags: RIDEV_PAGEONLY | RIDEV_INPUTSINK | RIDEV_DEVNOTIFY, Target: hwnd},
	}
	r, _, err := procRegisterRawInputDevices.Call(
		uintptr(unsafe.Pointer(&devices[0])),
		uintptr(len(devices)),
		unsafe.Sizeof(rawInputDevice{}),
	)
	diagnosticLogf("RegisterRawInputDevices page=0x01 usage=0x00 flags=0x%X target=0x%X result=%d err=%v structSize=%d", devices[0].Flags, hwnd, r, err, unsafe.Sizeof(rawInputDevice{}))
	return r != 0
}

func enumerateRawInputDevicesForLog() {
	var count uint32
	r, _, err := procGetRawInputDeviceList.Call(0, uintptr(unsafe.Pointer(&count)), unsafe.Sizeof(rawInputDeviceList{}))
	if r == ^uintptr(0) {
		diagnosticLogf("GetRawInputDeviceList count FAILED err=%v", err)
		return
	}
	diagnosticLogf("raw input device count=%d listStructSize=%d", count, unsafe.Sizeof(rawInputDeviceList{}))
	if count == 0 {
		return
	}
	list := make([]rawInputDeviceList, count)
	r, _, err = procGetRawInputDeviceList.Call(uintptr(unsafe.Pointer(&list[0])), uintptr(unsafe.Pointer(&count)), unsafe.Sizeof(rawInputDeviceList{}))
	if r == ^uintptr(0) {
		diagnosticLogf("GetRawInputDeviceList data FAILED err=%v", err)
		return
	}
	for i := 0; i < int(count) && i < len(list); i++ {
		device := list[i].Device
		name := getRawInputDeviceName(device)
		vendor, product, usagePage, usage, ok := getRawInputHIDInfo(device)
		diagnosticLogf("RAW_DEVICE index=%d handle=0x%X type=%d infoOK=%t vid=%04X pid=%04X usagePage=%04X usage=%04X name=%s", i, device, list[i].Type, ok, vendor, product, usagePage, usage, name)
	}
}

func getRawInputHIDInfo(device uintptr) (vendorID, productID, usagePage, usage uint32, ok bool) {
	var info [32]byte
	binary.LittleEndian.PutUint32(info[0:4], uint32(len(info)))
	sz := uint32(len(info))
	r, _, _ := procGetRawInputDeviceInfoW.Call(device, RIDI_DEVICEINFO, uintptr(unsafe.Pointer(&info[0])), uintptr(unsafe.Pointer(&sz)))
	if r == ^uintptr(0) || sz < 24 {
		return 0, 0, 0, 0, false
	}
	return binary.LittleEndian.Uint32(info[8:12]), binary.LittleEndian.Uint32(info[12:16]), uint32(binary.LittleEndian.Uint16(info[20:22])), uint32(binary.LittleEndian.Uint16(info[22:24])), true
}

func getRawInputDeviceName(device uintptr) string {
	var chars uint32
	r, _, _ := procGetRawInputDeviceInfoW.Call(device, RIDI_DEVICENAME, 0, uintptr(unsafe.Pointer(&chars)))
	if r == ^uintptr(0) || chars == 0 {
		return ""
	}
	buffer := make([]uint16, chars+1)
	r, _, _ = procGetRawInputDeviceInfoW.Call(device, RIDI_DEVICENAME, uintptr(unsafe.Pointer(&buffer[0])), uintptr(unsafe.Pointer(&chars)))
	if r == ^uintptr(0) {
		return ""
	}
	return syscall.UTF16ToString(buffer)
}

func wndProc(hwnd uintptr, message uint32, wParam, lParam uintptr) uintptr {
	switch message {
	case WM_INPUT:
		processRawInput(lParam)
		return 0

	case WM_INPUT_DEVICE_CHANGE:
		diagnosticLogf("WM_INPUT_DEVICE_CHANGE wParam=%d device=0x%X", wParam, lParam)
		if wParam == GIDC_REMOVAL {
			delete(deviceCache, lParam)
			if state.lastDevice == lParam {
				state.psDown = false
				state.shutdownTriggered = false
				state.simpleMode = false
				state.batteryKnown = false
				state.charging = false
				state.lowBatteryAlerted = false
				hideOverlay()
			}
		} else if wParam == GIDC_ARRIVAL {
			delete(deviceCache, lParam)
		}
		return 0

	case WM_APP_BATTERY:
		if state.simpleMode {
			applyBatteryReading(int(wParam), false)
		}
		return 0

	case WM_TIMER:
		onTimer()
		return 0

	case WM_DISPLAYCHANGE:
		if state.overlayVisible {
			diagnosticLogf("display changed while overlay visible; re-rendering")
			if state.overlayLowBattery {
				renderLowBatteryOverlay()
			} else {
				renderOverlay(state.batteryPercent, state.charging)
			}
		}
		return 0

	case WM_DESTROY:
		diagnosticLogf("WM_DESTROY received")
		procPostQuitMessage.Call(0)
		return 0
	}

	r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(message), wParam, lParam)
	return r
}

func processRawInput(rawHandle uintptr) {
	headerSize := uint32(unsafe.Sizeof(rawInputHeader{}))
	var sz uint32
	r, _, err := procGetRawInputData.Call(rawHandle, RID_INPUT, 0, uintptr(unsafe.Pointer(&sz)), uintptr(headerSize))
	if r == ^uintptr(0) || sz < headerSize+8 {
		diagnosticLogf("GetRawInputData size FAILED handle=0x%X result=%d size=%d err=%v", rawHandle, r, sz, err)
		return
	}

	buffer := make([]byte, sz)
	read, _, err := procGetRawInputData.Call(rawHandle, RID_INPUT, uintptr(unsafe.Pointer(&buffer[0])), uintptr(unsafe.Pointer(&sz)), uintptr(headerSize))
	if read == ^uintptr(0) || read < uintptr(headerSize+8) {
		diagnosticLogf("GetRawInputData read FAILED handle=0x%X result=%d size=%d err=%v", rawHandle, read, sz, err)
		return
	}

	header := (*rawInputHeader)(unsafe.Pointer(&buffer[0]))
	if header.Type != RIM_TYPEHID {
		return
	}
	if !isDualSenseDevice(header.Device) {
		return
	}

	hs := int(headerSize)
	reportSize := int(binary.LittleEndian.Uint32(buffer[hs : hs+4]))
	reportCount := int(binary.LittleEndian.Uint32(buffer[hs+4 : hs+8]))
	dataStart := hs + 8
	available := int(read)
	if reportSize <= 0 || reportCount <= 0 || dataStart+reportSize*reportCount > available {
		diagnosticLogf("invalid RAWHID layout device=0x%X reportSize=%d reportCount=%d dataStart=%d available=%d totalBuffer=%d", header.Device, reportSize, reportCount, dataStart, available, len(buffer))
		return
	}

	for i := 0; i < reportCount; i++ {
		start := dataStart + i*reportSize
		report := buffer[start : start+reportSize]
		processDualSenseReport(header.Device, report)
	}
}

func isDualSenseDevice(device uintptr) bool {
	if known, ok := deviceCache[device]; ok {
		return known
	}

	vendorID, productID, usagePage, usage, infoOK := getRawInputHIDInfo(device)
	name := getRawInputDeviceName(device)
	upperName := strings.ToUpper(name)
	nameMatches := strings.Contains(upperName, "VID_054C") && (strings.Contains(upperName, "PID_0CE6") || strings.Contains(upperName, "PID_0DF2"))
	isDualSense := (infoOK && vendorID == 0x054C && (productID == 0x0CE6 || productID == 0x0DF2)) || nameMatches
	deviceCache[device] = isDualSense
	diagnosticLogf("device classification handle=0x%X infoOK=%t vid=%04X pid=%04X usagePage=%04X usage=%04X nameMatch=%t dualSense=%t name=%s", device, infoOK, vendorID, productID, usagePage, usage, nameMatches, isDualSense, name)
	return isDualSense
}

func logDiagnosticReport(device uintptr, report []byte, decoded dualSenseReport) {
	reportDiagnosticMu.Lock()
	defer reportDiagnosticMu.Unlock()

	now := time.Now()
	d := reportDiagnostics[device]
	if d == nil {
		d = &reportDiagnosticState{}
		reportDiagnostics[device] = d
	}
	d.count++

	changed := make([]string, 0, 12)
	if len(d.lastReport) == len(report) {
		for i := range report {
			if report[i] != d.lastReport[i] {
				if len(changed) < 12 {
					changed = append(changed, fmt.Sprintf("%d:%02X>%02X", i, d.lastReport[i], report[i]))
				}
			}
		}
	}

	decodedChanged := decoded.Valid != d.lastDecoded.Valid || decoded.Simple != d.lastDecoded.Simple || decoded.PSPressed != d.lastDecoded.PSPressed || decoded.BatteryValid != d.lastDecoded.BatteryValid || decoded.BatteryPercent != d.lastDecoded.BatteryPercent || decoded.Charging != d.lastDecoded.Charging
	shouldLog := d.count <= 5 || decodedChanged || now.Sub(d.lastLogAt) >= 30*time.Second
	if shouldLog {
		maxBytes := len(report)
		if maxBytes > 96 {
			maxBytes = 96
		}
		hexData := strings.ToUpper(hex.EncodeToString(report[:maxBytes]))
		diagnosticLogf("RAW_REPORT device=0x%X n=%d len=%d id=%02X valid=%t simple=%t ps=%t batteryValid=%t battery=%d charging=%t changed=[%s] data=%s", device, d.count, len(report), report[0], decoded.Valid, decoded.Simple, decoded.PSPressed, decoded.BatteryValid, decoded.BatteryPercent, decoded.Charging, strings.Join(changed, ","), hexData)
		d.lastLogAt = now
	}

	d.lastReport = append(d.lastReport[:0], report...)
	d.lastDecoded = decoded
}

func processDualSenseReport(device uintptr, report []byte) {
	decoded := parseDualSenseReport(report)
	logDiagnosticReport(device, report, decoded)
	if !decoded.Valid {
		return
	}

	now := time.Now()
	state.lastReportAt = now
	state.lastDevice = device

	if decoded.Simple {
		// Bluetooth STANDARD mode contains the PS button but no battery byte.
		// Requesting a harmless feature report makes the controller begin sending
		// its complete 0x31 Bluetooth reports. The HID handle is opened shared and
		// closed immediately; the app never controls rumble, lights or triggers.
		requestEnhancedBluetoothMode(device, now)

		// Keep the Windows battery property as a passive fallback while the
		// enhanced report is being enabled. A recent full USB report has priority
		// if both transports briefly coexist while a cable is connected.
		if now.Sub(state.lastFullReportAt) > controllerTimeout {
			if !state.simpleMode {
				state.simpleMode = true
				state.charging = false
				state.lastBatteryPollAt = time.Time{}
			}
			requestSystemBatteryPoll(now)
		}
	} else {
		state.simpleMode = false
		state.lastFullReportAt = now
		if decoded.BatteryValid {
			applyBatteryReading(decoded.BatteryPercent, decoded.Charging)
		}
	}

	handlePSButton(decoded.PSPressed, now)
}

func handlePSButton(psPressed bool, now time.Time) {
	if psPressed && !state.psDown {
		diagnosticLogf("PS DOWN detected")
		state.psDown = true
		state.psDownAt = now
		state.shutdownTriggered = false
		showOverlayFor(psHoldTime + 500*time.Millisecond)
	} else if !psPressed && state.psDown {
		diagnosticLogf("PS UP detected held=%s", now.Sub(state.psDownAt))
		state.psDown = false
	}

	if state.psDown && !state.shutdownTriggered && now.Sub(state.psDownAt) >= psHoldTime {
		diagnosticLogf("PS hold threshold reached elapsed=%s", now.Sub(state.psDownAt))
		triggerShutdown()
	}
}

func applyBatteryReading(percent int, charging bool) {
	percent = clampPercent(percent)
	firstReading := !state.batteryKnown
	previousPercent := state.batteryPercent
	previousCharging := state.charging
	percentChanged := !firstReading && percent != previousPercent
	chargingChanged := !firstReading && charging != previousCharging

	state.batteryKnown = true
	state.batteryPercent = percent
	state.charging = charging

	// Warn only once while the controller remains in the critical 0-9% band,
	// represented by the midpoint value 5%. Charging or rising above that band
	// arms the warning for a future discharge cycle.
	if charging || percent > lowBatteryThreshold {
		state.lowBatteryAlerted = false
	}
	if !charging && percent <= lowBatteryThreshold && !state.lowBatteryAlerted {
		state.lowBatteryAlerted = true
		showLowBatteryAlert()
		return
	}

	// Show the normal overlay when the controller is first detected or when the
	// charging state changes (cable connected/disconnected). Ordinary battery
	// steps such as 45% -> 55% are kept silent and appear on the next PS press.
	if firstReading || chargingChanged {
		showOverlayFor(overlayVisibleTime)
	} else if percentChanged {
		diagnosticLogf("battery percentage updated silently old=%d new=%d charging=%t", previousPercent, percent, charging)
	}

	if state.psDown && !state.shutdownTriggered && !state.overlayVisible {
		// The PS button may have been pressed before the first asynchronous
		// Windows battery lookup completed. Never re-open the overlay after the
		// shutdown action has already been triggered.
		showOverlayFor(psHoldTime + 500*time.Millisecond)
	}
}

func requestSystemBatteryPoll(now time.Time) {
	if !state.simpleMode {
		return
	}
	if !state.lastBatteryPollAt.IsZero() && now.Sub(state.lastBatteryPollAt) < 2*time.Second {
		return
	}
	if !batteryPollInFlight.CompareAndSwap(false, true) {
		return
	}
	state.lastBatteryPollAt = now

	go func() {
		defer batteryPollInFlight.Store(false)
		diagnosticLogf("battery query started")
		percent, ok := queryDualSenseBatteryPercent()
		diagnosticLogf("battery query result ok=%t percent=%d", ok, percent)
		if !ok || mainWindow == 0 {
			return
		}
		procPostMessageW.Call(mainWindow, WM_APP_BATTERY, uintptr(clampPercent(percent)), 0)
	}()
}

func onTimer() {
	now := time.Now()
	reportsStopped := !state.lastReportAt.IsZero() && now.Sub(state.lastReportAt) > controllerTimeout

	// When the controller is powered off while PS is still held, Windows never
	// sends a final PS-UP report. Clear the held state and hide the overlay as
	// soon as the controller reports stop, including after shutdown was already
	// triggered. Otherwise an asynchronous battery update can show the overlay
	// again and leave it stuck on screen indefinitely.
	if reportsStopped {
		if state.psDown {
			diagnosticLogf("PS state cleared because controller reports stopped for %s shutdownTriggered=%t", now.Sub(state.lastReportAt), state.shutdownTriggered)
			state.psDown = false
		}
		if state.overlayVisible {
			diagnosticLogf("overlay hidden because controller reports stopped for %s", now.Sub(state.lastReportAt))
			hideOverlay()
		}
	}

	if state.psDown && !state.shutdownTriggered && now.Sub(state.psDownAt) >= psHoldTime {
		triggerShutdown()
	}

	if state.simpleMode && !reportsStopped {
		requestSystemBatteryPoll(now)
	}

	if state.overlayVisible && !state.psDown && now.After(state.overlayHideAt) {
		hideOverlay()
	}
}

func showOverlayFor(duration time.Duration) {
	if !state.batteryKnown {
		diagnosticLogf("overlay request skipped because battery is unknown")
		return
	}
	diagnosticLogf("show overlay battery=%d charging=%t duration=%s", state.batteryPercent, state.charging, duration)
	state.overlayHideAt = time.Now().Add(duration)
	state.overlayLowBattery = false
	renderOverlay(state.batteryPercent, state.charging)
	state.overlayVisible = true
	procShowWindow.Call(mainWindow, SW_SHOWNOACTIVATE)
}

func showLowBatteryAlert() {
	diagnosticLogf("show low battery alert battery=%d charging=%t", state.batteryPercent, state.charging)
	state.overlayHideAt = time.Now().Add(lowBatteryVisibleTime)
	state.overlayLowBattery = true
	renderLowBatteryOverlay()
	state.overlayVisible = true
	procShowWindow.Call(mainWindow, SW_SHOWNOACTIVATE)
}

func hideOverlay() {
	if state.overlayVisible {
		diagnosticLogf("hide overlay")
	}
	state.overlayVisible = false
	state.overlayLowBattery = false
	procShowWindow.Call(mainWindow, SW_HIDE)
}

func triggerShutdown() {
	diagnosticLogf("triggerShutdown called")
	state.shutdownTriggered = true
	hideOverlay()

	command := exec.Command(
		"shutdown.exe",
		"/s",
		"/t", "10",
		"/c", "Seu computador irá desligar em 10 segundos",
	)
	command.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: CREATE_NO_WINDOW,
	}
	if err := command.Start(); err != nil {
		diagnosticLogf("shutdown.exe start FAILED: %v", err)
	} else {
		diagnosticLogf("shutdown.exe started pid=%d", command.Process.Pid)
	}
}

func overlayPosition(frameWidth int) (int32, int32) {
	w, _, _ := procGetSystemMetrics.Call(SM_CXSCREEN)
	h, _, _ := procGetSystemMetrics.Call(SM_CYSCREEN)
	_ = h
	x := int32(w) - overlayRightMargin - int32(frameWidth)
	if x < 0 {
		x = 0
	}
	return x, overlayTopMargin
}

func renderOverlay(percent int, charging bool) {
	frame, err := loadOverlayFrame(clampPercent(percent), charging)
	if err != nil || frame == nil || mainWindow == 0 {
		diagnosticLogf("renderOverlay aborted percent=%d charging=%t frameNil=%t hwnd=0x%X err=%v", percent, charging, frame == nil, mainWindow, err)
		return
	}
	updateLayeredOverlay(frame, fmt.Sprintf("battery=%d charging=%t", percent, charging))
}

func renderLowBatteryOverlay() {
	frame, err := loadLowBatteryFrame()
	if err != nil || frame == nil || mainWindow == 0 {
		diagnosticLogf("renderLowBatteryOverlay aborted frameNil=%t hwnd=0x%X err=%v", frame == nil, mainWindow, err)
		return
	}
	updateLayeredOverlay(frame, "low-battery")
}

func updateLayeredOverlay(frame *overlayFrame, description string) {
	screenDC, _, _ := procGetDC.Call(0)
	if screenDC == 0 {
		diagnosticLogf("GetDC FAILED")
		return
	}
	defer procReleaseDC.Call(0, screenDC)

	memoryDC, _, _ := procCreateCompatibleDC.Call(screenDC)
	if memoryDC == 0 {
		diagnosticLogf("CreateCompatibleDC FAILED")
		return
	}
	defer procDeleteDC.Call(memoryDC)

	bmi := bitmapInfo{
		Header: bitmapInfoHeader{
			Size:        uint32(unsafe.Sizeof(bitmapInfoHeader{})),
			Width:       int32(frame.width),
			Height:      -int32(frame.height), // top-down DIB
			Planes:      1,
			BitCount:    32,
			Compression: BI_RGB,
			SizeImage:   uint32(len(frame.pixels)),
		},
	}

	var bits unsafe.Pointer
	bitmap, _, _ := procCreateDIBSection.Call(
		memoryDC,
		uintptr(unsafe.Pointer(&bmi)),
		DIB_RGB_COLORS,
		uintptr(unsafe.Pointer(&bits)),
		0,
		0,
	)
	if bitmap == 0 || bits == nil {
		diagnosticLogf("CreateDIBSection FAILED bitmap=0x%X bitsNil=%t", bitmap, bits == nil)
		return
	}
	defer procDeleteObject.Call(bitmap)

	pixelTarget := unsafe.Slice((*byte)(bits), len(frame.pixels))
	copy(pixelTarget, frame.pixels)

	old, _, _ := procSelectObject.Call(memoryDC, bitmap)
	if old != 0 {
		defer procSelectObject.Call(memoryDC, old)
	}

	x, y := overlayPosition(frame.width)
	destination := point{X: x, Y: y}
	source := point{X: 0, Y: 0}
	windowSize := size{CX: int32(frame.width), CY: int32(frame.height)}
	blend := blendFunction{
		BlendOp:             AC_SRC_OVER,
		SourceConstantAlpha: 255,
		AlphaFormat:         AC_SRC_ALPHA,
	}

	result, _, updateErr := procUpdateLayeredWindow.Call(
		mainWindow,
		screenDC,
		uintptr(unsafe.Pointer(&destination)),
		uintptr(unsafe.Pointer(&windowSize)),
		memoryDC,
		uintptr(unsafe.Pointer(&source)),
		0,
		uintptr(unsafe.Pointer(&blend)),
		ULW_ALPHA,
	)
	diagnosticLogf("UpdateLayeredWindow %s result=%d err=%v pos=(%d,%d) size=%dx%d", description, result, updateErr, x, y, frame.width, frame.height)
}

func loadLowBatteryFrame() (*overlayFrame, error) {
	const cacheKey = "low-battery"
	if cached, ok := frameCache[cacheKey]; ok {
		return cached, nil
	}

	f, err := overlayAssets.Open("assets/low_battery.png")
	if err != nil {
		return nil, err
	}
	decoded, _, err := image.Decode(f)
	_ = f.Close()
	if err != nil {
		return nil, err
	}

	bounds := decoded.Bounds()
	canvas := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(canvas, canvas.Bounds(), decoded, bounds.Min, draw.Src)
	frame := nrgbaToOverlayFrame(canvas)
	frameCache[cacheKey] = frame
	return frame, nil
}

func loadOverlayFrame(percent int, charging bool) (*overlayFrame, error) {
	cacheKey := fmt.Sprintf("%03d:%t", percent, charging)
	if cached, ok := frameCache[cacheKey]; ok {
		return cached, nil
	}

	path := fmt.Sprintf("assets/%03d.png", percent)
	f, err := overlayAssets.Open(path)
	if err != nil {
		return nil, err
	}
	decoded, _, err := image.Decode(f)
	_ = f.Close()
	if err != nil {
		return nil, err
	}

	bounds := decoded.Bounds()
	canvas := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(canvas, canvas.Bounds(), decoded, bounds.Min, draw.Src)

	if charging {
		boltFile, boltErr := overlayAssets.Open("assets/charging_bolt.png")
		if boltErr == nil {
			bolt, _, decodeErr := image.Decode(boltFile)
			_ = boltFile.Close()
			if decodeErr == nil {
				// The bolt sits in the original gap between the controller and battery.
				destination := image.Rect(99, 28, 99+bolt.Bounds().Dx(), 28+bolt.Bounds().Dy())
				draw.Draw(canvas, destination, bolt, bolt.Bounds().Min, draw.Over)
			}
		}
	}

	frame := nrgbaToOverlayFrame(canvas)
	frameCache[cacheKey] = frame
	return frame, nil
}

func nrgbaToOverlayFrame(canvas *image.NRGBA) *overlayFrame {
	w, h := canvas.Bounds().Dx(), canvas.Bounds().Dy()
	pixels := make([]byte, w*h*4)
	index := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, b, a := canvas.At(x, y).RGBA()
			// UpdateLayeredWindow expects premultiplied BGRA.
			pixels[index+0] = byte(b >> 8)
			pixels[index+1] = byte(g >> 8)
			pixels[index+2] = byte(r >> 8)
			pixels[index+3] = byte(a >> 8)
			index += 4
		}
	}
	return &overlayFrame{width: w, height: h, pixels: pixels}
}

func init() {
	// Keep the compiler from dropping filepath support in aggressive link modes.
	_ = filepath.Separator
}
