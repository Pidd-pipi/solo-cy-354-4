# CampusMarket（校园二手交易平台）

一款面向高校学生的校内 C2C 交易平台，覆盖闲置物品发布、价格协商私信、交易达成确认、信誉评分举报、毕业季专场与书籍交换等场景。

## 快速启动（Docker Compose 一键部署）

```bash
cp .env.example .env
docker compose up -d --build
```

启动后访问：

- 前端：http://localhost:28514
- 后端健康检查：http://localhost:29514/healthz
- MySQL：localhost:3306

预置账号（database/init.sql 种子数据）：

| 角色 | 手机号 | 密码 |
| --- | --- | --- |
| 学生（东校区） | 13700000001 | 123456 |
| 学生（西校区） | 13700000002 | 123456 |
| 学生（南校区） | 13700000003 | 123456 |
| 管理员 | 13800000001 | admin123 |

停止并清理（删除数据卷）：

```bash
docker compose down -v --remove-orphans
```

## 本地开发

后端（Go 1.22）：

```bash
cd backend
go mod tidy
go run ./cmd/server
go build ./...
go test ./...
```

前端（Vue 3 + Vite）：

```bash
cd frontend
npm install
npm run dev
npm run build
```

本地开发时前端 Vite 将 `/api` 代理到 `http://localhost:29514`。

## 技术栈

| 端 | 技术 |
| --- | --- |
| 前端 | Vue 3 + TypeScript + Element Plus + Vite |
| 后端 | Go 1.22 + Gin + GORM |
| 数据库 | MySQL 8.0 |
| 认证 | JWT（golang-jwt/jwt/v5）+ RBAC + bcrypt |
| 其他 | go-playground/validator/v10、log/slog、gin-contrib/cors |

## 项目目录结构

```
cy-354/
├── docker-compose.yml
├── .env.example
├── README.md
├── database/
│   └── init.sql             # MySQL 首启初始化（建表 + 种子数据）
├── backend/
│   ├── go.mod
│   ├── Dockerfile
│   ├── cmd/server/          # main.go + seed.go
│   └── internal/
│       ├── config/          # 环境变量配置
│       ├── constants/       # product.go, trade.go, user.go, error_codes.go, log_templates.go, messages.go
│       ├── model/           # user, product, conversation, message, trade_order, review, book_exchange
│       ├── repository/      # GORM 仓库（按实体分文件）
│       ├── service/         # 业务逻辑（按实体分文件）
│       ├── handler/         # HTTP 处理器（按实体分文件）
│       ├── router/          # router.go + 按实体路由文件
│       ├── middleware/      # auth, rbac, rate_limiter, error_handler, request_id
│       ├── dto/             # 请求/响应结构体
│       └── util/            # jwt, logger, formatters, app_error, credit_calculator, response
└── frontend/
    ├── Dockerfile
    ├── nginx.conf
    └── src/
        ├── api/             # user, product, conversation, tradeOrder, review, bookExchange
        ├── stores/          # authStore, userStore, productStore, tradeStore
        ├── components/common/# ProductCard, ProductForm, MessageBubble, TradeStatusBadge, ExchangeCard
        ├── hooks/           # useAuth, useProducts, useConversations
        ├── pages/           # Products, Publish, Messages, Orders, BookExchange, Graduation, Profile, Login, Register
        ├── router/          # index.ts + guards.ts
        ├── utils/           # request, dateFormat, priceFormatter
        ├── constants/       # product, trade, user, errorCodes
        └── types/           # 共享类型
```

