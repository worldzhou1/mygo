//go:build windows

package d3d11

import (
	"sync"
	"syscall"
	"unsafe"
)

// osVersionInfo mirrors RtlGetVersion's RTL_OSVERSIONINFOW, which reports the
// real version whatever the manifest says (GetVersionEx lies since 8.1).
type osVersionInfo struct {
	size        uint32
	major       uint32
	minor       uint32
	build       uint32
	platform    uint32
	servicePack [128]uint16
}

var (
	rtlGetVersion = syscall.NewLazyDLL("ntdll.dll").NewProc("RtlGetVersion")
	win7Once      sync.Once
	win7          bool
)

// windows7 reports whether this is Windows 7 (or Server 2008 R2), where the
// swap chain has to be the bitblt model: the flip model there flickers, and
// the frame latency object came later still.
func windows7() bool {
	win7Once.Do(func() {
		v := osVersionInfo{size: uint32(unsafe.Sizeof(osVersionInfo{}))}
		if r, _, _ := rtlGetVersion.Call(uintptr(unsafe.Pointer(&v))); r != 0 {
			return
		}
		win7 = v.major == 6 && v.minor == 1
	})
	return win7
}
