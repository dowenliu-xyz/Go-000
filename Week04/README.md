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

## API Naming Conventions
包名应为应用的标识（APP_ID），用于生成 gRPC 请求路径，或者 proto 之前进行引用 Message。  
proto 文件中声明的包名称应该与产品和服务名称保持一致。  
带有版本的 API 的软件包名称必须以此版本结尾

如 `/my/package/v1` 为 API 目录，proto 文件中 package 应为：
```
package my.package.v1;
```
对应的 gRPC RequestURL： `/my.pckage.v1.{service}/{method}`

命名规范示例

| API 名称 | 示例 |
| --- | --- |
| 产品名称 | Google Calendar API |
| 服务名称 | calendar.googleapis.com |
| 软件包名称 | google.calendar.v3 |
| 接口名称 | google.calendar.v3.CalendarService |
| 来源目录 | /google/calendar/v3 |
| API 名称 | calendar |

> 建议为每个 gRPC 服务方法定义输入输出消息（不要使用 google.protobuf.Empty ，无法扩展），
> 为未来可能的字段扩展留有空间。

## API Primitive Fields 基础类型字段
gRPC 默认使用 Protobuf v3 格式，消息结构默认全部都是 optional 字段。  
如果基础类型字段没有被赋值，默认会赋值为基础类型字段的默认值，0或者""。

> Protobuf v3 中，建议使用：
> https://github.com/protocolbuffers/protobuf/blob/master/src/google/protobuf/wrappers.proto
> Warpper 类型的字段，即包装一个 message，使用时变为指针。

## API Errors

### ❌ 全局错误码
全局错误码，是松散、易被破坏契约的。

在每个服务传播错误的时候，做一次翻译，这样保证每个服务 + 错误枚举，应该是唯一的，
而且在 proto 定义中是可以写出来文档的。
> https://github.com/googleapis/googleapis/blob/master/google/rpc/status.proto  
> https://github.com/googleapis/googleapis/blob/master/google/rpc/error_details.proto#L112

### 使用一小组标准错误配合大量资源
例如，服务器没有定义不同类型的“找不到”错误，
而是使用一个标准 google.rpc.Code.NOT_FOUND 错误代码并告诉客户端找不到哪个特定资源。
状态空间变小降低了文档的复杂性，在客户端库中提供了更好的惯用映射，并降低了客户端的逻辑复杂性，
同时不限制是否包含可操作信息。
> https://github.com/googleapis/googleapis/blob/master/google/rpc/error_details.proto

设计 API 错误码时，将业务状态码与 HTTP 状态码进行映射：
* 方便运维工具监控接口状态，及时报警。
* 有利用错误码收敛。

gRPC 错误码会映射到 HTTP 状态码。

| HTTP | RPC | 错误消息示例 |
| :--- | :--- | :--- |
| 400 | INVALID_ARGUMENT | 请求字段 x.y.z 是 xxx ，预期为 [yyy,zzz] 内的一个。 |
| 400 | FAILED_PRECONDITION |  资源 xxx 是非空目录，因此无法删除。 |
| 400 | OUT_OF_RANGE | 客户端指定了无效范围。 |
| 401 | UNAUTHENTICATED | 由于 OAuth 令牌丢失、无效或过期，请求未通过身份验证。 |
| 403 | PERMISSION_DENIED | 客户端权限不足。可能的原因包括 OAuth 令牌的覆盖范围不正确、客户端没有权限或者尚未为客户端项目启用 API。 |
| 404 | NOT_FOUND | 找不到指定的资源，或者请求由于未公开的原因（例如白名单）而被拒绝。 |
| 409 | ABORTED | 并发冲突，例如读取/修改/写入冲突。 |
| 409 | ALREADY_EXISTS | 客户端尝试创建的资源已存在。 |
| 429 | RESOURCE_EXHAUSTED | 资源配额不足或达到速率限制。如需了解详情，客户端应该查找 google.rpc.QuotaFailure 错误详细信息。 |
| 499 | CANCELLED | 请求被客户端取消。 |
| 500 | DATA_LOSS | 出现不可恢复的数据丢失或数据损坏。客户端应该向用户报告错误。 |
| 500 | UNKNOWN | 出现未知的服务器错误。通常是服务器错误。 |
| 500 | INTERNAL | 出现内部服务器错误。通常是服务器错误。 |
| 501 | NOT_IMPLEMENTED | API 方法未通过服务器实现。 |
| 503 | UNAVAILABLE | 服务不可用。通常是服务器已关闭。 |
| 504 | DEADLINE_EXCEEDED | 超出请求时限。仅当调用者设置的时限比方法的默认时限短（即请求的时限不足以让服务器处理请求）并且请求未在时限范围内完成时，才会发生这种情况。 |
> https://cloud.google.com/apis/design/errors#handling_errors

