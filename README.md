# Bubble Admin Go Kratos

[![Go Report Card](https://goreportcard.com/badge/github.com/sober-studio/bubble-admin-go-kratos)](https://goreportcard.com/report/github.com/sober-studio/bubble-admin-go-kratos)
[![License](https://img.shields.io/github/license/sober-studio/bubble-admin-go-kratos)](LICENSE)

基于 [Kratos](https://go-kratos.dev/) 微服务框架构建的现代化 Go 语言服务端管理后台脚手架。本项目集成了常见的后端开发组件与最佳实践，旨在帮助开发者快速搭建稳定、可扩展的后台服务。

## ✨ 特性

- **微服务架构**: 基于 Kratos v2，支持 gRPC 与 HTTP 双协议，遵循 clean architecture。
- **依赖注入**: 使用 [Wire](https://github.com/google/wire) 进行依赖注入，代码结构清晰，易于维护和测试。
- **ORM 框架**: 集成 [GORM](https://gorm.io/)，支持 MySQL/PostgreSQL，配合 GORM Gen 自动生成类型安全的查询代码。
- **认证鉴权**:
  - 完善的 JWT 认证流程（登录、注册、刷新、撤销）。
  - 基于 RBAC（Role-Based Access Control）的权限管理体系。
- **第三方服务集成**:
  - **对象存储 (OSS)**: 支持阿里云、七牛云、MinIO 及本地存储，接口统一，易于切换。
  - **短信服务 (SMS)**: 支持阿里云短信，模块化设计，易于扩展其他供应商。
  - **邮件服务**: 支持 SMTP 邮件发送，支持 HTML 模板。
- **任务调度**: 内置 Cron 定时任务管理，支持分布式环境下的任务调度。
- **实时通信**: 内置 WebSocket 服务支持，适用于消息推送、即时聊天等场景。
- **开发工具**: 包含 Protobuf 代码生成、Wire 依赖注入生成、GORM 代码生成等自动化脚本。

## 🛠 技术栈

- **Language**: [Go](https://golang.org/) (1.20+)
- **Framework**: [Kratos](https://github.com/go-kratos/kratos) v2
- **Database**: PostgreSQL / MySQL
- **Cache**: Redis
- **ORM**: [GORM](https://gorm.io/)
- **DI**: [Wire](https://github.com/google/wire)
- **API Definition**: Protobuf / OpenAPI

## 🚀 快速开始

### 前置要求

确保本地已安装以下环境：

- Go 1.20+
- Docker & Docker Compose
- Make
- Git

### 1. 克隆项目

```bash
git clone https://github.com/sober-studio/bubble-admin-go-kratos.git
cd bubble-admin-go-kratos
```

### 2. 初始化环境

安装必要的开发工具（protoc 插件、wire 等）：

```bash
make init
```

### 3. 配置环境

复制环境变量示例文件：

```bash
cp .env.example .env
```

根据需要修改 `.env` 中的配置（如数据库密码、端口等）。

### 4. 启动基础设施

使用 Docker Compose 启动 PostgreSQL 和 Redis：

```bash
make dev
```

### 5. 初始化数据库

将 `scripts/db/init.sql` 导入到数据库中。

```bash
# 示例：使用 docker exec 导入（需确保容器正在运行）
# 注意：容器名称取决于 .env 中的 COMPOSE_PROJECT_NAME，默认为 bubble-admin-go-kratos_pgsql
cat scripts/db/init.sql | docker exec -i bubble-admin-go-kratos_pgsql psql -U postgres -d bubble_admin
```

### 6. 生成代码

如果这是第一次运行，或者修改了 proto 文件/数据库模型，建议重新生成代码：

```bash
make all
make gormgen
```

### 7. 运行服务

```bash
go run ./cmd/bubble-admin-go-kratos/ -conf configs
```

服务启动后，默认监听端口：
- HTTP: `8000`
- gRPC: `9000`

你可以通过浏览器访问 `http://localhost:8000/api/public/v1/hello` (假设有此接口) 或使用 Postman 测试 API。

## 📂 目录结构

遵循 Kratos 官方 Layout：

```text
├── api/             # API 定义 (Protobuf 文件)
├── cmd/             # 程序入口 (main.go)
├── configs/         # 配置文件 (config.yaml)
├── internal/
│   ├── biz/         # 业务逻辑层 (Domain Layer) - 定义接口和业务实体
│   ├── conf/        # 配置定义 (Protobuf)
│   ├── data/        # 数据访问层 (Data Layer) - 实现 biz 接口，操作 DB/Redis
│   ├── server/      # 服务启动 (HTTP/gRPC Server 配置)
│   ├── service/     # 应用服务层 (Application Layer) - 处理 DTO 转换，调用 biz
│   └── pkg/         # 公共包 (工具类、中间件等)
├── third_party/     # 第三方 proto 文件
└── scripts/         # 数据库初始化脚本等
```

## 📝 开发指南

### 添加新 API

1. **定义 API**: 在 `api/` 目录下创建或修改 `.proto` 文件，定义 Service 和 Message。
2. **生成代码**: 运行 `make api` 生成 Go 接口代码。
3. **实现 Service**: 在 `internal/service` 实现生成的接口，处理请求参数，调用 Biz 层。
4. **实现 Biz**: 在 `internal/biz` 定义业务领域模型和接口，编写核心业务逻辑。
5. **实现 Data**: 在 `internal/data` 实现 Biz 层定义的接口，进行数据库或缓存操作。
6. **注册服务**: 在 `internal/server` 将 Service 注册到 HTTP/gRPC Server。

### 数据库变更

1. **修改模型**: 修改 `internal/data/model` 下的 GORM 模型结构体。
2. **生成 Query**: 运行 `make gormgen` 基于模型生成类型安全的查询代码。
3. **更新数据库**: 编写 SQL 迁移脚本或手动更新数据库表结构（开发环境可使用 GORM AutoMigrate，但不推荐用于生产）。

## 🗺️ Roadmap

- ✅ JWT 认证（支持 token 撤销）
- ✅ 短信服务（支持阿里云等）
- ✅ 邮件服务（SMTP）
- ✅ 对象存储服务（支持阿里云、七牛云、MinIO、本地存储等）
- ✅ 定时任务
- ✅ WebSocket 服务
- ……

## 📄 License

[MIT](LICENSE)
