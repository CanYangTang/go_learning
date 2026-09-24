# Day 27 学习记录

## 日期

2026-09-23

## 今日 Issue

- GitHub Issue：https://github.com/CanYangTang/go_learning/issues/27

## 今日教案

- 教案文档：`docs/daily/day-27-lesson.md`

## 核心任务

- 补齐纯单元测试缺口：
  - `config.JWTSecret` / `config.AllowedOrigins`（`t.Setenv` + 默认回退 + 空白清洗）。
  - `middleware.fallbackRequestID`（0%→100%）、`newRequestID` happy path、`Logging` 无 request_id 分支、`UserIDFromContext` 三态。
- 刷新 `docs/api/todo-api.md`：对齐 Day 25 鉴权端点 + Day 26 统一信封，示例全部实抓。

## 范围决策

- **文档 + 纯单元测试缺口**。不碰需要真实 DB 的 `repository.*` / `config.ConnectGorm`（集成测试已覆盖），不追 `handler.Login` 的 token 失败分支和 `newRequestID` 的 rand 失败分支（防御性分支，为覆盖率改 production 代码得不偿失，见教案 §6）。
- 分工：**AI 全写测试和文档，用户审查**（今天角色对调；production 无待填空函数）。

## 可选挑战题

- 集成测试或 HTTP 请求示例（Issue 里标 optional）——本次以 curl 示例形式并入 API 文档。

## 答疑记录

- 本日无独立答疑,学完直接进实现(测试 + 文档)。

## Quiz

满分 5/5。

**Q1｜为何 `t.Setenv` 而非 `os.Setenv`?**
- 回答:`t.Setenv` 在 `os.Setenv` 基础上自动精确还原(含"原本不存在"),并禁止与 `t.Parallel` 共存,避免全局状态污染。
- 结果:✓ 正确。
- 标准答案:一致。测试结束由框架恢复原值(包括原本 unset 的就重新 unset);一旦调用即禁止本测试 `t.Parallel`(改全局 env 与并行互斥,Go 直接 panic 提示)。

**Q2｜为何用 `package middleware` 而非 `package middleware_test`?**
- 回答:白盒能访问 `fallbackRequestID`、`userIDContextKey` 等私有符号,改成 `middleware_test` 会编译报 undefined。
- 结果:✓ 正确。
- 标准答案:一致。同包(白盒)可测未导出符号;`xxx_test` 是黑盒,只能测导出 API,今天要测的私有函数/常量就够不着。

**Q3｜`TestUserIDFromContext` 三 case 各覆盖哪条路径?**
- 回答:分别覆盖"早退 `return 0,false`""断言成功 `return 7,true`""断言失败 `return 0,false`";uint/非 uint 对应类型断言 `ok` 的两种结果,缺一不达 100%。
- 结果:✓ 正确。
- 标准答案:一致。① key 不存在 → `if !exists` 早退;② 存 `uint` → 断言 `ok==true`;③ 存非 `uint` → 断言 `ok==false` 走末行 `return id,ok`。②③ 是同一行 `v.(uint)` 的两种结果,只测一个覆盖不满。

**Q4｜`Login` 80% / `newRequestID` 75% 为何不追?**
- 回答:未覆盖的是 crypto/rand 失败、token 签发失败这类现实几乎不可达的防御分支,拉满需注入 fake/mock、扭曲生产代码,代价大于收益。
- 结果:✓ 正确。
- 标准答案:一致。要覆盖须把 `*auth.Manager` 抽成 interface 注入"必失败"实现 —— 为测试扭曲 production 结构。覆盖率是工具不是 KPI,防御分支留着兜底本身即价值(YAGNI)。

**Q5｜为何 API 文档示例必须实抓、不能手编?**
- 回答:手编会导致 D1 那样的文档漂移(7 个不存在的接口、字段/结构/状态码对不上);实抓"抓不到就写不了",用运行时真实响应做存在性证明,杜绝虚构与过时。
- 结果:✓ 正确。
- 标准答案:一致。手编凭记忆/意图,与真实实现随时间漂移(D1 就写了 7 个不存在的接口、6 处形状不符);实抓强制"接口真存在、形状真如此"才写得出,是一种存在性证明。

## 今日产出

- `internal/config/env_test.go`(新):`JWTSecret`(默认回退 + 环境覆盖)、`AllowedOrigins`(默认 / 单值 / 空白空段清洗)。
- `internal/middleware/coverage_test.go`(新):`newRequestID`(hex32)、`fallbackRequestID`、`UserIDFromContext`(三态)、`Logging` 无 request_id 分支。
- `docs/api/todo-api.md`:新增「快速上手(可直接粘贴的 curl)」全流程示例,token 用 shell 变量提取;实抓逐条比对 8 个端点形状全相符,更新核对日期为 Day 27。
- 覆盖率:目标函数全部 100%;总覆盖率 76.9% → 81.3%。

## 运行过的命令

```bash
gofmt -l .                                   # clean
go build ./... && go vet ./...               # OK
go test -count=1 ./internal/... ./pkg/...    # 全绿
go test -count=1 -coverprofile=cov.out ./internal/... ./pkg/...
go tool cover -func=cov.out | awk '$3 != "100.0%"'   # 只剩 DB/防御分支
# 集成模式确认 repository 已被覆盖：
RUN_INTEGRATION_TESTS=true go test -count=1 -coverprofile=covi.out ./internal/repository/ ./internal/config/
# API 文档示例：go run ./cmd/server 后逐端点 curl 实抓
```

## 遇到的问题

- 起服务抓文档时 8080 被上次 e2e 的进程占用(`bind: address already in use`)→ `lsof -ti tcp:8080 | xargs kill` 释放后重启。
- 原以为 API 文档停在旧形状,实抓比对后发现 Day 25 已更新到最新;真正的缺口是没有可粘贴的 curl 示例,遂聚焦补示例而非重写。

## 关键收获

1. `t.Setenv` 是测 env 依赖函数的正确工具:自动恢复 + 禁并行,`os.Setenv` 会泄漏脏状态。
2. 覆盖率是定位工具不是 KPI:`go tool cover -func/-html` 找红块补分支,但防御性分支(rand/签名失败)不值得为触发它扭曲 production 代码。
3. 文档示例实抓是一种"存在性证明":抓不到就写不出,从机制上杜绝 D1 式的文档漂移。


## 明日计划

- 配置迁环境变量收尾（DB 密码 → env、`.env.example`、连接池 A4/B1/D5/E2）；工程收尾 C3/C6（`make` 加 `-count=1`/`goimports`/lint、清早期遗留 TODO 注释）；C5 集成测试 teardown 的 `DROP TABLE` 收敛。
