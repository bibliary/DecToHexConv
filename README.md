# Decimal to Hexadecimal Converter

A Go desktop app that converts a decimal number to an 8-digit hexadecimal value. It was built for RFID tag UIDs: paste the decimal UID from a reader to get the hex value.

Current version: **1.0.0**. The window shows the version, the author, and a link to this repository. To stamp a different number onto a build:

```powershell
go build -ldflags="-H windowsgui -X main.appVersion=1.0.0" -o decToHexConv.exe
```

## Description

The window has two fields:

- **Decimal number:** type a decimal value. The hexadecimal field updates as you type.
- **Hexadecimal number:** shows the result. Press Enter to keep that result, clear the decimal field, and add the conversion to the list below.

The list keeps the last 5 conversions, newest first.

## Requirements

- Go 1.21 or newer
- A C compiler. Fyne uses CGO. On Windows that compiler is MinGW-w64 (`gcc`)

## Install the C compiler on Windows

Check whether it is already available:

```powershell
gcc --version
go env CGO_ENABLED
```

`CGO_ENABLED` should be `1`. If it is `0`, turn it on for the current terminal:

```powershell
$env:CGO_ENABLED = "1"
```

If `gcc` is missing, install [MSYS2](https://www.msys2.org/), then in the MSYS2 UCRT64 terminal run:

```bash
pacman -S mingw-w64-ucrt-x86_64-gcc
```

Add `C:\msys64\ucrt64\bin` to your user `PATH`, open a new PowerShell, and run `gcc --version` again.

Another option, if Chocolatey is already installed:

```powershell
choco install mingw
```

## Go dependencies

```bash
go mod download
```

## Build and run

From the project directory, rebuild without a console window:

```powershell
go build -ldflags="-H windowsgui" -o decToHexConv.exe
```

This build embeds the icon. `rsrc.syso` is the Windows file icon, and `icon.ico` is compiled into the program for the window icon. Those files do not need to sit next to the finished exe.

Close the converter if it is open, then start `decToHexConv.exe` again.

Run without building:

```bash
go run main.go
```

A plain build opens a console window together with the converter. Closing that console closes the app:

```bash
go build -o decToHexConv.exe
```

Use `-ldflags="-H windowsgui"` when you do not want that console.

### Verbose build

For build problems, `-v` lists packages as they compile and `-x` prints the commands, including the C compiler:

```bash
go build -v -x -ldflags="-H windowsgui" -o decToHexConv.exe
```

`-a` rebuilds every package, including ones already built. In PowerShell:

```powershell
$env:CGO_ENABLED = "1"
go build -v -x -a -ldflags="-H windowsgui" -o decToHexConv.exe
```

### Linux and macOS

```bash
go run main.go
```

Or build and run:

```bash
go build -o decToHexConv
./decToHexConv
```

On Linux, install the usual Fyne system packages (`gcc`, and the X11/OpenGL development headers) before building.

## Usage

1. Start the app.
2. Type a decimal number. The hexadecimal field updates while you type.
3. Press Enter. The hexadecimal value stays, the decimal field is cleared, and the conversion is added under **Recent conversions**.
4. Positive numbers are formatted as 8 uppercase hex digits. Negative numbers keep a leading minus.
5. Values that do not fit in 8 hex digits show an error and are not added to the history.

## Examples

- Input: `255` → Output: `000000FF`
- Input: `16` → Output: `00000010`
- Input: `-42` → Output: `-0000002A`
- Input: `4294967295` → Output: `FFFFFFFF`
- Input: `1` → Output: `00000001`

## Technologies

- [Fyne](https://fyne.io/) — cross-platform GUI toolkit for Go