## 环境变量

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| COMPOSE_PROJECT_NAME | lpcampusmarket | Compose 项目名/容器前缀 |
| DB_NAME | lpcampusmarket_db | 数据库名 |
| DB_USER | lpcampusmarket_user | 数据库用户 |
| DB_PASSWORD | lpcampusmarket_pwd | 数据库密码 |
| DB_ROOT_PASSWORD | lpcampusmarket_root | root 密码 |
| JWT_SECRET | change_me_to_a_long_random_string | JWT 签名密钥（生产必须修改） |
| JWT_EXPIRE_HOURS | 72 | Token 有效期（小时） |
| RATE_LIMIT_PER_MIN | 120 | 普通接口限流（次/分钟） |
| LOGIN_RATE_LIMIT_PER_MIN | 10 | 登录/注册限流（次/分钟） |
| SEEDING_ENABLED | true | 是否启动时播种数据 |
| CORS_ORIGINS | http://localhost:28514,http://localhost:5173 | 允许跨域来源（逗号分隔；生产严禁 `*`） |
| FRONTEND_PORT | 28514 | 前端端口 |
| BACKEND_PORT | 29514 | 后端端口 |
| DB_PORT | 3306 | MySQL 端口 |

## Docker 部署说明

- 端口映射：前端 `28514:80`，后端 `${BACKEND_PORT:-29514}:8080`，数据库 `${DB_PORT:-3306}:3306`。
- 数据持久化：命名卷 `mysql_data` 挂载到 `/var/lib/mysql`；`database/init.sql` 在首次启动自动执行建表与种子数据。
- 健康检查：db 使用 `mysqladmin ping`，backend 使用 `/healthz`，frontend 依赖 backend healthy。
- 常见问题：
  - 端口冲突：修改 `.env` 中的 `FRONTEND_PORT`/`BACKEND_PORT`/`DB_PORT`。
  - 数据重置：`docker compose down -v` 后重新 `up -d`。
  - 中文目录名：Compose 通过项目名与容器名隔离，任意目录下均可启动。

## API 说明

- 统一前缀 `/api/v1`，健康检查 `/healthz`。
- 响应格式：`{ "code": 0, "message": "ok", "data": ... }`，错误码见 `backend/internal/constants/error_codes.go`。
- 核心接口：
  - `POST /api/v1/users/register`、`POST /api/v1/users/login`、`GET/PUT /api/v1/users/me`
  - `GET/POST /api/v1/products`、`GET/DELETE /api/v1/products/:id`、`GET /api/v1/products/graduation`
  - `POST /api/v1/conversations`、`GET /api/v1/conversations/me`、`GET/POST /api/v1/conversations/:id/messages`
  - `POST /api/v1/trade-orders`、`GET /api/v1/trade-orders/me`、`POST /api/v1/trade-orders/:id/buyer-confirm|seller-confirm|cancel`
  - `POST /api/v1/reviews`、`GET /api/v1/reviews/me`
  - `GET/POST /api/v1/book-exchanges`、`POST /api/v1/book-exchanges/:id/close`
  - `GET /api/v1/admin/stats`（管理员）

## API 接口清单

统一前缀 `/api/v1`；鉴权列中「登录」表示需要 JWT，「管理员」表示需要管理员角色。

| 方法 | 路径 | 说明 | 鉴权 |
| --- | --- | --- | --- |
| GET | `/healthz` | 健康检查 | 无 |
| POST | `/api/v1/users/register` | 注册学生账号 | 无（登录限流） |
| POST | `/api/v1/users/login` | 登录获取 JWT | 无（登录限流） |
| GET | `/api/v1/users/me` | 当前用户信息 | 登录 |
| PUT | `/api/v1/users/me` | 更新昵称/头像/校区 | 登录 |
| GET | `/api/v1/products` | 商品分页列表 | 无 |
| GET | `/api/v1/products/graduation` | 毕业季专场列表 | 无 |
| GET | `/api/v1/products/:id` | 商品详情 | 无 |
| POST | `/api/v1/products` | 发布商品 | 登录 |
| DELETE | `/api/v1/products/:id` | 下架自己的商品 | 登录 |
| POST | `/api/v1/conversations` | 发起/复用私信会话 | 登录 |
| GET | `/api/v1/conversations/me` | 我的会话列表 | 登录 |
| GET | `/api/v1/conversations/:id/messages` | 会话消息记录 | 登录 |
| POST | `/api/v1/conversations/:id/messages` | 发送私信 | 登录 |
| POST | `/api/v1/trade-orders` | 创建购买订单 | 登录 |
| GET | `/api/v1/trade-orders/me` | 我的订单列表 | 登录 |
| POST | `/api/v1/trade-orders/:id/buyer-confirm` | 买家确认 | 登录 |
| POST | `/api/v1/trade-orders/:id/seller-confirm` | 卖家确认（订单完成+商品售出） | 登录 |
| POST | `/api/v1/trade-orders/:id/cancel` | 取消订单 | 登录 |
| POST | `/api/v1/reviews` | 交易后评价（含信誉积分） | 登录 |
| GET | `/api/v1/reviews/me` | 我收到的评价 | 登录 |
| GET | `/api/v1/book-exchanges` | 书籍交换列表 | 无 |
| POST | `/api/v1/book-exchanges` | 发布换书请求（自动匹配） | 登录 |
| POST | `/api/v1/book-exchanges/:id/close` | 关闭换书请求 | 本人 |
| GET | `/api/v1/admin/stats` | 平台统计占位接口 | 管理员 |

