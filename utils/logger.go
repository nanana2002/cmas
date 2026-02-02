package utils

import (
    "fmt"
    "time"
)

// Logger 全局日志记录器
var Logger *LogManager

// LogManager 日志管理器结构体
type LogManager struct{}

// InitLogger 初始化日志系统
func InitLogger() {
    Logger = &LogManager{}
}

// Log 通用日志函数（支持格式化）
func Log(module, level, format string, v ...interface{}) {
    timestamp := time.Now().Format("2006-01-02 15:04:05")
    msg := fmt.Sprintf(format, v...)
    fmt.Printf("[%s] [%s] [%s] %s\n", timestamp, module, level, msg)
}

// Info 信息日志（支持格式化）
func (lm *LogManager) Info(module, format string, v ...interface{}) {
    Log(module, "INFO", format, v...)
}

// Error 错误日志（支持格式化）
func (lm *LogManager) Error(module, format string, v ...interface{}) {
    Log(module, "ERROR", format, v...)
}

// Debug 调试日志（支持格式化）
func (lm *LogManager) Debug(module, format string, v ...interface{}) {
    Log(module, "DEBUG", format, v...)
}

// Warn 警告日志（新增：补充缺失的Warn级别）
func (lm *LogManager) Warn(module, format string, v ...interface{}) {
    Log(module, "WARN", format, v...)
}

// 全局函数兼容旧代码
func Info(module, format string, v ...interface{}) {
    Log(module, "INFO", format, v...)
}

func Error(module, format string, v ...interface{}) {
    Log(module, "ERROR", format, v...)
}

func Debug(module, format string, v ...interface{}) {
    Log(module, "DEBUG", format, v...)
}

func Warn(module, format string, v ...interface{}) {
    Log(module, "WARN", format, v...)
}