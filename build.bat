@echo off
REM Windows 构建脚本

echo 开始编译 MySQL 数据库同步工具...

go build -o db-sync.exe cmd\db-sync\main.go

if %ERRORLEVEL% EQU 0 (
    echo 编译成功！
    echo 可执行文件: db-sync.exe
    echo.
    echo 使用方法:
    echo   db-sync.exe -config config.yaml
    echo.
) else (
    echo 编译失败！
    exit /b 1
)