## 枚举出现位置清单

### ProductStatus（on_sale/reserved/sold/removed）

前端 `frontend/src/constants/product.ts`：

- `PRODUCT_STATUSES` 常量定义
- `productStatusLabel()` / `productStatusType()` 映射
- `src/components/common/ProductCard.vue` 状态徽章与购买按钮显隐
- `src/pages/Orders.vue` 交易联动

后端 `backend/internal/constants/product.go`：

- `ProductStatusOnSale/Reserved/Sold/Removed` 常量
- `ProductStatuses` 列表、`IsProductStatus()`
- `ProductStatusText()` 文案
- `backend/internal/model/product.go` Status 字段
- `backend/internal/service/product_service.go` 发布/下架/售出状态机
- `backend/internal/util/formatters.go` `ProductStatusText()`
- `backend/internal/constants/log_templates.go` 商品状态日志模板
- `backend/internal/constants/error_codes.go` 状态冲突错误码

### TradeStatus（pending/confirmed/completed/cancelled）

前端 `frontend/src/constants/trade.ts`：

- `TRADE_STATUSES` 常量定义
- `tradeStatusLabel()` / `tradeStatusType()` 映射
- `src/components/common/TradeStatusBadge.vue` 状态徽章
- `src/pages/Orders.vue` 按钮显隐（确认收货/确认收款/取消/评价）

后端 `backend/internal/constants/trade.go`：

- `TradeStatusPending/Confirmed/Completed/Cancelled` 常量
- `TradeStatuses` 列表、`IsTradeStatus()`
- `TradeStatusText()` 文案
- `backend/internal/model/trade_order.go` Status 字段
- `backend/internal/service/trade_order_service.go` 交易状态机
- `backend/internal/util/formatters.go` `TradeStatusText()`
- `backend/internal/constants/log_templates.go` 交易日志模板
- `backend/internal/constants/error_codes.go` 状态冲突错误码

### UserRole（student/admin）

前端 `frontend/src/constants/user.ts`：

- `USER_ROLES` 常量定义
- `roleLabel()` 映射
- `src/stores/authStore.ts` `isAdmin()`
- `src/hooks/useAuth.ts` `hasRole()`
- `src/pages/Profile.vue` 角色展示

后端 `backend/internal/constants/user.go`：

- `UserRoleStudent/Admin` 常量
- `UserRoles` 列表、`IsUserRole()`
- `UserRoleText()` 文案
- `backend/internal/model/user.go` Role 字段
- `backend/internal/middleware/rbac.go` 权限校验
- `backend/internal/router/*.go` 路由权限（管理员接口）
- `backend/internal/util/jwt.go` Claims.Role
- `backend/internal/util/formatters.go` `RoleText()`
- `backend/internal/constants/log_templates.go` 登录日志带角色

## 质量说明

- 后端 `go build ./...` 与 `go test ./...` 通过（含 service/util 表驱动单测）。
- 前端 `npm run build` 零错误。
- 分层依赖单向：handler → service → repository → model；构造器注入；`%w` 错误链 + 哨兵错误；统一响应 `{code,message,data}`。
- 日志模板集中于 `internal/constants/log_templates.go`（≥25 条），全栈引用，字段变更需联动修改（屎山设计约束）。

## License

MIT
