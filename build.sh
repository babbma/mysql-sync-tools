#!/bin/bash

# Linux/Mac 构建脚本

echo "开始编译 MySQL 数据库同步工具..."

go build -o db-sync cmd/db-sync/main.go

if [ $? -eq 0 ]; then
    echo "编译成功！"
    echo "可执行文件: db-sync"
    echo ""
    echo "使用方法:"
    echo "  ./db-sync -config config.yaml"
    echo ""
    chmod +x db-sync
else
    echo "编译失败！"
    exit 1
fi

