# Zero API

`apps/api` 是 Zero 项目的后端服务，基于 `go-zero` 搭建，负责登录注册、今日打卡、觉察记录、复盘、每周投票、每周讨论和后台管理等核心业务接口。

## 项目定位

这个服务当前承担两类职责：

- 面向用户侧页面的业务接口：今天做什么、今天记录了什么、后续什么时候复盘、本周投票和讨论是什么。
- 面向后台管理的配置接口：主题排期、投票周期、讨论信息、后台概览和管理员登录。

接口统一挂在前缀 `http://<host>:<port>/api/v1` 下，默认本地地址是 `http://localhost:8888/api/v1`。

---

## 架构总览

当前代码的主链路是：

`zero.api` → `handler` → `logic` → `svc` → `model` → MySQL

### 1. 接口契约层

- 文件：`zero.api`
- 作用：定义接口路径、请求体、响应体和主要数据结构。
- 说明：这是整个后端的契约入口，接口字段和返回结构优先以这里为准。

### 2. Handler 层

- 目录：`internal/handler`
- 作用：注册路由、解析请求、调用对应的 logic。
- 说明：`routes.go` 负责把所有接口统一注册到 `/api/v1` 前缀下。

### 3. Logic 层

- 目录：`internal/logic`
- 作用：承载业务规则，是当前项目最核心的一层。
- 说明：
  - 登录注册、今日打卡、觉察记录、复盘排期、周投票、讨论生成、后台管理都在这一层完成。
  - 这里既包含接口级 logic，也包含跨接口复用的业务帮助函数，比如：
    - `review_cycle.go`
    - `review_schedule.go`
    - `weekly_flow.go`
    - `repository_helpers.go`
    - `admin_helpers.go`

### 4. Service Context 层

- 文件：`internal/svc/servicecontext.go`
- 作用：统一初始化数据库连接和各张表对应的 model。
- 说明：
  - 配置了 `Mysql.DataSource` 时，走真实 MySQL。
  - 未配置 `Mysql.DataSource` 时，不初始化数据库连接，部分 logic 会回退到本地演示数据。

### 5. Model 层

- 目录：`model`
- 作用：封装各张表的数据访问。
- 说明：当前已经覆盖用户、管理员、主题、打卡、日志、复盘、投票、讨论等核心表。

---

## 目录说明

### 根目录关键文件

- `zero.go`：服务启动入口。
- `zero.api`：接口契约定义。
- `go.mod` / `go.sum`：Go 依赖管理。
- `README.md`：当前说明文档。

### 关键目录

- `cmd/initdb`：数据库初始化命令，会读取 SQL 文件并依次执行。
- `etc`：服务配置目录。
- `internal/config`：配置结构定义。
- `internal/handler`：HTTP 路由和请求入口。
- `internal/logic`：业务逻辑主目录。
- `internal/svc`：ServiceContext 初始化。
- `internal/types`：接口请求/响应类型。
- `model`：数据库 model。

---

## 当前核心功能说明

### 1. 用户认证与账号体系

对应能力：

- 密码登录：`POST /auth/login`
- 验证码登录：`POST /auth/code-login`
- 注册：`POST /auth/register`
- 发送验证码：`POST /auth/verify-codes`
- 刷新令牌：`POST /auth/refresh-token`
- 退出登录：`POST /auth/logout`
- 获取当前用户：`GET /me`

相关逻辑主要在：

- `passwordloginlogic.go`
- `codeloginlogic.go`
- `registerlogic.go`
- `sendverifycodelogic.go`
- `refreshtokenlogic.go`
- `logoutlogic.go`

补充说明：

- 当前密码使用 `sha256` 做哈希，逻辑位于 `internal/logic/auth_helpers.go`。
- 当前登录响应仍偏开发态，返回的是开发占位 token；后续若接正式鉴权体系，需要继续补强。

### 2. 今日打卡

对应能力：

