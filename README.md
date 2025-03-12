# xslog
补充slog的一些功能，实现可控制是否输入到文件或控制台。控制台输出根据等级标记颜色。

## 调用方法参考
```go
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
		LogFilePath:  "app.log",
	}

	logger, err := xslog.NewLogger(config)
	if err != nil {
		fmt.Println("Failed to create logger:", err)
		return
	}

	logger.Info("开始扫描!", slog.Group("data", slog.String("id", "123123")))
	logger.Error("This is an error message", "error", "something went wrong")
}
```
