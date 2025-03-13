package main

import (
	"fmt"
	"github.com/xbfding/xslog"
	"log/slog"
)

func main() {
	config := xslog.LogConfig{
		LogToConsole: true,
		LogToFile:    true,
		LogFilePath:  "1/2/app.log",
	}

	logger, err := xslog.NewLogger(config)
	if err != nil {
		fmt.Println("Failed to create logger:", err)
		return
	}

	logger.Info("开始扫描!", slog.Group("data", slog.String("id", "123123")))
	logger.Error("This is an error message", "error", "something went wrong")
	logger.Warn("This is an error message", "error", "something went wrong")
	logger.Debug("This is an error message", "error", "something went wrong")
}
