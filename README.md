# EXIF Helper Core

![License](https://img.shields.io/badge/License-MIT-dark_green)

这是EXIF Helper的核心组件（动态库），如果要查看使用方法见仓库[EXIF Helper](https://github.com/Zhoucheng133/EXIF-Helper)

## 生成动态库

```bash
go mod tidy

#  macOS
go build -buildmode=c-shared -ldflags="-s -w" -o build/image.dylib
# Windows
go build -buildmode=c-shared -ldflags="-s -w" -o build/image.dll
# iOS
chmod +x build_ios.sh
./build_ios.sh
# Android (macOS)
export NDK_PATH=/path/to/your/android-ndk
export CC_PATH=$NDK_PATH/toolchains/llvm/prebuilt/darwin-x86_64/bin/aarch64-linux-android30-clang

CGO_ENABLED=1 \
GOOS=android \
GOARCH=arm64 \
CC=$CC_PATH \
go build -buildmode=c-shared -o build/image.so
# Android (Windows ARM64)
$env:NDK_PATH=/path/to/your/android-ndk

$env:CC="$env:NDK_PATH\toolchains\llvm\prebuilt\windows-x86_64\bin\aarch64-linux-android30-clang.cmd"

$env:CGO_ENABLED="1"
$env:GOOS="android"
$env:GOARCH="arm64"

go build -buildmode=c-shared -o build/image.so
```