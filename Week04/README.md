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

> kit 库：工具包/基础库/框架库

每个公司都应当为不同的微服务建立一个统一的 kit 工具包项目.  
基础库 kit 为**独立项目**，**公司级建议只有一个**（通过行政手段保证），
按照功能目录来拆分会带来不少的管理工作，因此建议合并整合。

尽量不要在 Kit 项目中引入第三方依赖。容易受到第三方变更的影响。

> https://www.ardanlabs.com/blog/2017/02/package-oriented-design.html  
> > To this end, the Kit project is not allowed to have a vendor folder.
> > If any of packages are dependent on 3rd party packages, 
> > they must always build against the latest version of those dependences.

kit 项目必须具备的特点:
* 统一
* 标准库方式布局
* 高度抽象
* 支持插件

## Service Application Project Layout

* `/api` API协议定义目录。`xxapi.proto` protobuf 文件，以及生成的 go 文件。  
  B 站通常把 API 文件直接在 proto 文件中描述。
* `/configs` 配置文件模板或者默认配置。Toml、Yaml、Ini、Properties
* `/test` 额外的外部测试应用程序和测试数据。  
  可以随时根据需求构造 `/test` 目录。对于较大的项目，有一个数据子目录是有意义的。
  例如，你可以使用 `/test/data` 或 `/test/testdata` (如果你需要忽略目录中的内容)。
  请注意，Go 还会忽略以“.”或“_”开头的目录或文件，因此在如何命名测试数据目录方面有更大的灵活性。

> **不应该包含 `/src`**。不要将项目级别 src 目录 与 Go 用于其工作空间的 src 目录。

```
├── README.md
├── api/
├── cmd/
├── configs/
├── go.mod
├── go.sum
├── internal/
└── test/
```