- 创建今日打卡：`POST /daily-tasks`
- 获取打卡列表：`GET /daily-tasks`
- 获取单条打卡：`GET /daily-tasks/:id`
- 更新打卡内容：`PATCH /daily-tasks/:id`
- 提交打卡：`POST /daily-tasks/:id/submit`
- 获取“我的今日打卡”：`GET /me/today-task`

相关逻辑主要在：

- `createdailytasklogic.go`
- `listdailytaskslogic.go`
- `getdailytasklogic.go`
- `updatedailytasklogic.go`
- `submitdailytasklogic.go`
- `getmytodaytasklogic.go`

业务要点：

- 今日打卡围绕当天主题生成。
- 提交后会进入后续复盘节奏。
- 逻辑里已经处理“可编辑”“可补反思”等状态字段。

### 3. 觉察记录 / 日志

对应能力：

- 获取单条日志：`GET /daily-logs/:id`
- 更新单条日志：`PATCH /daily-logs/:id`
- 获取某条打卡下的日志列表：`GET /daily-tasks/:id/logs`
- 新建某条打卡下的日志：`POST /daily-tasks/:id/logs`

相关逻辑主要在：

- `getdailyloglogic.go`
- `updatedailyloglogic.go`
- `listdailytasklogslogic.go`
- `createdailytaskloglogic.go`

业务要点：

- 觉察记录和当天打卡主题关联。
- 日志内容会被首页、今日页、日志页等多个用户侧页面回读。

### 4. 复盘体系

对应能力：

- 获取待复盘列表：`GET /review-items`
- 获取单条复盘项：`GET /review-items/:id`
- 创建复盘记录：`POST /review-items/:id/records`
- 获取复盘记录列表：`GET /review-items/:id/records`
- 获取复盘历史：`GET /review-records`
- 创建补做复盘记录：`POST /review-recovery-records`

相关逻辑主要在：

- `listreviewitemslogic.go`
- `getreviewitemlogic.go`
- `createreviewrecordlogic.go`
- `listreviewrecordslogic.go`
- `listreviewhistorylogic.go`
- `createrecoveryreviewlogic.go`
- `review_cycle.go`
- `review_schedule.go`
- `review_submission.go`

业务要点：

- 复盘不是单次动作，而是按周期推进。
- 当前项目已经把“待复盘项生成”“复盘阶段推进”“历史回看”纳入后端逻辑。

### 5. 首页聚合 / 用户概览

对应能力：

- 首页聚合：`GET /me/home`

相关逻辑：

- `gethomelogic.go`
- `cycle_summary.go`
- `reinforcement_hints.go`

业务要点：

- 首页接口负责把今日打卡、最近觉察、待复盘、补做复盘等信息聚合给前端首页使用。

### 6. 每周投票与讨论

对应能力：

- 获取当前投票：`GET /weekly-votes/current`
- 提交当前投票：`POST /weekly-votes/current/records`
- 获取当前讨论：`GET /discussions/current`
- 获取主题列表：`GET /topics`

相关逻辑主要在：

- `getcurrentweeklyvotelogic.go`
- `createcurrentweeklyvoterecordlogic.go`
- `getcurrentdiscussionlogic.go`
- `listtopicslogic.go`
- `weekly_flow.go`
- `weekly_repository.go`

业务要点：

- 周投票面向“下一个周期”的主题候选。
- 当前讨论主题与投票结果、周排期相关。
- 讨论与投票已经形成一条后端闭环，而不是孤立页面。

### 7. 后台管理

对应能力：

- 管理员登录：`POST /admin/auth/login`
- 后台概览：`GET /admin/stats/overview`
- 主题管理：
  - `GET /admin/topics`
  - `POST /admin/topics`
  - `PATCH /admin/topics/:id`
- 讨论管理：
  - `GET /admin/discussions`
  - `POST /admin/discussions`
  - `PATCH /admin/discussions/:id`
- 周投票管理：
  - `GET /admin/weekly-votes`
  - `POST /admin/weekly-votes`
  - `PATCH /admin/weekly-votes/:id`

相关逻辑主要在：

