学习笔记
---

# 管住 Goroutine 的生命周期

> Concurrency is not Parallelism.

## Keep yourself busy or do the work yourself.

> main goroutine 结束，程序退出。

* ❌ `go` 一个 goroutine 去 `ListenAndServe`，main 使用 `select{}` 阻塞。

main goroutine 会阻塞，无法处理别的事情，即使 `ListenAndServe` 的 goroutine 出了错，
它也不会得知，也无法处理，两个 goroutine 之间缺少通讯机制。

* main goroutine 自己来执行 `ListenAndServe`。

消除了将结果从 goroutine 返回到其启动器所需的大量状态跟踪和 chan 操作。

> ❌ `log.Fatal()` 底层会调用 `os.Exit()`，会导致 `defer`失效，应用直接退出！

但 main goroutine 会阻塞在 `ListenAndServe` ，无法处理更多的事情。

## Never start a goroutine without knowing when it will stop

当启动一个 goroutine 时，要明确两个问题：
* **它什么时候会结束（terminate）？**
* **它要怎样结束，要达到什么样的条件，怎么让它退出？
What could prevent it from terminating?**

### 案例1. 控制 http 服务退出  
尝试在两个不同的端口上提供 http 流量：8080 用于应用程序流量；8081 用于访问 /debug/pprof 端点。  
示例 [demo2](cmd/demo02/demo2.go) 问题在于
    * 启动的 goroutine 是否成功、出错，主 goroutine 完全无法得知，
    * 主 goroutine 也因用于监听服务阻塞，没有能力处理其他事务。

> 让 main 函数流程简洁
  先将业务服务监听和 debug 监听分解为独立的函数，由 main 函数调用。
  [demo3](cmd/demo03/demo3.go) [demo4](cmd/demo04/demo4.go)

> **Only use log.Fatal from main.main or init functions**

> 在调用者处显示使用 go 调用一个函数，而不是在调用的函数内使用 go。
  明确直接的告知别人启动了一个 goroutine。

我们期望使用一种方式，同时启动业务端口和 debug 端口，如果任一监听服务出错，应用都退出。  
通过 done、stop 两个 channel 实现。 [demo5](cmd/demo05/demo5.go)  
> 如果再有一个 goroutine 可以向 stop 传入一个 struct{}，就可以控制整个进程平滑停止。

参考：[go-workgroup](https://github.com/da440dil/go-workgroup)

### 案例2 小心 goroutine 泄漏。
buggy example:
```go
package demo

import "fmt"

// leak is a buggy function. It launches a goroutine that
// blocks receiving from a channel. Nothing will ever be
// send on that channel and the channel is never closed so
// that goroutine will be blocked forever.
func leak() {
    ch := make(chan int)

    go func() {
        val := <-ch
        fmt.Println("We received a value:", val)
    }()
}
```

### 案例3 对异步调用要做超时控制
对于某些应用程序，顺序调用产生的延迟可能是不可接受的。

使用 `context.WithTimeout()` 实现超时控制 [demo6](cmd/demo06/demo6.go)

### 案例4 Incomplete Work

[demo7](cmd/demo07/demo7.go) go 后不管版。问题：
1 不能确定 goroutine 会阻塞多久；
2 http Server 退出时，异步 goroutine 可能还没执行完，有数据丢失;
3 每个请求启动一个 goroutine ，不推荐

[demo8](cmd/demo08/demo8.go) goroutine 结束管控版。
解决了 goroutine 结束时间未知的问题，保证数据不丢失，
但仍然创建大量的 goroutine, 也没有限制关闭服务等待的时长，可能很久很久都不会关闭。

[demo9](cmd/demo09/demo9.go) 管控+超时限制版（使用 context)。
解决了 demo8 的超时问题，但仍然创建大量的 goroutine 来处理任务，代价高。

[demo10](cmd/demo10/demo10.go) 毛老师版。
个人疑问：
  * Tracker 内部的 chan 会限制并行上报的数量，所以 Event() 方法应该异步调用吧？不能阻塞请求主流程
  * Run() 方法只有一个 goroutine 处理上报，但可能有大量的 Request 导致上报，
    处理能力不对等，应该使用一个 goroutine pool 吧。
    如果有多个 goroutine 来 Run() 以提高处理能力，stop chan 就不适合了，应该换成 sync.WaitGroup


## Leave concurrency to the caller

```go
package demo

// ListDirectory returns the contents of dir.
func ListDirectory(dir string) ([]string, error)

// ListDirectory returns a channel over which
// directory entries will be published. When the list
// of entries is exhausted, the channel will be closed.
func ListDirectory(dir string) chan string
```
这两个API：
* 将目录读取到一个 slice 中，然后返回整个切片，或者如果出现错误，则返回错误。
  这是同步调用的，ListDirectory 的调用方**会阻塞，直到读取所有目录条目**。
  根据目录的大小，这**可能需要很长时间**，并且**可能会分配大量内存**来构建目录条目名称的 slice。
