package main

import (
	"fmt"
	"github.com/xbfding/xslog"
	"log/slog"
)

func main() {
	// 创建日志器
	logger, err := xslog.NewLogger(xslog.LogConfig{
		LogToConsole:    true,
		LogToFile:       true,
		LogFilePath:     "logs/app.log",
		LevelForConsole: slog.LevelInfo,  // 控制台初始级别为 Info
		LevelForFile:    slog.LevelDebug, // 文件初始级别为 Debug
	})
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	// 记录一些日志
	logger.Debug("这是调试信息") // 只会写入文件，不会显示在控制台
	logger.Info("这是普通信息")  // 同时写入文件和显示在控制台

	// 动态调整控制台级别（例如在开发环境调试时）
	logger.SetConsoleLevel(slog.LevelDebug)
	logger.Debug("现在控制台也会显示调试信息了")

	// 动态调整文件级别（例如在生产环境中减少日志量）
	logger.SetFileLevel(slog.LevelWarn)
	logger.Info("这条信息只会显示在控制台，不会写入文件")
	logger.Warn("这条警告信息会同时显示在控制台和写入文件")

	// 临时禁用控制台输出（例如在运行批处理任务时）
	logger.EnableConsole(false)
	logger.Error("这条错误只会写入文件，不会显示在控制台")

	// 重新启用控制台输出
	logger.EnableConsole(true)

	// 更改日志文件（例如按日期轮转日志）
	err = logger.ChangeFilePath("logs/app.log.1")
	if err != nil {
		fmt.Println("更改日志文件失败:", err)
	}
}
