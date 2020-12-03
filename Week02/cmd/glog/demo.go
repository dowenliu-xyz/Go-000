package main

import (
	"flag"
	"github.com/dowenliu-xyz/Go-000/Week02/cmd/glog/foo"
	"github.com/golang/glog"
)

// 运行前 `mkdir log`
// 执行参数 `-log_dir=log -alsologtostderr -v=4 -vmodule=bar=8 -log_backtrace_at=bar.go:6`

// glog 使用 demo
func main() {
	flag.Parse()
	defer glog.Flush()

	// log levels

	glog.Info("This is info message")
	glog.Infof("This is info message: %v", 12345)
	glog.InfoDepth(1, "This is info message", 12345)

	glog.Warning("This is waring message")
	glog.Warningf("This is warning message: %v", 12345)
	glog.WarningDepth(1, "This is warning message", 12345)

	glog.Error("This is error message")
	glog.Errorf("This is error message: %v", 12345)
	glog.ErrorDepth(1, "This is error message", 12345)

	// 致命错误，会打印堆栈并退出进程(os.Exit(255))！！！。
	// 与 panic 类似，业务处理中绝对不要用。
	//glog.Fatal("This is fatal message")
	//glog.Fatalf("This is fatal message: %v", 12345)
	//glog.FatalDepth(1, "This is fatal message", 12345)

	// vmodule, for debug

	// 只有 level <= -v 指定的 level 时才会输出。
	// 运行参数中 -v=4 ，所以以下代码中只有前两行有输出
	glog.V(3).Info("LEVEL 3 message")
	glog.V(4).Info("LEVEL 4 message")
	glog.V(5).Info("LEVEL 5 message")
	glog.V(8).Info("LEVEL 8 message")

	// 注意 funcInBar() 函数在文件 bar.go 中。
	// 即使 bar.go 来自其他包，只要文件名是  bar，-vmodule=bar=8都会生效
	// 参数 -vmodule=bar=8 中 bar 指 bar.go
	funcInBar()
	foo.FuncInBar()

	// 与-vmodule 同样的，-log_backtrace_at参数也是对文件名生效。

}