* ListDirectory 返回一个 chan string，将通过该 chan 传递目录。
  当通道关闭时，这表示不再有目录。
  由于在 ListDirectory 返回后发生通道的填充，ListDirectory
  可能内部启动 goroutine 来填充通道。
  这个版本有两个问题：
  * 通过使用一个关闭的通道作为不再需要处理的项目的信号，
    ListDirectory 无法告诉调用者通过通道返回的项目集不完整，因为中途遇到了错误。
    调用方无法区分空目录与完全从目录读取的错误之间的区别。
    这两种方法（读完或出错）都会导致从 ListDirectory 返回的通道会立即关闭。
  * 调用者必须持续从通道读取，直到它关闭，
    因为这是调用者知道开始填充通道的 goroutine 已经停止的唯一方法。
    这对 ListDirectory 的使用是一个严重的限制，调用者必须花时间从通道读取数据，
    即使它可能已经收到了它想要的答案。
    对于大中型目录，它可能在内存使用方面更为高效，但这种方法并不比原始的基于 slice 的方法快。

更好的 API：
```go
package demo

func ListDirectory(dir string, fn func(string))
```

`filepath.Walk`也是类似的模型。
如果函数启动 goroutine，则必须向调用方提供显式停止该goroutine 的方法。
通常，将异步执行函数的决定权交给该函数的调用方通常更容易。

# Memory Model GO内存模型

> https://golang.org/ref/mem
> 
> https://www.jianshu.com/p/5e44168f47a3

为了串行化访问，请使用 channel 或其他同步原语，例如 sync 和 sync/atomic 来保护数据。
**Don't be clever.**
> Go 中没有 Java\C++ 中的 volatile 原语。要保证可见性，请使用锁/原子操作/channel

**如果没有同步原语保证，并发环境中什么状态都可能发生**，反直觉、反逻辑。

## 并发问题原因
* 指令重排，为了提高读写内存的效率。CPU重排/内存重排；编译重排。
  > 多线程环境下无法轻易断定两段代码是"等价"的。
* 多核心CPU架构、多级CPU缓存结构导致变量变更的可见性问题。
  > store buffer 对单核心是完美的。

  对于多线程的程序，所有的 CPU 都会提供“锁”支持，称之为 barrier，或者 fence。
  它要求：barrier 指令要求所有对内存的操作都必须要“扩散”到 memory
  之后才能继续执行其他对 memory 的操作。
  因此，我们可以用高级点的 atomic compare-and-swap，或者直接用更高级的锁，通常是标准库提供。

## Happens Before
### 定义
为了说明读和写的必要条件，我们定义了先行发生(Happens Before)。
如果事件 e1 发生在 e2 前，我们可以说 e2 发生在 e1 后。
如果 e1不发生在 e2 前也不发生在 e2 后，我们就说 e1 和 e2 是并发的。
* 在单一的独立的 goroutine 中先行发生的顺序即是程序中表达的顺序。
  > 编译重排和内存重排不会破坏单一 goroutine 中逻辑的正确性
* 当下面条件满足时，对变量 v 的读操作 r 是被允许看到对 v 的写操作 w 的：
  1. r 不先行发生于 w
  2. 在 w 后 r 前没有对 v 的其他写操作
* 为了保证对变量 v 的读操作 r 看到对 v 的写操作 w，要确保 w 是 r 允许看到的唯一写操作。
  即当下面条件满足时，r 被保证看到 w：
  1. w 先行发生于 r
  2. 其他对共享变量 v 的写操作要么在 w 前，要么在 r 后。
  > 这一对条件比前面的条件更严格，需要没有其他写操作与 w 或 r 并发发生。

### 实现方式
* 单个 goroutine 中没有并发，所以上面两个定义是相同的：
  读操作 r 看到最近一次的写操作 w 写入 v 的值。
* 当多个 goroutine 访问共享变量 v 时，它们必须使用同步事件（channel/atomic/locks）来建立
  Happens Before 这一条件来保证读操作能看到需要的写操作。
  * 对变量 v 的零值初始化在内存模型中表现的与写操作相同。
  * 原子赋值：对大于 single machine word 的变量的读写操作表现的像以不确定顺序对多个
    single machine word 的变量的操作。
    > 不要自信认为某些结构是 single machine word：
      slice/interface 不是 single machine word。
      map 是 single machine word，但不一定哪一天 Go 底层实现修改了就不是了。
    >
    > 另外要注意，single machine word 的操作只是保证原子，但不影响可见性。
      保证可见性还是需要使用原语
