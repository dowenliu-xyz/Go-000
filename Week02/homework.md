作业说明
---
执行入口 [serve](./cmd/homework/serve.go)  
业务代码 简单的员工查询。[staff](./homework/staff)  
[Kit](./homework/kit)

GET 请求 8080 端口，`/staff/:id` 。id为负时，触发panic；id为1、2时有值，其他id无值。