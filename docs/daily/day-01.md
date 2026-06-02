# Day 01 学习记录

## 日期

2026-06-02

## 今日 Issue

- GitHub Issue：https://github.com/CanYangTang/go_learning/issues/1

## 核心任务

- 初始化 Go 后端学习项目。
- 理解 Go module、项目目录结构、服务入口和 Makefile。
- 验证最小 HTTP 服务可以运行。

## 可选挑战题

- 给 `/health` 响应增加 `version` 字段。

## 今日产出

- 初始化 Git 仓库和 Go module。
- 创建项目 README、学习计划、每日记录和周复盘模板。
- 创建最小 HTTP 服务入口：`cmd/server/main.go`。
- 创建统一响应包：`pkg/response`。
- 创建基础错误包：`pkg/apperror`。
- 创建常用命令入口：`Makefile`。
- `/health` 响应已包含 `status`、`service` 和 `version`。

## 运行过的命令

```bash
make fmt
make test
make vet
go run ./cmd/server
curl -sS http://localhost:8080/health
```

## 遇到的问题

- 初始环境中没有安装 Go 和 GitHub CLI，已通过 Homebrew 安装。
- 推送远程仓库前需要完成 GitHub CLI 认证。

## 答疑记录

- 问题：Makefile 文件中的内容做了什么，有什么用？
- 答案：Makefile 把常用命令封装成统一入口，例如 `make fmt`、`make test`、`make vet`、`make run`，用于格式化、测试、静态检查和启动服务。
- 问题：package 和文件有什么区别？
- 答案：文件是磁盘上的 `.go` 文件，负责存放代码；package 是 Go 的代码组织和编译单位，一个 package 通常由同一目录下多个 `.go` 文件共同组成。Go 命令、import、test、build 主要围绕 package 工作，而不是围绕单个文件工作。

## 今日小测试

1. `go.mod` 的作用是什么？
   - 回答：定义项目的 module path，项目内部的包都会通过这个路径导入。
   - 结果：通过。
2. `go fmt ./...` 中的 `./...` 表示什么？
   - 回答：`.` 是相对路径，这里是根目录；`...` 是代表所有目录和文件，还不完全清楚。
   - 结果：基本通过。
   - 补充：`./...` 表示当前目录及其所有子目录中的 Go package，不是所有文件。`go fmt` 只会处理这些 package 里的 Go 源文件。
3. `http.NewServeMux()` 和 `http.ListenAndServe()` 分别负责什么？
   - 回答：分别负责注册路由和启动服务并监听 8080 端口。
   - 结果：基本通过。
   - 补充：`http.NewServeMux()` 创建路由器，真正注册路由的是 `mux.HandleFunc()`；`http.ListenAndServe()` 负责监听端口并把请求交给 mux 处理。
4. 为什么 Makefile 中要写 `.PHONY`？
   - 回答：告诉 make 工具把这些字符当作命令来处理而不是文件名。
   - 结果：通过。
5. 当前 `/health` 接口返回了哪些字段？
   - 回答：`status`、`service`、`version`。
   - 结果：通过。

## 关键收获

1. Go 程序从 `package main` 中的 `main()` 函数开始执行。
2. `go.mod` 定义当前项目的 module path，项目内部包通过这个路径导入。
3. `net/http` 中 `http.NewServeMux()` 负责创建路由器，`mux.HandleFunc()` 负责注册路由，`http.ListenAndServe()` 负责监听端口并把请求交给路由器处理。
4. Makefile 可以把常用项目命令标准化，减少重复输入并方便后续接入 CI。
5. `.PHONY` 用来告诉 `make`：`fmt`、`test`、`vet`、`run` 是命令目标，不是文件名。
6. `./...` 表示当前目录及所有子目录中的 Go package，不是所有文件；`go fmt ./...` 会格式化这些 package 中的 Go 源文件。
7. 文件是物理单位，package 是代码组织和编译单位；Go 的 import、test、build 通常围绕 package 工作。

## 明日计划

- 进入 Day 2：练习变量、常量、基础类型和类型转换。
- 创建 `week01-basics/day02-syntax/` 目录。
- 写可运行的基础语法练习和测试。
