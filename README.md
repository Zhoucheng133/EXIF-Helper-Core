# EXIF Helper Core

![License](https://img.shields.io/badge/License-MIT-dark_green)

这是EXIF Helper的核心组件（动态库），如果要查看使用方法见仓库[EXIF Helper](https://github.com/Zhoucheng133/EXIF-Helper)

## 生成动态库

```bash
go mod tidy
# Windows系统
go build -o build/image.dll -buildmode=c-shared .
# macOS系统
go build -o build/image.dylib -buildmode=c-shared .
```