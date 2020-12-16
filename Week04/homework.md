课后作业
---
#题目
按照自己的构想，写一个项目满足基本的目录结构和工程，
代码需要包含对数据层、业务层、API 注册，
以及 main 函数对于服务的注册和启动，信号处理，
使用 Wire 构建依赖。可以使用自己熟悉的框架。

## 要点分解：
* 满足基本的目录结构。要包含数据层、业务层、API层
* 在 main 函数中组装、启动服务
* 要有信号处理
* 要使用 Wire

## 场景设定：CRM客户跟进日志服务。
跟进日志：销售在与客户跟进互动中，对客户状态、对客户执行的沟通行为等手工填写的操作记录。  
日志服务主要提供跟进日志的记录（增）和查询功能，一般不提供修改功能，但提供删除功能。  
实现案例假设该服务在内网微服务场景中，对调用方传递的参数 Full Trust，
在实际场景中通常会检查请求元数据中的 JWT。

## 项目实现

整体项目布局如下：
```
├── api/
│   └── crm/
│       └── customer_follow/
│           └── v1/ # api定义和生成代码。仅 http
│               ├── api.http.go # 使用了 https://github.com/nametake/protoc-gen-gohttp 扩展生成
│               ├── api.pb.go
│               └── api.proto
├── cmd/
│   └── customer_follow/
│       └── main.go # 加载配置、组装、启用服务
├── configs/
│   └── application.yaml
├── go.mod
├── go.sum
├── internal/
│   ├── customer_follow/
│   │   ├── biz/
│   │   │   └── usecase.go # 业务用例。直接使用了 ent 生成的结构体，没有单独定义 Domain Object
│   │   ├── di/ # Wire DI 和 配置 struct
│   │   │   ├── config.go # 配置 struct
│   │   │   ├── ent.go # 初始化 ent
│   │   │   ├── wire.go # service、biz、ent 初始化 wire.Build
│   │   │   └── wire_gen.go # wire gen 生成代码
│   │   ├── ent/ # ent 生成目录
│   │   │   ├── ... # 略过生成代码
│   │   │   ├── generate.go
│   │   │   ├── schema/
│   │   │   │   └── customerfollow.go
│   │   │   └── ...# 略过生成代码
│   │   └── service/ # 实现 api 中接口，进行 proto message 与 ent 生成的结构体（应该是 PO) 的转换。
│   │       └── service.go
│   └── pkg/ # 基础库，部分抄录自 kratos 
│       ├── app/
│       │   └── app.go # 抄写自 kratos v2 app.go
│       ├── errors/ # 错误 struct 和 错误码
│       │   ├── code.go
│       │   └── errors.go
│       └── server/
│           ├── grpc/
│           │   └── errors.go # 错误码和 gRPC 错误码转换
│           ├── http/
│           │   ├── errors.go # 错误码和 HTTP 错误码转换
│           │   ├── middleware.go
│           │   └── server.go # HTTP 服务，借鉴了 kratos v2
│           └── server.go
└── test/ # 测试资源。目前没有使用
    ├── 0_db.sql # 测试数据表定义
    └── docker-compose.yaml # 启动开发/测试库
```
