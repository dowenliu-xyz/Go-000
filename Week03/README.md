学习笔记
---
[homework](homework.md)

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

# sync 包和其他同步工具

## Share Memory By Communicating

Go 鼓励使用 chan 在 goroutine 之间传递对数据的引用。

> https://blog.golang.org/codelab-share

**Do not communicate by share memory, instead, share memory by communicating.** 

## Detecting Race Conditions With Go

Race detector:

* `go build -race` 在线上环境如果不是查问题，不建议使用，对性能有影响
* go test -race

[demo11](cmd/demo11/demo11.go) 使用 `-race` 标记编译和运行，可以得到 Data Race 的警告输出。

> 写入单个 machine word 将是原子的，但 interface 内部是是两个 machine word 的值：
> 一个类型指针+一个值指针。对interface的赋值不是原子操作。
> 另一个goroutine 可能在更改接口值时观察到它的内容。
> [demo12](cmd/demo12/demo12.go)

> **不要凭直觉判断一个值是原子值**，最好使用并发原语。**Don't be clever.**

> 用好锁：最晚加锁、最早释放。锁里/临界区的代码要轻量，越短越好、越简单越好。
> 可以不用放到临界区的代码不要放到临界区。

> 加锁时要注意顺序，防止死锁。活跃性问题。
> 死锁条件：互斥、占用且等待、不可强行占有、循环等待条件。
> 最简单可行的避免死锁策略是破坏循环等待条件，如按序加锁。其次破坏占用且等待条件，如超时回避。
> 互斥、不可强行占有这两个条件通常是不可破坏的，否则锁就没有意义了。

没有安全的 data race(safe data race)。您的程序要么没有 data race，要么其操作未定义。
**Don't be clever.** 请使用 chan、 Mutex 或 atomic

## sync.aotmic

* atomic.Store()

bad case demo: [config.go](cmd/demo13/config.go)
使用同步原语 [config_test.go](cmd/demo13/config_test.go)

> 这个场景读写相当，RWMutex 的效率比 Mutex 差的多。

> Benchmark 是出结果真相的真理

go test -bench=.

Mutex 相对更重。因为涉及到更多的 goroutine 之间的上下文切换 pack blocking goroutine，
以及唤醒 goroutine。

* Copy on Write  
  在微服务降级或者 local cache 场景中经常使用。
  写时复制指的是，写操作时候复制全量老数据到一个新的对象中，携带上本次新写的数据，
  之后利用原子替换(atomic.Value)，更新调用者的变量。来完成**无锁访问**共享数据。
    * Redis bgsave。利用操作系统 COW 机制。
    * 微服务降级。降级数据缓存的更新。在复制出的新副本中更新，再CAS换成工作副本。
      进程内缓存，定期后台更新。
    * 动态更新配置 local cache。
  > map 是原子赋值。但尽量还是使用 atomic 操作。

* Mutex  
  演进： 1.8及之前，非公平，有饥饿问题；1.9之后解决了饥饿问题  
  Mutex 的三种模式：
  1. Barging 为吞吐量最大化设计的。
     当锁被释放时，会唤醒第一个等待者，然后把锁给第一个请求锁的人。
     可能第一个请求锁的人不是第一个等待者，可能导致饥饿。非公平。
  2. Handsoff。当锁释放时，锁会一直持有直到第一个等待者准备好获取锁。
     公平，但降低了吞吐量。这种模式迫使释放锁的 goroutine 等待等待锁的 goroutine 获得锁。
  3. Spinning。请求锁而不得的 goroutine 在进入等待队列前先自旋，
     期望在接下来有限次CPU时间片执行期间获得锁。
     自旋在等待队列为空或者应用程序重度使用锁是效果不错。
     Parking 和 Unparking goroutines 有不低的性能成本开销，相比自旋来说要慢得多。
  
  1.8之前，Go使用了 Barging 和 Spinning 的结合实现。
  Go 1.9 通过添加一个新的饥饿模式来解决先前的 goroutine 饥饿问题，
  该模式将会在释放时候触发 handsoff：
  所有等待锁超过一毫秒的 goroutine(也称为有界等待)将被诊断为饥饿。
  当被标记为饥饿状态时，unlock 方法会 handsoff 把锁直接扔给第一个等待者。
  在饥饿模式下，自旋也被停用，因为传入的goroutines 将没有机会获取为下一个等待者保留的锁。

