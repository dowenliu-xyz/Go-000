学习笔记
---
[课后作业说明](./homework.md)

> 技巧：看一手资料，优先看官方文档，不要看非官方的翻译

# Error 错误处理

## Error vs Exception Error的设计思路

Go `error` 就是普通的一个接口。常使用 `errors.New()`来返回一个 `error` 对象。  
基础库中大量自定义的 `error`。 包级。**哨兵`error`**。  
`errors.New()` 返回的是内部 `errorString` 对象的指针。

> 建议使用 `errors.New()` 遵循以下规范：以`[package]: `开头

> Q: 为什么标准库 `errors.New()` 要返回指针。  
> A: 防止同字面值创建的 `error` 被判断相等。
> `errors`包中`errorString`是值类型，同样内容的 `errorString` 变量会被判断为相等，
> 但不同的 `errorString` 变量的地址不同，
> 所以要返回指针值，保证同样的值（也就是地址）不会被外部创建出来。参考
> [demo](./cmd/why_errors_new_ptr/demo.go)

### 其他语言的演进历史：
* C  
单返回值，一般通过传递指针作为入参，返回值为 int 表示成功还是失败。
* C++
引入了 exception，但是无法知道被调用方会抛出什么异常。
* Java
引入了 checked exception。  
但Java的异常太过常见，为了编码上的方便，有很多没有正确处理：  
  * catch and ignore
  * catch and rethrow as unchecked exception
  * just throws unchecked exception / error

### Go error
Go 的处理异常逻辑是不引入 exception，而是使用多返回值的方式，在函数签名中带上 `error`，
交由调用者来判定。（？也有可能像 Java 中一样被错误处理，ignore）

> 如果一个函数返回了 value, error，你不能对这个 value 做任何假设，必须先判定 error。
> 唯一可以忽略 error 的是，如果你连 value 也不关心。

Go 中有 panic 机制，但它和其他语言的 exception 机制是**完全不一样**的。  
Go panic 意味着 fatal error(就是挂了)。不能假设调用者来解决 panic，意味着代码不能继续运行。

#### Request Driven 的兜底
* 在写 Http/gPRC 服务时，通常注入的第一个 middleware 就是 recover
捕获 panic 、打印并 abort 掉请求，响应失败

* 避免创建"野生" goroutine  
不要直接使用 go 关键字，使用一个如下的方式创建 goroutine。参见demo: 
[bad](./cmd/go/bad/bad.go) 和 [better](./cmd/go/better/better.go)
```go
package sync

func Go(f func()) {
    go func() {
        defer func() {
            if err := recover(); err != nil {
                // handle the err, logging, etc.
            }
        }()
        f() 
    }()
}
```
* 使用 work pool 模式。将请求通过 channel 传递给 work pool 来处理。

> Q: 什么时候 panic  
> A: main 函数、init 函数资源初始化，如果失败无法正常服务；读配置有问题时，防御性编程

> Q: 如果应用启动时，连接不上数据库但可以连上缓存，是允许启动还是 panic?  
> A: 对读多写少场景，可以，多数读请求可能会命中缓存，服务可用，写请求响应失败就是，
> 不能无节点可用，待数据库连接恢复，写服务也将恢复。**具体还是要看场景**

> Q: 如果依赖的服务不可用，启动不启动，ready不ready？
> A: 分情况：
> * 强依赖策略，blocking 直到依赖的服务恢复。不用启用，服务不可用。
> * 弱依赖策略，nonblocking，允许启动，之后不断尝试重连。启动后，虽然服务可用，但会大量报错，
> 但有些服务已经实现服务容错降级的策略，影响会比服务不可用小些。
> * 中庸策略，blocking 10s + nonblocking。
> 先尝试等待依赖服务恢复，不行了再以 nonblocking 方式提供服务，期待之后依赖服务可能恢复

_使用多个返回值和一个简单的约定，Go 解决了让程序员知道什么时候出了问题，并为真正的异常情况保留了 panic。_

对于预期外的参数，通常返回一个 error 而不是返回 ok or not、空指针，
绝对不允许使用 panic + recover 的方式处理。

> Q: 如果 DAO 查一条记录没有找到，返回空指针还是 error  
> A: 建议是返回零值+error，不要返回空指针，绝对不能用 panic + recover，**发现开除**！！！