### 错误传播
> https://cloud.google.com/apis/design/errors#error_propagation

如果您的 API 服务依赖于其他服务，则不应盲目地将这些服务的错误传播到客户端。
在翻译错误时，建议执行以下操作：
* 隐藏实现详细信息和机密信息
* 调整负责该错误的一方。例如，从另一个服务接收到 INVALID_ARGUMENT 错误的服务器应该将
  INTERNAL 传播给它自己的调用者。
> 吞掉外部依赖返回的错误，响应自己的错误，使用`errors.Wrapf`。  
> 如底层的 `sql.ErrNoRows` 而响应给客户端 NotFound:
> ```
> func GetUsers() ([]User, error) {
>     sql := ...
>     rows, err := db.Query(sql)
>     if err != nil {
>         ...
>     }
>     ...
>     err := rows.Err()
>     if err != nil {
>         if err == sql.ErrNoRows {
>             return nil, errors.Wrapf(code.ErrNotFound,
>                 "query %q failed(%v)", sql err)
>         }
>         ...
>     }
>     ...
> }
> ```

> Kratos v2 错误处理  
> https://github.com/go-kratos/kratos/blob/v2/errors/errors.go  
> https://github.com/go-kratos/kratos/blob/v2/errors/codes.go  
> 

gRPC 将错误放到元数据中传递给客户端，这样：
* 不会污染请求/响应消息体
* 错误只需要在调用返回中判断一次，不必再次从消息体中取出判断业务错误

> service error (Server side) -> gRPC error -> service error (Client side)

## API Design

FieldMask 部分更新方案。 `google.protobuf.FieldMask`
> https://github.com/protocolbuffers/protobuf/blob/master/src/google/protobuf/field_mask.proto

客户端可以指定需要更新的字段信息：
```
paths: "author"
paths: "submessage.submessage.field"
```
空 FieldMask 默认应用到所有字段

⭐️ 极力推荐：谷歌API设计指南：https://cloud.google.com/apis/design

# 配置管理
大体分为四类：
* 环境（变量）配置
  * Region 区域。 如华北、华南。 大区
  * Zone 可用区。如杭州01、青岛01。
  * Cluster 集群
  * Environment 环境 如PRD、UAT、FAT、test、dev
  * Color 染色信息
  * Discovery 服务发现用的 IP 、端口号等信息。
  * AppID 应用id，
  * Host 主机名
  
  之类的环境信息，通过在线运行时平台打入到容器或者物理机，供 kit 库读取使用。
* 静态配置
  资源需要初始化的配置信息，比如 http/gRPC server、redis、mysql 等。
  这类资源在线变更配置的风险非常大，不鼓励 on-the-fly 变更，很可能会导致业务出现不可预期的事故。
  **变更静态配置和发布 binaray app 没有区别，应该走一次迭代发布的流程。**
* 动态配置
  应用程序可能需要一些在线的开关，来控制业务的一些简单策略，动态变更业务流，会频繁的调整和使用。
  这类在线开头配置建议使用基础类型（int, bool等）配置，
  同时可以考虑结合类似 https://pkg.go.dev/expvar 来结合使用。
* 全局配置
  通常，我们依赖的各类组件、中间件都有大量的默认配置或者指定配置，
  在各个项目里大量拷贝复制，容易出现意外，
  所以我们使用全局配置模板来定制化常用的组件，然后再特化的应用里进行局部替换。

## Functional options
### 原由
假设以下场景：提供一个函数，连接到Redis，并返回连接+error。  
可以轻易的给出以下代码：
```
// DialTimeout acts like Dial for establishing the
// connection to the server, writing a command and reading a reply.
func Dial(network, address string) (Conn, error)
```
但使用方很快提出了新的需求：我要自定义超时时间！”，“我要设定 Database！”，
“我要控制连接池的策略！”，“我要安全使用 Redis，让我填一下 Password！”，
“可以提供一下慢查询请求记录，并且可以设置 slowlog 时间？”

