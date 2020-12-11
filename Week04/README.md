学习笔记
---

# 工程项目结构 Layout

## Standard Go Project Layout 标准Go项目目录布局

> https://github.com/golang-standards/project-layout/blob/master/README_zh.md

非常简单的项目通常只需一个 `main.go` 就可以，不需要项目布局。  
当有更多的人参与这个项目时，你将需要更多的结构，
包括需要一个 toolkit （脚手架） 来方便生成项目的模板， 尽可能大家统一的工程目录布局。

### `/cmd`
项目的主干。

每个应用程序的目录名应该与你想要的可执行文件的名称相匹配(例如，`/cmd/myapp`)。
> `go build` 默认会将 bin 项目编译成 `main.main()` 函数所在文件所在的文件夹名。
> 如 `/cmd/myapp/main.go` 会编译成名为 `myapp` 的可执行文件。  
> 如果没有 `myapp` 这样一层文件夹，直接编译 `/cmd/main.go` 会得到名为 `cmd` 的可执行文件，
> 需要手工重命名为类似 `myapp` 的名称。  
> 如果没有文件夹间的隔离，不同的包含 `main.main()` 函数的go文件处于同一文件夹下，
> 会导致包无法编译错误，只能编译单文件，无法使用项目内包依赖功能。

```
├── cmd/
│   ├── demo/
│   │   ├── demo    # <- go build 输出
│   │   └── main.go
│   └── demo1/
│       ├── demo1   # <- go build 输出
│       └── main.go
```

不要在这个目录中放置太多代码，通常这个目录不会被其他项目导入。
> 除了 Plugin 项目，其他项目导入另一个项目的 `main` 包是没有意义的。
>
> 如果你认为代码可以导入并在其他项目中使用，那么它应该位于 `/pkg` 目录中。
> 如果代码不是可重用的，或者你不希望其他人重用它，请将该代码放到 `/internal` 目录中。

### `/internal`
私有应用程序和库代码。这是你不希望其他人在其应用程序或库中导入代码。

> Go 1.4 之后强制保证。引用其他包的 `internal` 子包无法通过编译。

> 注意，你并不局限于顶级 `internal` 目录。在项目树的任何级别上都可以有多个内部目录。

> 一个大业务的不同子模块通常共用一个项目。
> 项目可以独立一个代码仓库也可与其他业务项目共用代码仓库（独立较多，像Google那样的 Mono 仓比较少）。
> 一个大业务的子模块间可能有共通的逻辑代码，统一在一个项目中可以在项目内进行代码重用。

你可以选择向 `internal` 包中添加一些额外的结构，以分隔共享和非共享的内部代码。
这不是必需的(特别是对于较小的项目)，但是最好有有可视化的线索来显示预期的包的用途。
你的实际应用程序代码可以放在 `/internal/app` 目录下(例如 `/internal/app/myapp`)，
这些应用程序共享的代码可以放在 `/internal/pkg` 目录下(例如 `/internal/pkg/myprivlib`)。
```
├── internal/
│   ├── app/           # <- 存放各 bin 应用专用的程序代码
│   │   └── myapp/     # <- 存放 myapp 专用的程序代码
│   ├── demo/          # <- 也可忽略 app 层。存放 demo 的专用程序代码。如果只有一个 bin 应用，这个层也可以去除。
│   │   ├── biz/
│   │   ├── data/
│   │   └── service/
│   └── pkg/           # <- 存放各 bin 共享程序代码，但因为有 internal 下，其他项目无法引用。
│       └── myprivlib/ # <- 按功能分 lib 包
```

### `/pkg`
外部应用程序可以使用的库代码(例如 `/pkg/mypubliclib`)。  
其他项目可以导入这些库，所以**在这里放东西之前要三思**

要显示地表示目录中的代码对于其他人来说可安全使用的，使用 `/pkg` 目录是一种很好的方式。

`/pkg` 目录内，可以参考 go 标准库的组织方式，按照功能分类。
> `/internal/pkg` 一般用于项目内的跨多应用的公共共享代码，但其作用域仅在单个项目内。

```
├── pkg/
│   ├── cache/
│   │   ├── memcache/
│   │   └── redis/
│   └── conf/
│       ├── dsn/
│       ├── env/
│       ├── flagvar/
│       └── paladin/
```

当项目根目录包含大量非 Go 组件和目录时，
使用 `pkg` 目录也是一种将 Go 代码分组到一个位置的好方法，
这使得运行各种 Go 工具变得更加容易组织。
```
.
├── README.md
├── docs/
├── example/
├── go.mod
├── go.sum
├── misc/
├── pkg/
├── third_party/
└── tool/
```
> https://travisjeffery.com/b/2019/11/i-ll-take-pkg-over-internal/

## Kit Project Layout

## 

# TBD Continue at 00:28:23

# API设计

# 配置管理

# 包管理

# 测试
> 单元测试是系统演进中基层稳定可靠的必要保证。