- `adminloginlogic.go`
- `adminstatsoverviewlogic.go`
- `adminlisttopicslogic.go`
- `admincreatetopiclogic.go`
- `adminupdatetopiclogic.go`
- `adminlistdiscussionslogic.go`
- `admincreatediscussionlogic.go`
- `adminupdatediscussionlogic.go`
- `adminlistweeklyvoteslogic.go`
- `admincreateweeklyvotelogic.go`
- `adminupdateweeklyvotelogic.go`
- `admin_guard.go`
- `admin_helpers.go`

业务要点：

- 后台不是简单 CRUD，而是围绕主题时间线、讨论配置、投票周期来维护用户侧内容。

---

## 接口分组速查表

> 下面所有路径都默认带前缀：`/api/v1`

### 用户认证

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/auth/login` | 用户密码登录 |
| POST | `/auth/code-login` | 用户验证码登录 |
| POST | `/auth/register` | 用户注册 |
| POST | `/auth/verify-codes` | 发送验证码 |
| POST | `/auth/refresh-token` | 刷新登录令牌 |
| POST | `/auth/logout` | 退出登录 |

### 当前用户与首页聚合

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/me` | 获取当前用户信息 |
| GET | `/me/home` | 获取首页聚合数据 |
| GET | `/me/today-task` | 获取“我的今日打卡” |

### 今日打卡

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/daily-tasks` | 创建今日打卡 |
| GET | `/daily-tasks` | 获取打卡列表 |
| GET | `/daily-tasks/:id` | 获取单条打卡 |
| PATCH | `/daily-tasks/:id` | 更新打卡内容 |
| POST | `/daily-tasks/:id/submit` | 提交打卡 |

### 觉察记录 / 日志

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/daily-logs/:id` | 获取单条日志 |
| PATCH | `/daily-logs/:id` | 更新单条日志 |
| GET | `/daily-tasks/:id/logs` | 获取某条打卡下的日志列表 |
| POST | `/daily-tasks/:id/logs` | 在某条打卡下新增日志 |

### 复盘体系

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/review-items` | 获取待复盘项列表 |
| GET | `/review-items/:id` | 获取单条复盘项 |
| POST | `/review-items/:id/records` | 提交复盘记录 |
| GET | `/review-items/:id/records` | 获取某个复盘项的记录列表 |
| GET | `/review-records` | 获取复盘历史列表 |
| POST | `/review-recovery-records` | 提交补做复盘记录 |

### 每周主题 / 投票 / 讨论

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/topics` | 获取主题列表 |
| GET | `/weekly-votes/current` | 获取当前周投票 |
| POST | `/weekly-votes/current/records` | 提交当前周投票 |
| GET | `/discussions/current` | 获取当前讨论信息 |

### 管理后台

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/admin/auth/login` | 管理员登录 |
| GET | `/admin/stats/overview` | 获取后台概览 |
| GET | `/admin/topics` | 获取主题排期列表 |
| POST | `/admin/topics` | 新建主题 |
| PATCH | `/admin/topics/:id` | 更新主题 |
| GET | `/admin/discussions` | 获取讨论配置列表 |
| POST | `/admin/discussions` | 新建讨论配置 |
| PATCH | `/admin/discussions/:id` | 更新讨论配置 |
| GET | `/admin/weekly-votes` | 获取周投票配置列表 |
| POST | `/admin/weekly-votes` | 新建周投票配置 |
| PATCH | `/admin/weekly-votes/:id` | 更新周投票配置 |
## 当前数据层覆盖范围

当前 `model` 已覆盖以下核心表：

- `users`
- `user_profiles`
- `admin_users`
- `auth_verification_codes`
- `refresh_tokens`
- `topics`
- `daily_tasks`
- `daily_logs`
- `review_items`
- `review_records`
- `weekly_topic_votes`
- `weekly_topic_vote_candidates`
- `weekly_topic_vote_records`
- `discussion_infos`
- `operation_logs`

对应 model 文件都在：`model`

---

## 配置说明

配置文件默认路径：`etc/zero-api.yaml`

示例模板：`etc/zero-api.example.yaml`

### 最小配置示例

```yaml
Name: zero-api
Host: 0.0.0.0
Port: 8888