## errgroup



我们把一个复杂的任务，尤其是依赖多个微服务 rpc 需要聚合数据的任务，分解为依赖和并行。
依赖的意思为: 需要上游 a 的数据才能访问下游 b 的数据进行组合。
但是并行的意思为: 分解为多个小任务并行执行，最终等全部执行完毕。

* golang.org/x/sync/errgroup
  > https://pkg.go.dev/golang.org/x/sync/errgroup
  >
  > 核心原理: 利用 sync.Waitgroup 管理并行执行的 goroutine
  >
  * 并行工作流
  * 错误处理 或者 优雅降级
  * context 传播和取消
  * 利用局部变量+闭包

* github.com/go-kratos/kratos/pkg/sync/errgroup
  > https://github.com/go-kratos/kratos/tree/master/pkg/sync/errgroup
  >
  x/sync/errgroup的问题：
  * Go(fun()error) 方法只启动了 goroutine 异步处理，但没有做 recover 兜底。
  * 创建 goroutine 数量没有限制，允许启动大量 goroutine
  * WithContext() 方法的返回值中 context.Context 是一个 cancelContext ，容易被误用。
    很容易将返回值中的 context 赋值给参数 context 变量或变量名覆盖，
    然后将这个 context 传递给其他函数使用。 
    一旦 errgroup 取消，使用此 context 的其他操作会大量报错

## sync.Pool

sync.Pool 的场景是用来保存和复用临时对象，以减少内存分配，降低 GC 压力(Request-Driven 特别合适)。

Get 返回 Pool 中的任意一个对象。如果 Pool 为空，则调用 New 返回一个新创建的对象。
放进 Pool 中的对象，会在说不准什么时候被回收掉。
所以如果事先 Put 进去 100 个对象，下次 Get 的时候发现 Pool 是空也是有可能的。
不过这个特性的一个好处就在于不用担心 Pool 会一直增长，因为 Go 已经帮你在 Pool 中做了回收机制。
这个清理过程是在每次垃圾回收之前做的。
之前每次GC 时都会清空 pool，而在1.13版本中引入了 victim cache，
会将 pool 内数据拷贝一份，避免 GC 将其清空，即使没有引用的内容也可以保留最多两轮 GC。

> 不要将数据库连接这样的资源型对象放到 sync.Pool中！

# context

通过 context 可实现传递数据、超时控制、级联取消

如何将context集成到API：
* 显示传递，尽管会污染API
  The first parameter of a function call。首参数传递 context 对象。
  比如，参考  net 包 Dialer.DialContext。
  此函数执行正常的 Dial 操作，但可以通过 context 对象取消函数调用。
* Optional config on a request structure (不推荐)
  在第一个 request 对象中携带一个可选的 context 对象。
  例如 net/http 库的 Request.WithContext，通过携带给定的 context 对象，
  返回一个新的 Request 对象

## Do not store Contexts inside a struct type
Do not store Contexts inside a struct type;
instead, pass a Context explicitly to each function that needs it.
The Context should be the first parameter, typically named ctx:

Incoming requests to a server should create a Context.

使用 context 的一个很好的心智模型是它应该在程序中流动，应该贯穿你的代码。
这通常意味着您不希望将其存储在结构体之中。它从一个函数传递到另一个函数，并根据需要进行扩展。
理想情况下，每个请求都会创建一个 context 对象，并在请求结束时过期。

不存储上下文的一个例外是，当您需要将它放入一个结构中时，该结构纯粹用作通过通道传递的消息。

## context.WithValue

