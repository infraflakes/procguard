package app

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Re-defined structures with syscall types
type ICONINFO struct {
	FIcon    int32
	XHotspot uint32
	YHotspot uint32
	HbmMask  syscall.Handle
	HbmColor syscall.Handle
}

type BITMAPINFOHEADER struct {
	BiSize          uint32
	BiWidth         int32
	BiHeight        int32
	BiPlanes        uint16
	BiBitCount      uint16
	BiCompression   uint32
	BiSizeImage     uint32
	BiXPelsPerMeter int32
	BiYPelsPerMeter int32
	BiClrUsed       uint32
	BiClrImportant  uint32
}

type RGBQUAD struct {
	RgbBlue     byte
	RgbGreen    byte
	RgbRed      byte
	RgbReserved byte
}

type BITMAPINFO struct {
	BmiHeader BITMAPINFOHEADER
	BmiColors [1]RGBQUAD
}

var (
	modShell32 = windows.NewLazySystemDLL("shell32.dll")
	modUser32  = windows.NewLazySystemDLL("user32.dll")
	modGdi32   = windows.NewLazySystemDLL("gdi32.dll")

	procExtractIconExW     = modShell32.NewProc("ExtractIconExW")
	procDestroyIcon        = modUser32.NewProc("DestroyIcon")
	procGetIconInfo        = modUser32.NewProc("GetIconInfo")
	procGetSystemMetrics   = modUser32.NewProc("GetSystemMetrics")
	procGetDC              = modUser32.NewProc("GetDC")
	procReleaseDC          = modUser32.NewProc("ReleaseDC")
	procCreateCompatibleDC = modGdi32.NewProc("CreateCompatibleDC")

	procGetDIBits    = modGdi32.NewProc("GetDIBits")
	procDeleteObject = modGdi32.NewProc("DeleteObject")
	procDeleteDC     = modGdi32.NewProc("DeleteDC")
)

const (
	SM_CXICON      = 11
	SM_CYICON      = 12
	DIB_RGB_COLORS = 0
	BI_RGB         = 0
)

func ExtractIcons(exePath string, nIcons uint) ([]syscall.Handle, []syscall.Handle, error) {
	pExePath, err := syscall.UTF16PtrFromString(exePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert path to UTF16: %w", err)
	}

	numIcons, _, _ := procExtractIconExW.Call(uintptr(unsafe.Pointer(pExePath)), uintptr(0xFFFFFFFF), 0, 0, 0)
	if numIcons == 0 {
		return nil, nil, fmt.Errorf("no icons found in %s", exePath)
	}

	largeIcons := make([]syscall.Handle, numIcons)
	smallIcons := make([]syscall.Handle, numIcons)

	ret, _, err := procExtractIconExW.Call(
		uintptr(unsafe.Pointer(pExePath)),
		0,
		uintptr(unsafe.Pointer(&largeIcons[0])),
		uintptr(unsafe.Pointer(&smallIcons[0])),
		uintptr(numIcons),
	)
	if ret == 0 {
		return nil, nil, fmt.Errorf("ExtractIconExW failed: %w", err)
	}

	return largeIcons, smallIcons, nil
}

func GetSystemMetrics(nIndex int) int32 {
	ret, _, _ := procGetSystemMetrics.Call(uintptr(nIndex))
	return int32(ret)
}

func HICONToImage(hIcon syscall.Handle) (image.Image, error) {
	var iconInfo ICONINFO
	ret, _, err := procGetIconInfo.Call(uintptr(hIcon), uintptr(unsafe.Pointer(&iconInfo)))
	if ret == 0 {
		return nil, fmt.Errorf("GetIconInfo failed: %w", err)
	}
	defer func() {
		_, _, _ = procDeleteObject.Call(uintptr(iconInfo.HbmColor))
		_, _, _ = procDeleteObject.Call(uintptr(iconInfo.HbmMask))
	}()

	width := GetSystemMetrics(SM_CXICON)
	height := GetSystemMetrics(SM_CYICON)

	screenDC, _, _ := procGetDC.Call(0)
	defer func() { _, _, _ = procReleaseDC.Call(0, screenDC) }()

	memDC, _, _ := procCreateCompatibleDC.Call(screenDC)
	defer func() { _, _, _ = procDeleteDC.Call(memDC) }()

	var bmiColor BITMAPINFO
	bmiColor.BmiHeader.BiSize = uint32(unsafe.Sizeof(bmiColor.BmiHeader))
	bmiColor.BmiHeader.BiWidth = width
	bmiColor.BmiHeader.BiHeight = -height
	bmiColor.BmiHeader.BiPlanes = 1
	bmiColor.BmiHeader.BiBitCount = 32
	bmiColor.BmiHeader.BiCompression = BI_RGB

	colorData := make([]byte, width*height*4)

	ret, _, err = procGetDIBits.Call(
		memDC,
		uintptr(iconInfo.HbmColor),
		0,
		uintptr(height),
		uintptr(unsafe.Pointer(&colorData[0])),
		uintptr(unsafe.Pointer(&bmiColor)),
		DIB_RGB_COLORS,
	)
	if ret == 0 {
		return nil, fmt.Errorf("GetDIBits (color data) failed: %w", err)
	}

	img := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
	for y := 0; y < int(height); y++ {
		for x := 0; x < int(width); x++ {
			offset := (y*int(width) + x) * 4
			img.SetRGBA(x, y, color.RGBA{R: colorData[offset+2], G: colorData[offset+1], B: colorData[offset], A: colorData[offset+3]})
		}
	}

	return img, nil
}

func imageToBase64(img image.Image) (string, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func GetAppIconAsBase64(exePath string) (string, error) {
	largeIcons, _, err := ExtractIcons(exePath, 1)
	if err != nil {
		return "", err
	}
	if len(largeIcons) == 0 {
		return "", fmt.Errorf("no icons found")
	}
	defer func() { _, _, _ = procDestroyIcon.Call(uintptr(largeIcons[0])) }()

	img, err := HICONToImage(largeIcons[0])
	if err != nil {
		return "", err
	}

	return imageToBase64(img)
}