对于真正意外的情况，那些表示不可恢复的程序错误，例如索引越界、不可恢复的环境问题、栈溢出，
我们才使用 panic。对于其他的错误情况，我们应该是期望使用 error 来进行判定。

### Go error 机制特点总结
* 简单。
* 考虑失败，而不是成功(Plan for failure, not success)。
* 没有隐藏的控制流。
* 完全交给你来控制 error。
* Error are values。

## Error Type 错误类型

### Sentinel Error 哨兵错误

预定义的特定错误。

> 这个名字来源于计算机编程中使用一个特定值来表示不可能进行进一步处理的做法。

```
if err == ErrSomething {...} // io.EOF、syscall.ENOENT...
```

使用 sentinel 值是最不灵活的错误处理策略，因为调用方必须使用 `==` 将结果与预先声明的值进行比较。
当您想要提供更多的上下文时，这就出现了一个问题，因为返回一个不同的错误将破坏相等性检查。  
甚至是一些有意义的 `fmt.Errorf()` 携带一些上下文，也会破坏调用者的 `==` ，
调用者将被迫查看 `error.Error()` 方法的输出，以查看它是否与特定的字符串匹配。

* 不依赖检查 `error.Error()` 的输出。  
不应该依赖检测 `error.Error()` 的输出，Error 方法存在于 error 接口主要用于方便程序员使用，
但不是程序(编写测试可能会依赖这个返回)。这个输出的字符串用于记录日志、输出到 stdout 等。

* Sentinel errors 会成为你 API 公共部分。  
如果您的公共函数或方法返回一个特定值的错误，那么该值必须是公共的，当然要有文档记录，
这会增加 API 的表面积。  
如果 API 定义了一个返回特定错误的 interface，则该接口的所有实现都将被限制为仅返回该错误，
即使它们可以提供更具描述性的错误。
比如 `io.Reader`。像 `io.Copy` 这类函数需要 `reader` 的实现者比如返回 `io.EOF`
来告诉调用者没有更多数据了，但这又不是错误。

* Sentinel errors 在两个包之间创建了依赖。  
Sentinel errors 最糟糕的问题是它们在两个包之间创建了源代码依赖关系。
例如，检查错误是否等于 `io.EOF` ，您的代码必须导入 `io` 包。
这个特定的例子听起来并不那么糟糕，因为它非常常见，但是想象一下，
当项目中的许多包导出错误值时，存在耦合，项目中的其他包必须导入这些错误值
才能检查特定的错误条件(in the form of an import loop)。

结论: 尽可能避免 sentinel errors。  
建议避免在编写的代码中使用 sentinel errors。
在标准库中有一些使用它们的情况，但这不是一个您应该模仿的模式。

### Error types

Error type 是实现了 error 接口的自定义类型。
例如如下的 `MyError` 类型记录了文件和行号，以展示发生了什么。
```go
package log

type MyError struct {
    Msg string
    File string
    Line int
}
```

调用者可以使用断言将 error 转换成特定实现类型，来获取更多的上下文信息。

与错误值相比，错误类型的一大改进是它们能够包装底层错误以提供更多上下文。
一个不错的例子就是 `os.PathError` 他提供了底层执行了什么操作、那个路径出了什么问题。

调用者要使用类型断言和类型 `switch`，就要让自定义的 error 变为 public。
这种模型会导致和调用者产生强耦合，从而导致 API 变得脆弱。  
结论是尽量避免使用 error types，虽然错误类型比 sentinel errors 更好，
因为它们可以捕获关于出错的更多上下文，但是 error types 共享 error values 许多相同的问题。  
因此，建议避免使用错误类型，或者至少避免将它们作为公共 API 的一部分。

### Opaque errors 不透明错误
不透明错误处理。只有出错或没有出错。

这是最灵活的错误处理策略，因为它要求代码和调用者之间的耦合最少。  
我将这种风格称为不透明错误处理，因为虽然您知道发生了错误，但您没有能力看到错误的内部。
作为调用者，关于操作的结果，您所知道的就是它起作用了，或者没有起作用(成功还是失败)。  
这就是不透明错误处理的全部功能：只需返回错误而不假设其内容。