context.WithValue 每次都返回一个新对象，该对象包含 key value 和对父层 context 的引用。
context.Value 查找是递归的向上层找 key/value，而不是使用一个 map 来存储 key/value，
这种结构保证了 context 中传递数据的并发安全性，因为这些 key/value 是只读的。

Use context values only for request-scoped data that transits processes and APIs,
not for passing optional parameters to functions. 比如染色、API重要性、Trace
> https://github.com/go-kratos/kratos/blob/master/pkg/net/metadata/key.go
> 

Context.Value should inform, not control.
context.WithValue 中携带的数据必须是（对接收context的函数/方法）是安全的。
函数/方法不能通过假定需要的参数来自 context.Value 方法。
Context.Value 的数据更多的是面向请求的原数据，不应该作为函数/方法的可选参数来使用。
比如通过 context 传递一个 sql.Tx 对象到 Dao 层使用。
元数据相对函数参数是更加隐含的、面向请求的。而参数是更加显式的。

同一个 context 对象可以传递给不同的 goroutine 中运行的函数，
所以 context value 应该是不可变的，以保证 goroutine 可以安全的使用。
如果 context value 是可变的，如 map, 每次需要变更 context 中的值都应该使用
context.WithValue 函数。
> https://pkg.go.dev/google.golang.org/grpc/metadata

对于 value 是 map 的 context，更新 map 中的 k/v 要使用 copy on write 方式，
以新 map 使用 context.WithValue 创建新的 context。

## context.WithTimeout / context.WithDeadline / context.WithCancel
> context.Context.Deadline() ，计算超时，配置网络请求超时取消。
> 参见 kratos pkg/cache/redis/util.go shrinkDeadline

* When a Context is canceled, all Contexts derived from it are also canceled.

Done() 返回 一个 chan，当我们取消某个parent context, 
实际上会递归层层 cancel 掉自己的 child context 的 done chan
从而让整个调用链中所有监听 cancel 的 goroutine退出。[demo17](cmd/demo17/demo17.go)

> WithCancel / WithTimeout / WithDeadline 会返回 cancel 函数，
> 一定要使用 defer 保证 cancel 被调用，防止 goroutine 泄露。

* All blocking/long operations should be cancelable

[demo18](cmd/demo18/demo18.go)

## Final Notes
* Incoming requests to a server should create a Context.  
  一般建议使用超时机制。Root 为 Background.
* Outgoing calls to servers should accept a Context.  
  调用外部服务一定要显示传递 Context。RPC/http calls，DB queries，etc.
* Do not store Contexts inside a struct type; instead, pass a Context explicitly to
  each function that needs it.
* The chain of function calls between them must propagate the Context.  
  Context 要在函数调用链间传播。
* Replace a Context using WithCancel, WithDeadline, WithTimeout, WithValue.  
  注意，context value 的结构中允许修改字段也不要修改，
  如果要修改，COW+WithValue 创建新的 Context。
* When a Context is canceled, call Contexts derived from it are also canceled.
* The same Context may be passed to functions running in different goroutines;
  Contexts are safe for simultaneous use by multiple goroutines.
* Do not pass a nil Context, even if a function permits it. Pass a TODO context if
  you are unsure about which Context to use.
* Use context values only for request-scoped data that transits processes and APIs,
  not for passing optional parameters to functions.  
  业务逻辑参数要显式传递，不要放到 Context 里。
* All blocking/long operations should be cancelable.
  > net.Conn.SetDeadline()
* Context.Value obscures your program's flow.
  Context value 不应该影响你的应用的业务逻辑。
* Context.Value should inform, not control.
* Try not to use context.Value.

> https://talks.golang.org/2014/gotham-context.slide#1

# Channels

## Unbuffered Channels
`ch := make(chan struct{})`

