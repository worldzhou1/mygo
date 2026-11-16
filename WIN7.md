# mygo on Windows 7

This fork is upstream `github.com/egoist/mygo` at **v0.8.3** (the version
`indulink-client-mygo` pins) with four small changes that let the native UI
run on Windows 7 SP1. One commit, `win7: make the Direct3D 11 swap chain
without the frame latency object` (+68/-3).

## What had to change

| File | Change | Why |
|---|---|---|
| `internal/gpu/d3d11/d3d11.go` | retry `CreateSwapChainForHwnd` without `DXGI_SWAP_CHAIN_FLAG_FRAME_LATENCY_WAITABLE_OBJECT` (0x40) when it fails, and resize with the same flags | 0x40 came with Windows 8.1; Win7 + Platform Update refuses the whole call with `DXGI_ERROR_INVALID_CALL`. Measured on a Win7 SP1 guest: `flags=0x40` → `0x887A0001`, `flags=0` → `S_OK` |
| `internal/gpu/d3d11/d3d11.go` | `if part != nil && r.ctx1 != 0` before `ID3D11DeviceContext1::ClearView` | Win7 without KB2670838 has no D3D11.1 context; the old path called a nil interface |
| `internal/windows/app.go` | `SetProcessDPIAware()` fallback | `SetProcessDpiAwarenessContext` is Windows 10; without a fallback the process is DPI-unaware and Windows bitmap-scales the whole window at 125% |
| `internal/windows/win32.go` | `dpiOf` falls back to `GetDeviceCaps(LOGPIXELSX)` | neither `GetDpiForWindow` nor `GetDpiForSystem` exists on Win7, so it always said 96 |

Everything else degrades on its own: DLLs are loaded lazily and every
Windows 8/10 API is behind `has(proc)`, DirectComposition is simply absent
(transparent windows and flyout shadows become plain windows), and a renderer
that cannot be made at all falls back to the in-memory rasterizer + GDI.

## Requirements on the target machine

1. Windows 7 **SP1**.
2. **KB2670838** (Platform Update for Windows 7): DXGI 1.2, Direct3D 11.1.
   Without it mygo logs `d3d11: DXGI 1.2 is required` and draws in memory.
3. A WDDM display driver. A VM without VMware Tools reports
   `D3D11CreateDevice` → `DXGI_ERROR_UNSUPPORTED` and only exposes WARP.
4. Build with `thongtech/go-legacy-win7` (`GOTOOLCHAIN=local`): the official
   Go toolchain writes PE minimum OS version 10.0 and Windows 7 refuses to
   load the exe at all. The patched toolchain writes 6.1.

## Build

```sh
export GOROOT=/path/to/go-legacy-win7
export PATH=$GOROOT/bin:$PATH GOTOOLCHAIN=local
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" ./examples/counter-native
```

Use it from the client:

```
replace github.com/egoist/mygo => /Users/zhoushijie/Documents/Vscode/go/mygo-win7
```

or build with a separate modfile, so the app's `go.mod` stays untouched:

```sh
cp go.mod go.win7.mod && cp go.sum go.win7.sum
printf '\nreplace github.com/egoist/mygo => /Users/zhoushijie/Documents/Vscode/go/mygo-win7\n' >> go.win7.mod
GOOS=windows go build -modfile=go.win7.mod -o client-win7.exe .
```