Mysql:
  DataSource: "your_user:your_password@tcp(127.0.0.1:3306)/zero_app?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai"
```

### 配置项说明

- `Host` / `Port`：服务监听地址。
- `Mysql.DataSource`：MySQL 连接串。
- `Cycle`：复盘周期相关配置结构已预留在 `internal/config/config.go`，当前 README 不把它当必须配置项。

### MySQL 注意事项

- 数据库名默认使用 `zero_app`
- 必须开启 `parseTime=true`
- 字符集建议使用 `utf8mb4`
- 时区建议使用 `Asia/Shanghai`

---

## 数据库初始化

初始化命令：

```bash
cd apps/api
go run ./cmd/initdb -f ./etc/zero-api.yaml
```

### initdb 的行为

`cmd/initdb` 会：

- 读取配置中的 `Mysql.DataSource`
- 自动去掉 DSN 中的数据库名，先连到 MySQL 实例
- 按文件名排序执行 SQL 文件
- 依次执行当前工作区根目录下 `docs/sql/` 里的 `.sql` 文件

### 重要说明

当前 `docs` 目录不属于 `api` 仓库本身，而是本地工作区资料。
如果你把 `apps/api` 单独作为仓库使用，需要保证本地仍然提供等价的 SQL 初始化文件，否则 `initdb` 找不到 `docs/sql`。

---

## 启动服务

```bash
cd apps/api
go run zero.go -f etc/zero-api.yaml
```

启动后默认访问：

- `http://localhost:8888/api/v1`

---

## 鉴权与上下文注入说明

当前项目的“当前用户 / 当前管理员”是通过请求头注入上下文：

- `X-User-Id`
- `X-Admin-Id`

中间件文件：`internal/handler/currentusermiddleware.go`

### 当前行为

- 如果请求头里有合法 `X-User-Id`，会注入到上下文里。
- 如果没有显式注入用户 ID，普通用户链路会回退到默认用户 `1`。
- 管理员链路不会回退默认管理员，必须明确带管理员上下文。

### 这意味着什么

- 这套机制当前更偏“开发联调态”而不是正式生产态鉴权。
- 如果后续上生产，需要继续补真正的 token 校验 / session 校验，而不是长期依赖默认用户回退。

---

## 演示数据 / 无库模式说明

当 `Mysql.DataSource` 为空时，`ServiceContext` 不会初始化数据库连接。
此时部分接口会回退到本地演示数据逻辑，方便前端联调。

适用场景：

- 前端页面结构开发
- 页面文案和状态流转联调
- 不想每次都先起数据库时的本地占位测试

不适用场景：

- 正式数据验证
- 权限校验验证
- 投票、复盘、后台配置等真实写入链路验证

---

## 测试与验证

运行所有 Go 测试：

```bash
cd apps/api
go test ./...
```

当前 `internal/logic` 下已经补了较多业务测试，重点覆盖：

- 今日打卡创建与访问
- 当前用户 / 管理员上下文
- 复盘周期与排期
- 周投票流程
- 部分后台守卫逻辑

---

## 目前最值得优先关注的文件

如果后续继续开发，这几个位置最关键：

- `zero.api`
- `internal/handler/routes.go`
- `internal/svc/servicecontext.go`
- `internal/logic/review_cycle.go`
- `internal/logic/review_schedule.go`
- `internal/logic/weekly_flow.go`
- `internal/logic/repository_helpers.go`
- `internal/logic/gethomelogic.go`

---

## 开发注意事项

- 这是个人项目后端，优先保证主链路闭环，不要过度工程化。
- 现阶段允许开发态过渡实现，但 README 中提到的“默认用户回退”“开发 token”都应视为后续待补强点。
- 改接口时，优先先看 `zero.api`，再看对应 `handler`、`logic`、`types`、`model` 是否需要同步。
- 改数据库相关逻辑时，先确认 `model`、`repository_helpers.go` 和对应测试是否需要一起更新。