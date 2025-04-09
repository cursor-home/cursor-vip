package main

import (
	"os"        // 操作系统功能包
	"os/signal" // 信号处理包
	"syscall"   // 系统调用包

	"github.com/gofrs/flock" // 用于文件锁定，确保程序只有一个实例运行

	"github.com/kingparks/cursor-vip/auth"         // 认证相关功能包
	"github.com/kingparks/cursor-vip/tui"          // 终端用户界面包
	"github.com/kingparks/cursor-vip/tui/params"   // 参数配置包
	"github.com/kingparks/cursor-vip/tui/shortcut" // 快捷键处理包
	"github.com/kingparks/cursor-vip/tui/tool"     // 工具函数包
)

// lock 文件锁变量，用于确保应用程序只有一个实例在运行
var lock *flock.Flock

// pidFilePath 进程ID文件路径，用于记录当前运行的实例
var pidFilePath string

// main 主函数，程序入口点
func main() {
	// 确保应用程序只有一个实例运行
	lock, pidFilePath, _ = tool.EnsureSingleInstance("cursor-vip")
	// 运行终端用户界面，获取用户选择的产品和模型索引
	productSelected, modelIndexSelected := tui.Run()
	// 启动服务器，处理用户选择
	startServer(productSelected, modelIndexSelected)
}

// startServer 启动服务器函数
// productSelected: 用户选择的产品
// modelIndexSelected: 用户选择的模型索引
func startServer(productSelected string, modelIndexSelected int) {
	// 创建信号通道，用于接收系统信号
	params.Sigs = make(chan os.Signal, 1)
	// 注册需要监听的系统信号
	signal.Notify(params.Sigs, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGKILL)
	// 启动goroutine处理接收到的信号
	go func() {
		<-params.Sigs                     // 阻塞等待接收信号
		_ = lock.Unlock()                 // 释放文件锁
		_ = os.Remove(pidFilePath)        // 删除PID文件
		auth.UnSetClient(productSelected) // 注销客户端
		os.Exit(0)                        // 正常退出程序
	}()
	// 启动快捷键处理goroutine
	go shortcut.Do()
	// 运行认证流程
	auth.Run(productSelected, modelIndexSelected)
}