于是作为API提供者，你需要增加一系列函数以提供新特性：
```
// DialTimeout acts like Dial for establishing the
// connection to the server, writing a command and reading a reply.
func Dial(network, address string) (Conn, error)

// DialTimeout acts like Dial but takes timeouts for establishing the
// connection to the server, writing a command and reading a reply.
func DialTimeout(network, address string, connectTimeout, readTimeout, writeTimeout time.Duration) (Conn, error)

// DialDatabase acts like Dial but takes database for establishing the
// connection to the server, writing a command and reading a reply.
func DialDatabase(network, address string, database int) (Conn, error)

// DialPool
func DialPool...
```

`net/http` 中 `Server` 结构体提供了另一种思路：  
在组装配置阶段，向`Server` 结构体中填充需要配置的字段，
不需要配置的字段不填充，让 `Server` 对象使用默认值填充或执行默认行为。这样做：  
好处：
* 配置字段定义在结构体中，可以为字段编写全面（且复杂）的说明文档。

不好的地方：
* 字段的默认值，代表的含义或导致的行为必须有文档中说明。
* 字段是否可选还是必需，只能在文档中指出，无法在编译阶段进行检查。

参考`Server`的定义与使用方式，我们很直接的想到可以将之前场景中的各种需求封装归集到 `Config` 结构体中：
```
// Config redis settings.
type Config struct {
  *pool.Config
  Addr string
  Auth string
  DialTimeout time.Duration
  ReadTimeout time.Duration
  WriteTimeout time.Duration
}
```

Config 对象如何使用呢？
```
// NewConn new a redis conn. [1]
func NewConn(c Config) (cn Conn, err error)

// NewConn new a redis conn. [2]
func NewConn(c *Config) (cn Conn, err error)

// NewConn new a redis conn. [3]
func NewConn(c ...*Config) (cn Conn, err error)
```
* [1]的方式可以保证`NewConn(c)`之后，修改 `c` 不会对 `cn` 有影响，
  因为 `Config` 参数 `c` 在传入后被复制，`NewConn` 使用的是参数 `c` 的复本。  
  但[1]的方式无法传入 `nil`，无法在 `NewConn` 函数内提供配置默认值。
* [2]的方式在`NewConn(c)` 之后，修改 `c` 的字段值，对 `cn` 的影响是未知的，Undefined，
  调用者必须按照君子约定，不再修改 `c`。
  好处就是，如果 `NewConn` 传入 `nil`，`NewConn` 可以应用一个默认配置值
* [1]、[2] 的方式都有一个共同的问题，无法在 `Config` 中区分可选必选字段，并为可选字段提供默认值。
  [3] 的方式可以帮助应用默认配置：传入两个 `Config`，第一个提供完整的默认配置，第二个提供覆盖配置，
  在 `NewConn` 内部进行配置合并。
  但因为方法参数签名使用了`...` 无法限制调用者会传入多少个 `Config` 对象。