无缓冲 chan 没有容量，因此进行任何交换前需要两个 goroutine 同时准备好。
当 goroutine 试图将一个资源发送到一个无缓冲的通道并且没有goroutine 等待接收该资源时，
该通道将锁住发送 goroutine 并使其等待。
当 goroutine 尝试从无缓冲通道接收，并且没有 goroutine 等待发送资源时，
该通道将锁住接收 goroutine 并使其等待。

**无缓冲信道的本质是保证同步。**

[demo19](cmd/demo19/demo19.go)

* Receive 先于 Send 完成。
* 好处：100% 保证能收到。
* 代价：延迟时间未知

## Buffered Channels

buffered channel 具有容量，因此其行为可能有点不同。
当 goroutine 试图将资源发送到缓冲通道，
而该通道已满时， 该通道将锁住 goroutine并使其等待缓冲区可用；
如果通道中有空间，发送可以立即进行，goroutine 可以继续。
当goroutine 试图从缓冲通道接收数据，而缓冲通道为空时，
该通道将锁住 goroutine 并使其等待资源被发送。

> 在 chan 创建过程中定义的缓冲区大小可能会极大地影响性能。
> chan 锁住和解锁 goroutine 时，goroutine parking/unparking 会比较耗时，
> goroutine 上下文切换消耗比较多。

* Send 先于 Receive 发生。
* 好处: 延迟更小。
* 代价: 不保证数据到达，越大的 buffer，越小的保障到达。buffer = 1 时，给你延迟一个消息的保障。

## Design Philosophy 设计理念

* If any given Send on a channel CAN cause the sending goroutine to block:  
  当向一个 channel 发送时，允许发送 goroutine 阻塞：
  * Not allowed to use a Buffered channel larger than 1.  
    不要使用缓冲区大小超过1的有缓冲 channel
    * Buffers larger than 1 must have reason/measurements.
      设置缓冲区大小1必须要有充分的理由或压测。
  * Must know what happens when the sending goroutine blocks.  
    必须明确知道发送 goroutine 阻塞的原因。

* If any given Send on a channel WON'T cause the sending goroutine to block:  
  当向一个 channel 发送时，不想让发送 goroutine 阻塞：
  * You have the exact number of buffers for each send.  
    如果你要发送的内容数量与缓冲区大小刚好匹配。
    * Fan Out pattern 使用扇出模式。
  * You have the buffer measured for max capacity.  
    如果你要发送的内容数量超出 channel 的最大容量
    * Drop pattern 放弃部分数据，使用丢弃模式
      ```
      select {
      case ch<-data:
      default:
      }
      ```

* Less is more with buffers. 通道缓冲越少越好。
  * Don't think about performance when thinking about buffers.  
    缓冲区大小与性能无关
  * Buffers can help to reduce blocking latency between signaling.  
    缓冲区大小只与阻塞延迟有关
    * Reducing blocking latency towards zero does not necessarily mean better throughput.  
      阻塞延迟与吞吐量无关。吞吐与消费 channel 的 goroutine 数量有关。
    * If a buffer of one is giving you good enough throughput then keep it.
    * Question buffers that are larger than one and measure for size.
      想要将缓冲区大小设置为超过1时，不要想当然，要通过压测确定缓冲区大小。
    * Find the smallest buffer possible that provides good enough throughput.  
      在保证足够好的吞吐量的前提下，缓冲区大小要尽量小。

# Go Concurrency Patterns

* Timing out 超时处理
  ```
  select {
  case data1<-ch1:
      // if recieved
  case ch2<-data2:
      // if sent
  case <-time.After(duration):
      // if timeout
  ```
* Moving on 放弃数据
  * Drop pattern
* Pipeline
* Fan-out, Fan-in
* Cancellation
  * Close 等于 Receive 发生（类似 Buffered）。
  * 不需要传递数据，或者传递 nil
  * 非常适合去做超时控制
* Context

> https://blog.golang.org/concurrency-timeouts  
> https://blog.golang.org/pipelines  
> https://talks.golang.org/2013/advconc.slide#1  
> https://github.com/go-kratos/kratos/tree/master/pkg/sync

> **一定由Sender关闭channel**
