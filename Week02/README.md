学习笔记
---
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

* 使用合适的工具

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

### 

## Go 1.13 errors 1.13 Wrapper

## Go 2 Error Inspection
