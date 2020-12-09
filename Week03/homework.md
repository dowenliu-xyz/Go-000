课后作业

---

## Question
基于 errgroup 实现一个 http server 的启动和关闭 ，
以及 linux signal 信号的注册和处理，
要保证能够一个退出，全部注销退出。

## 题目分析
* 保证能够一个退出，全部注销退出  
  应该是指同时启动 App 和 Debug 两个端口服务，App 处理主流程请求、Debug 用于诊断调试。  
  * 通过 WithCancel 的 errgroup.Group 可以分别启动 app 和 debug 服务，
    并在任一服务出错时结束 Wait、关闭内部的 Context。
  * 通过接收 Group 内部 Context 的 Done channel，就可以获得通知来关闭 App 、Debug 服务，
    就可以实现全部注销。
  * 因为 group 是 WithCancel 的，只要有一个服务，wait 就结束了，
    可能各服务的关闭过程不能及时完成，再加一个缓冲区大小为2的 channel，用于等待服务完成关闭。
* linux signal 信号的注册和处理  
  这里只关注 `Ctrl-C` 和 `kill` 信号，分别是 SIGINT 和 SIGTERM。使用 `signal` 包就可实现。  
  为了在捕捉到 signal 信号后能关闭服务，需要构建 errgroup.Group 的 Context 本身是可取消的。
* 模拟服务启动失败，可以指定一个负数端口号触发。
* 其他退出应用的方式还有 `/shutdown` debug 端点。  
  实现方式为将可取消的根 Context 的 cancel 方法传递给 debug Server，
  在 `/shutdown` 端点注册执行该 cancel 方法即可

## 实现

代码： [server.go](cmd/homework/server.go)

说明：

* 支持启动参数 `-appAddr` 和 `-debugAddr`，分别用于设定服务地址。可通过指定非法地址触发启动错误

```shell
$ go run github.com/dowenliu-xyz/Go-000/Week03/cmd/homework -appAddr=':-1'
2020/12/09 19:32:21 Pid: 88244
2020/12/09 19:32:21 Starting app server at :-1
2020/12/09 19:32:21 Shutting down debug server...
2020/12/09 19:32:21 Starting debug server at 127.0.0.1:8081
2020/12/09 19:32:21 Shutting down app server...
2020/12/09 19:32:21 App server stopped.
2020/12/09 19:32:21 Debug server stopped.
2020/12/09 19:32:21 Shutdown: listen tcp: address -1: invalid port
2020/12/09 19:32:21 All servers stopped.
```

* 启动后可通过 `Ctrl-C` 关闭服务
```shell
$ go run github.com/dowenliu-xyz/Go-000/Week03/cmd/homework               
2020/12/09 19:34:06 Pid: 88290
2020/12/09 19:34:06 Starting app server at :8080
2020/12/09 19:34:06 Starting debug server at 127.0.0.1:8081
^C2020/12/09 19:34:08 Recieved signal: interrupt
2020/12/09 19:34:08 Shutting down debug server...
2020/12/09 19:34:08 Shutting down app server...
2020/12/09 19:34:09 App server stopped.
2020/12/09 19:34:09 Debug server stopped.
2020/12/09 19:34:09 Shutdown: http: Server closed
2020/12/09 19:34:09 All servers stopped.
```

* 启动后会输出 `pid` ，可通过 `kill` 关闭服务
```shell
$ go run github.com/dowenliu-xyz/Go-000/Week03/cmd/homework
2020/12/09 19:35:08 Pid: 88341
2020/12/09 19:35:08 Starting app server at :8080
2020/12/09 19:35:08 Starting debug server at 127.0.0.1:8081
2020/12/09 19:35:27 Recieved signal: terminated
2020/12/09 19:35:27 Shutting down debug server...
2020/12/09 19:35:27 Shutting down app server...
2020/12/09 19:35:28 App server stopped.
2020/12/09 19:35:28 Debug server stopped.
2020/12/09 19:35:28 Shutdown: http: Server closed
2020/12/09 19:35:28 All servers stopped.
```

```shell
$ kill 88341
```

* 启动后可通过 `curl` debug port `/shutdown` 关闭服务
```shell
$ go run github.com/dowenliu-xyz/Go-000/Week03/cmd/homework
2020/12/09 19:36:56 Pid: 88799
2020/12/09 19:36:56 Starting app server at :8080
2020/12/09 19:36:56 Starting debug server at 127.0.0.1:8081
2020/12/09 19:37:02 Shutdown request received, shutting down...
2020/12/09 19:37:02 Shutting down app server...
2020/12/09 19:37:02 Shutting down debug server...
2020/12/09 19:37:03 Debug server stopped.
2020/12/09 19:37:03 App server stopped.
2020/12/09 19:37:03 Shutdown: http: Server closed
2020/12/09 19:37:03 All servers stopped.
```

```shell
$ curl http://127.0.0.1:8081/shutdown
Shutting down...
```