> 尽量不要向公开函数传 nil 参数
> 
> “I believe that we, as Go programmers,
> should work hard to ensure that nil is never a parameter
> that needs to be passed to any public function.”
> – Dave Cheney
### Functional options
> [Self-referential functions and the design of options](https://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html)
> by Rob Pike
> 
> [Functional options for friendly APIs](https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis)
> by Dave Cheney

通过参数指定必需参数和可选配置。
```
// DialOption specifies an option for dialing a Redis server.
type DialOption struct {
  f func(*dialOptions)
}

// Dial connects to the Redis server at the given network and
// address using the specified options.
func Dial(network, address string, options ...DialOption) (Conn, error) {
  do := dialOptions{ // 配置选项结构体对象
    dial: net.Dial,
  }
  for _, option := range options {
    option.f(&do)
  } // ...
}
```
通过方法签名，强制保证了必需参数，同时，使用不定长的 `options` 参数，指定对可选参数的设置行为。
> https://github.com/go-kratos/kratos/blob/master/pkg/cache/redis/conn.go#L98

前面的代码中 `DialOption` 还可以直接简化为 `type DialOption func(*dialOptions)`
```
// DialOption specifies an option for dialing a Redis server.
type DialOption func(*dialOptions)


// Dial connects to the Redis server at the given network and
// address using the specified options.
func Dial(network, address string, options ...DialOption) (Conn, error) {
  do := dialOptions{
    dial: net.Dial,
  }
  for _, option := range options {
    option(&do)
  }
  // ...
}
```
高级玩法：清理/还原模式，应用配置并使用，用后还原配置：
```
type option func(f *Foo) option // 注意这里返回了 option

// Verbosity sets Foo's verbosity level to v.
func Verbosity(v int) option {
  return func(f *Foo) option {
    prev := f.verbosity
    f.verbosity = v
    return Verbosity(prev)
  }
}
func DoSomethingVerbosely(foo *Foo, verbosity int) {
  // Could combine the next two lines,
  // with some loss of readability.
  prev := foo.Option(pkg.Verbosity(verbosity)) // 应用配置
  defer foo.Option(prev) // defer 还原配置
  // ... do some stuff with foo under high verbosity.
}
```
DialOption + dialOptions 仍然有一些缺点：
dialOptions 是库内置且隐藏结构体，配置项只能由库定义无法外部扩展；
DialOption 方法也只能由库自己提供，外部无法编写。

gRPC `CallOption` 的做法更进一步，它将选项的定义也交由外部扩展：
```
type GreeterClient interface {
SayHello(ctx context.Context, in *HelloRequest, opts ...grpc.CallOption) (*HelloReply, error)
}

type CallOption interface {
before(*callInfo) error
after(*callInfo)
}
// EmptyCallOption does not alter the Call configuration.
type EmptyCallOption struct{} // 这个类型实现了 CallOption 接口

// TimeoutCallOption timeout option.
type TimeoutCallOption struct {
grpc.EmptyCallOption // 内嵌 EmptyCallOption 自动实现 CallOption 接口
Timeout time.Duration
}
```

### Hybrid APIs 混合 API
使用 `DialOption` + `dialOptions` 这样的 Functional Options ，不方便映射加载 JSON/YAML 配置，
所以不得不需要保留 `func NewConn(c *Config) (Conn, error)`这样的配置函数。

其实这种结果是因为没有将加载配置文件和配置选项这两件事解耦。

### Configuration & APIs

“For example, both your infrastructure and interface might use plain JSON.
**However, avoid tight coupling between the data format you use as the
interface and the data format you use internally.**
For example, you may use a data structure internally that contains the data
structure consumed from configuration.
The internal data structure might also contain completely
implementation-specific data that never needs to be surfaced outside of the system.”  
  -- the-site-reliability-workbook 2

> the-site-reliability-workbook 2


> 编辑配置用的工具最好支持：
> * 语义验证
> * 语法高亮
> * Lint
> * 格式化

* 仅保留 options API
* config file 和 options struct 解耦
  ```
  package redis
  
  // Option configures how we set up the connection.
  type Option interface {
    apply(*options)
  }
  ```
  ```
  // Options apply config to options.
  func (c *Config) Options() []redis.Options {
    return []redis.Options{
      redis.DialDatabase(c.Database),
      redis.DialPassword(c.Password),
      redis.DialReadTimeout(c.ReadTimeout),
    }
  }
  
  func main() {
    // instead use load yaml file.
    c := &Config{
      Network: "tcp",
      Addr: "127.0.0.1:3389",
      Database: 1,
      Password: "Hello",
      ReadTimeout: 1 * time.Second,
    }
    r, _ := redis.Dial(c.Network, c.Addr, c.Options()...)
  }
  ```
* 使用 Protobuf 定义配置文件、 YAML 存储配置文件
  使用 protobuf 定义配置文件:
  * 可以为配置字段定义加注解，加验证规则
  * 多语言之间配置保持一致
  ```
  syntax = "proto3";
  
  import "google/protobuf/duration.proto";
  
  package config.redis.v1;
  
  // redis config.
  message redis {
    string network = 1;
    string address = 2;
    int32 database = 3;
    string password = 4;
    google.protobuf.Duration read_timeout = 5;
  }
  ```
  由 protobuf 文件生成 Config 结构，应用 YAML 数据恢复配置
  ```
  func ApplyYAML(s *redis.Config, yml string) error {
    js, err := yaml.YAMLToJSON([]byte(yml))
    if err != nil {
      return err
    }
    return ApplyJSON(s, string(js))
  }
  // Options apply config to options.
  func Options(c *redis.Config) []redis.Options {
    return []redis.Options{
      redis.DialDatabase(c.Database),
      redis.DialPassword(c.Password),
      redis.DialReadTimeout(c.ReadTimeout),
    }
  }
  func main() {
    // load config file from yaml.
    c := new(redis.Config)
    _ = ApplyYAML(c, loadConfig())
    r, _ := redis.Dial(c.Network, c.Address, Options(c)...)
  }
  ```

### Configuration Best Practice
代码更改系统功能是一个冗长且复杂的过程，往往还涉及Review、测试等流程，
但更改单个配置选项可能会对功能产生重大影响，通常配置还未经测试。配置的目标：
* 避免复杂
* 多样的配置
* 简单化努力
* 以基础设施 -> 面向用户进行转变
* 配置的必选项和可选项
* 配置的防御编程
* 权限和变更跟踪
* 配置的版本和应用对齐
* 安全的配置变更：逐步部署、回滚更改、自动回滚

# 包管理
> https://github.com/gomods/athens  
> https://goproxy.cn

> https://blog.golang.org/modules2019  
> https://blog.golang.org/using-go-modules  
> https://blog.golang.org/migrating-to-go-modules  
> https://blog.golang.org/module-mirror-launch  
> https://blog.golang.org/publishing-go-modules  
> https://blog.golang.org/v2-go-modules  
> https://blog.golang.org/module-compatibility

Go 项目依赖都是源码依赖。

Go Mod 依赖版本冲突的话比较难以解决。

# 测试
> 单元测试是系统演进中基层稳定可靠的必要保证。

* 小型测试带来优秀的代码质量、良好的异常处理、优雅的错误报告；大中型测试会带来整体产品质量和数据验证。
* 不同类型的项目，对测试的需求不同，总体上有一个经验法则，即70/20/10原则：70%是小型测试，20%是中型测试，10%是大型测试。
* 如果一个项目是面向用户的，拥有较高的集成度，或者用户接口比较复杂，他们就应该有更多的中型和大型测试；如果是基础平台或者面向数据的项目，例如索引或网络爬虫，则最好有大量的小型测试，中型测试和大型测试的数量要求会少很多。

> Kit 库项目一定要写大量的单元测试。
> 
> 中间件项目需要大量的单元测试和混沌测试。
> 
> 微服务 API，直接做接口测试就好了。

## Unit Test
“自动化实现的，用于验证一个单独函数或独立功能模块的代码是否按照预期工作，
着重于典型功能性问题、数据损坏、错误条件和大小差一错误
（译注：大小差一(off-by-one)错误是一类常见的程序设计错误）等方面的验证”  
-- 《Google软件测试之道》


单元测试的基本要求：
* 快速
* 环境一致
* 任意顺序  
  sync.Once
* 并行

> https://pkg.go.dev/testing

利用 go 官方提供的 subtests + Gomock 完成整个单元测试。
* /api  
  比较适合进行集成测试，直接测试 API，使用 API 测试框架(例如: yapi)，维护大量业务测试 case。
* /data  
  docker compose 把底层基础设施真实模拟，因此可以去掉 infra 的抽象层。
* /biz  
  依赖  repo、rpc client，利用 gomock 模拟 interface 的实现，来进行业务单元测试。
* /service
  依赖 biz 的实现，构建 biz 的实现类传入，进行单元测试。

基于 git branch 进行 feature 开发，本地进行 unittest，
之后提交 gitlab merge request 进行 CI 的单元测试，
基于 feature branch 进行构建，完成功能测试，
之后合并 master，进行集成测试，上线后进行回归测试。

> Without integration tests, it's difficult to trust the end-to-end operation of a web service.
> 
> 对于微服务应用，不要做 /data 、/biz 、/service的单元测试，直接使用 API 测试框架（如 YAPI）测试 API 接口即可。

基于 docker-compose 实现跨平台跨语言环境的容器依赖管理方案，
以解决运行 unittest 场景下的(mysql, redis, mc)容器依赖问题:
* 本地安装 Docker。
* 无侵入式的环境初始化。
* 快速重置环境。
* 随时随地运行(不依赖外部服务)。
* 语义式 API 声明资源。
* 真实外部依赖，而非 in-process 模拟。

使用容器进行单元测试需要注意：
* 正确的对容器内服务进行健康检测，避免unittest 启动时候资源还未 ready。
* 应该交由 app 自己来初始化数据，比如 db 的scheme，初始的 sql 数据等，
  为了满足测试的一致性，在每次结束后，都会销毁容器。


* 在单元测试开始前，导入封装好的 testing 库，方便启动和销毁容器。
* 对于 service 的单元测试，使用 gomock 等库把 dao mock 掉，所以在设计包的时候，应该面向抽象编程。
* 在本地执行依赖 Docker，在 CI 环境里执行Unittest，需要考虑在物理机里的 Docker 网络，
  或者在 Docker 里再次启动一个 Docker。