#### **Assert errors for behaviour, not type** 断言error实现了特定行为而不是类型
在少数情况下，这种二分错误处理方法是不够的。
例如，与进程外的世界进行交互(如网络活动)，需要调用方调查错误的性质，以确定重试该操作是否合理。
在这种情况下，我们可以断言错误实现了特定的行为，而不是断言错误是特定的类型或值。

```go
package demo

type temporary interface{
    Temporary() bool
}

// IsTemporary returns true if err is temporary
func IsTemporary(err error) bool {
    te, ok := err.(temporary)
    return ok && te.Temporary()
}
```

**只对行为感兴趣**

> 典型使用：[k8s api errors](https://github.com/kubernetes/apimachinery/blob/master/pkg/api/errors/errors.go)
> 中各种 IsXXX 方法

## Handing Error 高效处理Error的套路

### Indented flow is for errors 缩进的代码只是用于处理错误

无错误的正常流程代码，将成为一条直线，而不是缩进的代码。

### Eliminate error handling by eliminating errors 通过消减error来减少错误处理

* 如果调用返回结果与需要 return 的结果是 match 的，直接返回，
不要多写罗嗦的 if err != nil 判断代码

```go
package bad

func Authenticate(r *Request) error {
    err := authenticate(r.User) // return authenticate(r.User)
    if err != nil { // [1]
        return err  // [2]
    }               // [3]
    return nil      // [4]
}
// 1-4 行是没有意义的，直接返回 `authenticate()`方法的结果就好了
```

* 通过包装重复的错误处理过程，简化错误处理

使用`io.Reader`读取内容的行数
```go
package bad

import (
    "io"
    "bufio"
)

func CountLine(r io.Reader) (int, error) {
    var (
        br    = bufio.NewReader(r)
        lines int
        err   error
    )

    for {
        _, err = br.ReadString('\n')
        lines++
        if err != nil {
            break
        }
    }

    if err != io.EOF {
        return 0, err
    }
    return lines, nil
}
```
使用`bufio.Scanner`改进
```go
package batter

import (
    "io"
    "bufio"
)

func CountLines(r io.Reader) (int, error) {
    sc := bufio.NewScanner(r)
    lines := 0

    for sc.Scan() {
        lines++
    }

    return lines, sc.Err()
}
```

类似的，还有 `sql.Rows`。

`errWriter`，标准库中有多处应用（多出现于测试文件中），这是一种常用套路，要学会应用。
```
type errWriter struct {
    io.Writer
    err error
}

func (e *errWriter) Write(buf []byte) (int, error) {
    if e.err != nil {
        return 0, e.err
    }

    var n int
    n, e.err = e.Writer.Write(buf)
    return n, nil
}
```

> `errWriter` 的用法不是完美的，因为没有提前返回，可能有大量的数据计算被丢弃浪费了。

## Wrap errors 错误包装

如果错误没有就地处理，需要向调用者输出，最终在调用栈的根部需要处理错误，这时将错误输出，
打印出来的只有基本的错误信息，缺少错误生成时的 file:line 信息、没有调用堆栈。

为了追踪错误，有一种做法是使用 `fmt.Errorf` 以原 err 加一些描述信息生成新的 error 抛出，
但这种模式与 sentinel errors 或 type assertions 的使用不兼容：
破坏了原始错误，导致等值判定失败。

* You should only handle errors once. Handling an error means inspecting the
error value, and making a single decision.
只应处理错误一次。

如下的代码，在错误处理中，执行了两个任务：记录日志、抛出错误
```
func WriteAll(w io.Writer, buf []byte) error {
    _, err := w.Write(buf)
    if err != nil {
        log.Println("unable to write:", err) // 记录错误到日志
        return err                           // 将错误交给调用者
    }
    return nil
}
```
如果在错误上抛的过程中，调用者也会记录并返回原错误，那么会重复输出大量日志，形成噪音。

**Go 中的错误处理契约规定，在出现错误的情况下，不能对其他返回值的内容做出任何假设。**

如果就地处理错误，要处理完整。

* 错误要被日志记录。
* 应用程序处理错误，保证100%完整性。
* 之后不再报告当前错误。

### github.com/pkg/errors

`errors.Wrap(err, msg)`: withStack -> (withMessage -> err + msg) + stack 
`errors.WithMessage(err, msg)`: withMessage -> err + msg  
`errors.WithStack(err, msg)`: withStack -> err + stack  
`errors.Cause(err)`: 取出层层包装中的根因  
`%+v`： `fmt`包扩展格式，打出调用栈

正确使用姿势：
* 在你的应用代码中，使用 `errors.New` 或者  `errros.Errorf` 返回错误。
（此处的`errors`包是 `github.com/pkg/errors`包）
* 如果调用应用代码中其他的函数，通常简单的直接返回。
* 如果和其他库（第三方库、基础库kit）进行协作，
考虑使用 `errors.Wrap` 或者 `errors.Wrapf` 保存根因和堆栈信息。
同样适用于和标准库协作的时候。
* 直接返回错误，而不是每个错误产生的地方到处打日志。
* 在程序的顶部或者是工作的 goroutine 顶部(请求入口)，使用 `%+v` 把堆栈详情记录。
* 使用 `errors.Cause` 获取 root error，再进行和 sentinel error 判定。

总结：
* Packages that are reusable across many projects only return root error values.  
**选择 wrap error 是只有 applications 可以选择应用的策略。**
具有最高可重用性的包只能返回根错误值。
此机制与 Go 标准库中使用的相同(kit 库的 sql.ErrNoRows)。
* If the error is not going to be handled, wrap and return up the call stack.  
这是关于函数/方法调用返回的每个错误的基本问题。
如果函数/方法不打算处理错误，那么用足够的上下文 wrap errors 并将其返回到调用堆栈中。
例如，额外的上下文可以是使用的输入参数或失败的查询语句。
确定您记录的上下文是足够多还是太多的一个好方法是检查日志并验证它们在开发期间是否为您工作。
* Once an error is handled, it is not allowed to be passed up the call stack any longer.  
一旦确定函数/方法将处理错误，错误就不再是错误。
如果函数/方法仍然需要发出返回，则它不能返回错误值。
它应该只返回零(比如降级处理中，你返回了降级数据，然后需要 return nil)。

## Go 1.13 errors

### Unwrap、Is、As

go1.13为 `errors` 和 `fmt` 标准库包引入了新特性，以简化处理包含其他错误的错误。
其中最重要的是: 包含另一个错误的 `error` 可以实现返回底层错误的 `Unwrap` 方法。
如果 `e1.Unwrap()` 返回 `e2`，那么我们说 `e1` 包装 `e2`，您可以展开 `e1` 以获得 `e2`。

go1.13 `errors` 包包含两个用于检查错误的新函数：`Is` 和 `As`。  
`Is`: 判断根因是否是指定 sentinel error  
`As`: 尝试展开错误并断言根因为指定类型错误（通过第二个参数，指定类型错误的指针），
如果成功，`As` 会返回 `true`，同时第二个参数会被赋值为根因

在 Go 1.13中 `fmt.Errorf` 支持新的 `%w` 谓词。  
用 `%w` 包装错误可用于 `errors.Is` 以及 `errors.As`

> 因为`%w`包装的错误没有包含调用栈信息，使用的并不多，通常使用 `pkg/errors` 的 `Wrapf`更多

#### Customizing error tests with Is and As methods

标准库`errors.Is` 函数中会尝试使用`err.(interface { Is(error) bool })`断言要判断的错误`err`
是否实现了 `Is(error) bool` 方法。如果实现了，会使用`err`的`Is`方法进行判定。

通过（覆盖/）实现`Is(error) bool` 方法，可以自定义`errors.Is`函数的判断结果。
因为默认情况下，`Is` 只是比较错误值的指针值。

> 建议在方法/函数备注中说明会返回什么错误，是否被包装

> 通常不返回 sentinel 错误值，而是返回包装值，以携带上下文信息。调用者使用 `Is`进行判断处理。

#### errors & github.com/pkg/errors

`github.com/pkg/errors` 兼容 `errors`，使用相同的签名包装了`errors`中的
`Is`、`As`、`Unwrap`等函数

> 优先使用 `errors.Wrapf` 而不是 `fmt.Errorf` + `%w`，以包含堆栈信息。

## Go 2 Error Inspection
略

> [Proposal: Go 2 Error Inspection](https://go.googlesource.com/proposal/+/master/design/29934-error-values.md)