如果一个 project 里要放置多个微服务的 app (类似 monorepo)：
* app目录内的每个微服务按照自己的全局唯一名称（比如 “account.service.vip”）来建立目录，
  如: account/vip/*。
* 和app平级的目录pkg存放业务有关的公共库(非基础框架库)。
  如果应用不希望导出这些目录，可以放置到 myapp/internal/pkg 中。

> Service Tree ...

微服务中的 app 服务类型分为：

* interface 对外的BFF服务，接受来自用户的请求， 比如暴露了 HTTP/gRPC 接口。
* service 对内的微服务，仅接受来自内部其他服务或 者网关的请求，比如暴露了gRPC 接口只对内服务。
* admin 区别于service，更多是面向运营测的服务， 通常数据权限更高，隔离带来更好的代码级别安全。
* job 流式任务处理的服务，上游一般依赖message broker。
* task 定时任务，类似cronjob，部署到task托管平 台中。

> cmd 目录中代码负责启动、关闭、配置初始化等

```
├── cmd/
│   ├── myapp1-admin/
│   ├── myapp1-interface/
│   ├── myapp1-job/
│   ├── myapp1-service/
│   └── myapp1-task/
```

> 依赖倒置。IoC/DI。

### Layout v1
```
├── xxxservice/
│   ├── api/ # <- 存放 API 定义（protobuf）及对应生成的 stub 代码、swagger.json
│   ├── cmd/ # <- 存放服务 bin 代码
│   ├── configs/ # <- 存放服务所需的配置文件
│   ├── internal/ # <- 避免有同业务下有人跨目录引用内部的 model、dao 等内部 struct 。
│   │   ├── model/ # <- 存放 Model 对象
│   │   ├── dao/ # <- 数据读写层，数据库和缓存全部在这层统一处理，包括 cache miss 处理。
│   │   ├── service/ # <- 组合各种数据访问来构建业务逻辑。
│   │   ├── server/ # <- 放置 HTTP/gRPC 的路由代码，以及 DTO 转换的代码。
```

DTO，Data Transfer Object: 
数据传输对象，泛指用于展示层/ API 层与服务层（业务逻辑层）之间的数据传输对象。
从概念上讲，包含了 VO（View Object） 视图对象。

直接使用 Model 对象 / Entity 对象，用于数据传输/展示有以下问题：
* Model 对应的是存储层，与存储一一映射。直接用于传输，会过度暴露字段，需要专门处理
* 展示形式可能不匹配，或存在兼容性问题，也需要专门处理
* 以上处理逻辑的代码位置分层定位职责不清，可能导致代码管理混乱

server 层依赖proto定义的服务作为入参，提供快捷的启动服务全局方法。这一层的工作可以被 kit 库功能取代。

在 api 层，protobuf 文件生成了 stub 代码 interface，在 service 层中提供了实现。

DO, Domain Object: 领域对象。
v1 版中没有引入 DO 对象，或者说使用了贫血模型，缺乏 DTO -> DO 的对象转换。

### Layout v2
```
├── CHANGELOG
├── OWNERS
├── README
├── api/
├── cmd/
│   ├── myapp1-admin/
│   ├── myapp1-interface/
│   ├── myapp1-job/
│   ├── myapp1-service/
│   └── myapp1-task/
├── configs/
├── go.mod
└── internal/ # <- 避免有同业务下有人跨目录引用了内部的 biz、 data、service 等内部 struct
    ├── biz/ # <- 业务逻辑的组装层，类似DDD的domain层。repo 接口在这里定义，使用依赖倒置的原则。
    ├── data/ # <- 业务数据访问，包含cache、db等封装，实现了biz的repo 接口。
    ├── pkg/
    └── service/
```

data 层：可能会把 data 与 dao 混淆在一起，data 偏重业务的含义，
它所要做的是将领域对象重新拿出来，去掉了 DDD 的 infra层

service 层，实现了 api 层定义的 stub 接口。
类似DDD的application层，处理 DTO 到 biz 领域实体的转换(DTO -> DO)，
同时协同各类 biz 交互， 但是不应处理复杂逻辑。

PO，Persistent Object：持久化对象，
它跟持久层（通常是关系型数据库）的数据结构形成一一对应的映射关系。
如果持久层是关系型数据库，那么数据表中的每个字段（或若干个）就对应PO的一个（或若干个）属性。

> https://github.com/facebook/ent

## Lifecycle

依赖注入：1、方便测试；2、单次初始化和复用

所有 HTTP/gRPC 依赖的前置资源初始化，包括 data、biz、service，之后再启动监听服务。
> https://github.com/go-kratos/kratos/blob/v2/transport/http/service.go

使用 https://github.com/google/wire ，来管理所有资源的依赖注入。
手撸资源的初始化和关闭是非常繁琐，容易出错的。
使用依赖注入的思路 DI，结合 google wire，静态的 go generate 生成静态的代码，
可以在很方便诊断和查看，不是在运行时利用 reflection 实现。

# API设计
## gRPC VS HTTP RESTFul

* gRPC 基于 IDL，文档、API定义、代码都是一致的，而 HTTP RESTFul 文档与接口常常脱节。
* 可以生成各客户端调用 Stub 代码，而 HTTP RESTFul 客户端代码通常需要开发人员手工实现。
* gRPC 定义了调用使用的 message ，实际上等同于给定了 DTO，
  促进（强迫）服务端实现进行 DTO <-> DO 间的转换。
* gRPC 可以方便的实现元数据交换，如认证或跟踪等元数据。HTTP 通常需要将这些数据放置到请求 Header 中。
* gRPC 使用标准化状态码。

内网间的RPC调用推荐使用 gRPC

## API Project

Q: API 定义 proto 如何共享使用？

一种做法是API定义方/提供方在 `/api` 目录中生成 Client Stub 代码，将代码仓库访问权限授与API使用方，
API使用方引用 Client Stub 代码。
这样做项目权限管理比较麻烦，太过宽松，Git 无法进行细粒度权限限制，可能过度暴露API提供方的内部代码。

另一种做法是使用一个统一的 API 仓库，统一检索和规范 API。
将所有对内对外的项目的 API `/api` 中 protobuf 文件整合到一个统一的项目中。
> https://github.com/googleapis/googleapis  
> https://github.com/envoyproxy/data-plane-api  
> https://github.com/istio/api

* 规范化检查，API Lint
* 方便跨部门协作
* 基于git，版本管理
* API Design review，基于 commit diff

为了控制对 API 文件的读写操作，需要权限管理，使用目录 OWNERS 文件：
关闭主 API 仓的写权限，使用 Merge Request + Approve 的方式进行管理，
其中可以使用自动化工具进行检查是否 Merge Request 发起人、API 目录是否匹配进行自动拒绝越权操作。

API protobuf 仓还可以有不同编程语言的子仓，
通过 Hook，自动推送生成各语言的 Stub 代码到对应语言的代码仓库中

### API Project Layout

项目中定义 proto，以 `api` 为包名根目录：
```
├── prject-demo/
│   ├── api/ # <- 服务 API 定义
│   │   ├── path/            # <- ↓
│   │   │   ├── of/          # <- 服务 API 定义路径
│   │   │   │   ├── service/ # <- ↑
│   │   │   │   │   ├── v1/ # <- API 定义大版本
│   │   │   │   │   │   ├── demo.proto # <- API 定义文件
```

在统一仓库中管理 proto, 以仓库为包名根目录：
```
├── api/ # <- 服务API定义
│   ├── path/                     # <- ↓
│   │   └── of/                   #
│   │       └── service1/         #    与各项目中 /api 目录中内容路径对应
│   │           └── v1/           #
│   │               ├── api.proto # <- ↑
│   │               └── OWNERS
│   └── path/
│       └── of/
│           └── service2/
│               ├── v1/
│               │   ├── api.proto
│               │   └── OWNERS
│               └── v2/
│                   ├── api.proto
│                   └── OWNERS
├── annotations/ # <- 注解定义 options
├── metadata/ # <- 定义对外服务的统一元数据
│   ├── locale/
│   ├── network/
│   ├── device/
│   └── ... 
├── rpc/ # <- 定义统一状态码
│   └── status.proto
├── third_party/ # <- 第三方引用
```

## API Compatibility 兼容性
向后兼容（非破坏性）的修改：
* 给API服务定义添加 API 接口。从协议的角度看，这始终是安全的。
* 给请求消息添加字段。只要客户端在新版和旧版中对该字段的处理不保持一致，添加请求字段就是兼容的。  
  客户端不应在处理新字段时忽略对旧字段的处理。
* 给响应消息添加字段。
  在不改变其他响应字段的行为的前提下，非资源（例如，ListBooksResponse）
  的响应消息可以扩展而不必破坏客户端的兼容性。
  即使会引入冗余，先前在响应中填充的任何字段应继续使用相同的语义填充。

向后不兼容（破坏性）的修改：
* 删除或重命名：服务、字段、方法或枚举值。  
  从根本上说，如果客户端代码可以引用某些东西，那么删除或重命名它都是不兼容的变化，
  这时必须修改 major 版本号。
* 修改字段的类型  
  即使新类型是传输格式兼容的，这也可能会导致客户端库生成的代码发生变化，因此必须增加 major 版本号。
  对于编译型静态语言来说，会容易引入编译错误。
* 修改现有请求的可见行为  
  客户端通常依赖于 API 行为和语义，即使这样的行为没有被明确支持或记录。
  因此，在大多数情况下，修改 API 数据的行为或语义将被消费者视为是破坏性的。
  如果行为没有加密隐藏，您应该假设用户已经发现它，并将依赖于它。
* 给（会导致更新的）资源消息添加读取/写入字段


# 配置管理

# 包管理

# 测试
> 单元测试是系统演进中基层稳定可靠的必要保证。