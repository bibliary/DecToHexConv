package main

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

//go:embed icon.ico
var iconICO []byte

const historyLimit = 5

func main() {
	myApp := app.New()
	if icon := embeddedIcon(); icon != nil {
		myApp.SetIcon(icon)
	}
	myWindow := myApp.NewWindow("DecToHexConv (8 digits)")
	myWindow.Resize(fyne.NewSize(350, 360))
	myWindow.CenterOnScreen()

	decimalEntry := widget.NewEntry()
	decimalEntry.SetPlaceHolder("Enter a decimal number, then press Enter")

	// Keep the result in an entry so the text stays at full contrast.
	// OnChanged puts the previous value back, so the field cannot be edited.
	// ScrollNone removes the scrollbar that otherwise sits on the right edge.
	hexEntry := widget.NewEntry()
	hexEntry.SetPlaceHolder("Hexadecimal value will appear here...")
	hexEntry.TextStyle = fyne.TextStyle{Monospace: true}
	hexEntry.Wrapping = fyne.TextWrapOff
	hexEntry.Scroll = container.ScrollNone

	var currentHexValue string
	hexEntry.OnChanged = func(text string) {
		if text != currentHexValue {
			hexEntry.SetText(currentHexValue)
		}
	}

	setHex := func(value string) {
		currentHexValue = value
		hexEntry.SetText(value)
	}

	history := make([]string, 0, historyLimit)
	historyLabel := widget.NewLabel("No conversions yet")
	historyLabel.TextStyle = fyne.TextStyle{Monospace: true}

	pushHistory := func(decimal, hex string) {
		line := decimal + " -> " + hex
		history = append([]string{line}, history...)
		if len(history) > historyLimit {
			history = history[:historyLimit]
		}
		historyLabel.SetText(strings.Join(history, "\n"))
	}

	// Set while the decimal field is cleared after a successful conversion.
	// Without it, the empty field would wipe the hex result.
	ignoreDecimalChange := false

	showConversion := func(text string) {
		text = strings.TrimSpace(text)
		if text == "" {
			setHex("")
			return
		}

		hex, err := decimalToHex8(text)
		if err != nil {
			setHex("Error: " + err.Error())
			return
		}
		setHex(hex)
	}

	// Update the hex field while the number is being typed.
	decimalEntry.OnChanged = func(text string) {
		if ignoreDecimalChange {
			return
		}
		showConversion(text)
	}

	// Enter keeps the hex result, records it, and clears the decimal field.
	decimalEntry.OnSubmitted = func(text string) {
		text = strings.TrimSpace(text)
		if text == "" {
			return
		}

		hex, err := decimalToHex8(text)
		if err != nil {
			setHex("Error: " + err.Error())
			return
		}

		setHex(hex)
		pushHistory(text, hex)

		ignoreDecimalChange = true
		decimalEntry.SetText("")
		ignoreDecimalChange = false
	}

	content := container.NewVBox(
		widget.NewLabel("Decimal number:"),
		decimalEntry,
		widget.NewLabel("Hexadecimal number:"),
		hexEntry,
		widget.NewLabel("Recent conversions:"),
		historyLabel,
	)

	myWindow.SetContent(container.NewPadded(content))
	myWindow.ShowAndRun()
}

// decimalToHex8 converts a decimal string to exactly 8 uppercase hex digits.
// Negative values keep a leading minus. Values outside the unsigned 32-bit
// range (or signed 32-bit range, for negatives) are rejected.
func decimalToHex8(text string) (string, error) {
	if strings.HasPrefix(text, "-") {
		num, err := strconv.ParseInt(text, 10, 32)
		if err != nil {
			return "", parseNumberError(err)
		}
		if num < 0 {
			// num is int64, so negating MinInt32 does not overflow.
			return fmt.Sprintf("-%08X", -num), nil
		}
		return fmt.Sprintf("%08X", num), nil
	}

	raw := strings.TrimPrefix(text, "+")
	if raw == "" {
		return "", errors.New("invalid number format")
	}

	num, err := strconv.ParseUint(raw, 10, 32)
	if err != nil {
		return "", parseNumberError(err)
	}
	return fmt.Sprintf("%08X", num), nil
}

func parseNumberError(err error) error {
	if errors.Is(err, strconv.ErrRange) {
		return errors.New("number does not fit in 8 hex digits")
	}
	return errors.New("invalid number format")
}

// embeddedIcon converts the compiled-in icon.ico into a PNG resource.
func embeddedIcon() fyne.Resource {
	pngBytes, err := largestICOToPNG(iconICO)
	if err != nil {
		return nil
	}
	return fyne.NewStaticResource("icon.png", pngBytes)
}

// largestICOToPNG decodes the largest 32-bit image stored in an ICO file.
func largestICOToPNG(data []byte) ([]byte, error) {
	if len(data) < 6 || binary.LittleEndian.Uint16(data[2:4]) != 1 {
		return nil, errors.New("invalid icon")
	}

	count := int(binary.LittleEndian.Uint16(data[4:6]))
	if count == 0 || len(data) < 6+count*16 {
		return nil, errors.New("invalid icon")
	}

	bestOffset, bestSize, bestWidth := 0, 0, 0
	for i := 0; i < count; i++ {
		entry := 6 + i*16
		width := int(data[entry])
		if width == 0 {
			width = 256
		}
		size := int(binary.LittleEndian.Uint32(data[entry+8 : entry+12]))
		offset := int(binary.LittleEndian.Uint32(data[entry+12 : entry+16]))
		if width < bestWidth || size <= 0 || offset < 0 || offset+size > len(data) {
			continue
		}
		bestWidth = width
		bestOffset = offset
		bestSize = size
	}

	img, err := decodeIconDIB(data[bestOffset : bestOffset+bestSize])
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func decodeIconDIB(dib []byte) (image.Image, error) {
	if len(dib) < 40 || binary.LittleEndian.Uint32(dib[0:4]) != 40 {
		return nil, errors.New("unsupported icon image")
	}

	width := int(int32(binary.LittleEndian.Uint32(dib[4:8])))
	height := int(int32(binary.LittleEndian.Uint32(dib[8:12])))
	if width <= 0 || height <= 0 || binary.LittleEndian.Uint16(dib[14:16]) != 32 {
		return nil, errors.New("unsupported icon image")
	}
	// The DIB height counts the color bitmap and the 1-bit mask.
	if height >= width*2 {
		height /= 2
	}

	stride := width * 4
	if len(dib) < 40+stride*height {
		return nil, errors.New("truncated icon image")
	}

	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		row := 40 + (height-1-y)*stride
		for x := 0; x < width; x++ {
			pixel := row + x*4
			img.SetNRGBA(x, y, color.NRGBA{
				R: dib[pixel+2],
				G: dib[pixel+1],
				B: dib[pixel],
				A: dib[pixel+3],
			})
		}
	}
	return img, nil
}
