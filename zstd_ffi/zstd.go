//go:build cgo
// +build cgo

package zstd_ffi

/*
	#cgo CFLAGS: -I${SRCDIR}/include
	#cgo LDFLAGS: -lkernel32 -lntdll -luserenv -lws2_32 -ldbghelp -L${SRCDIR}/bin -lzstd
	#include <stdlib.h>
	#include <zstd_interface.h>
*/
import "C"
import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"unsafe"
)

func init() {
	// 动态库最终路径
	var libFile string
	switch runtime.GOOS {
	case "windows":
		libFile = "bin/zstd.dll"
	case "darwin":
		libFile = "bin/libzstd.dylib"
	default:
		libFile = "bin/libzstd.so"
	}

	// 如果库不存在，则编译 Rust 并复制到 bin/
	if _, err := os.Stat(libFile); os.IsNotExist(err) {
		// Rust 源码目录（Cargo.toml 所在目录）
		rustDir := "../" // 根据你的目录结构调整
		buildCmd := exec.Command("cargo", "build", "--release")
		buildCmd.Dir = rustDir
		buildCmd.Stdout = os.Stdout
		buildCmd.Stderr = os.Stderr
		if err := buildCmd.Run(); err != nil {
			panic("Failed to build Rust library: " + err.Error())
		}

		// 源文件路径（默认 target/release/）
		var srcLib string
		switch runtime.GOOS {
		case "windows":
			srcLib = filepath.Join(rustDir, "target", "release", "zstd.dll")
		case "darwin":
			srcLib = filepath.Join(rustDir, "target", "release", "libzstd.dylib")
		default:
			srcLib = filepath.Join(rustDir, "target", "release", "libzstd.so")
		}

		// 确保 bin 目录存在
		_ = os.MkdirAll("bin", 0755)

		// 复制库到 bin/
		input, err := os.ReadFile(srcLib)
		if err != nil {
			panic("Failed to read Rust library: " + err.Error())
		}
		if err := os.WriteFile(libFile, input, 0644); err != nil {
			panic("Failed to write library to bin/: " + err.Error())
		}
	}
}

// Compress compresses data with zstd level (typical -7..22).
// It queries required size first, then allocates and fetches result.
func Compress(data []byte, level int) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("zstd compress: empty input")
	}

	var required C.size_t
	// Query required size: pass out_ptr=NULL, out_capacity=0
	rc := C.zstd_compress(
		(*C.uchar)(unsafe.Pointer(&data[0])),
		C.size_t(len(data)),
		C.int(level),
		nil,
		0,
		&required,
	)
	if rc != 2 {
		if rc == 0 {
			// extremely unlikely: compressed to zero bytes
			return []byte{}, nil
		}
		return nil, errors.New("zstd compress query failed")
	}

	outPtr := C.malloc(required)
	if outPtr == nil {
		return nil, errors.New("zstd compress: malloc failed")
	}
	defer C.free(outPtr)

	var outLen C.size_t
	rc = C.zstd_compress(
		(*C.uchar)(unsafe.Pointer(&data[0])),
		C.size_t(len(data)),
		C.int(level),
		(*C.uchar)(outPtr),
		required,
		&outLen,
	)
	if rc != 0 {
		return nil, errors.New("zstd compress failed")
	}

	// copy to Go slice
	out := C.GoBytes(outPtr, C.int(outLen))
	return out, nil
}

// Decompress decompresses zstd data.
// It queries required size first, then allocates and fetches result.
func Decompress(z []byte) ([]byte, error) {
	if len(z) == 0 {
		return nil, errors.New("zstd decompress: empty input")
	}

	var required C.size_t
	rc := C.zstd_decompress(
		(*C.uchar)(unsafe.Pointer(&z[0])),
		C.size_t(len(z)),
		nil,
		0,
		&required,
	)
	if rc != 2 {
		if rc == 0 {
			return []byte{}, nil
		}
		return nil, errors.New("zstd decompress query failed")
	}

	outPtr := C.malloc(required)
	if outPtr == nil {
		return nil, errors.New("zstd decompress: malloc failed")
	}
	defer C.free(outPtr)

	var outLen C.size_t
	rc = C.zstd_decompress(
		(*C.uchar)(unsafe.Pointer(&z[0])),
		C.size_t(len(z)),
		(*C.uchar)(outPtr),
		required,
		&outLen,
	)
	if rc != 0 {
		return nil, errors.New("zstd decompress failed")
	}

	out := C.GoBytes(outPtr, C.int(outLen))
	return out, nil
}
