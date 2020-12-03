作业说明
---
执行入口 [serve](./cmd/homework/serve.go)  
业务代码 简单的员工查询。[staff](./homework/staff)  
[Kit](./homework/kit)

GET 请求 8080 端口，`/staff/:id` 。  
id为-1时，触发 panic 一个 withStack 的错误；-2时，触发一个运行时数组访问越界 panic，没有堆栈。  
id为1、2时有值，其他id无值。