# MXShop 微服务商城项目

这是一个基于 Go 语言的微服务电商项目，适合学习微服务架构和 Go 后端开发。

## 项目状态

⚠️ **当前版本为学习版本**：已保留完整的基础框架和工具库，但删除了具体的业务实现代码。你需要跟着教程逐步实现各个服务的业务逻辑。

## 项目架构

```
mxshop_refactor/
├── api/                 # Protobuf API 定义
│   ├── user/           # 用户服务 API
│   ├── goods/          # 商品服务 API
│   ├── order/          # 订单服务 API
│   └── inventory/      # 库存服务 API
├── app/                 # 应用层
│   ├── user/           # 用户微服务 ⚠️ 需实现
│   ├── goods/          # 商品微服务 ⚠️ 需实现
│   ├── order/          # 订单微服务 ⚠️ 需实现
│   ├── inventory/      # 库存微服务 ⚠️ 需实现
│   ├── mxshop/         # API 网关 ⚠️ 需实现
│   └── pkg/            # 应用层公共代码 ✅ 已实现
├── cmd/                 # 服务启动入口
├── configs/             # 配置文件
├── gmicro/              # 自定义微服务框架 ✅ 已实现
├── pkg/                 # 通用工具库 ✅ 已实现
├── third_party/         # 第三方依赖 ✅ 已实现
├── build/               # 构建配置（Docker等）
├── scripts/             # 脚本工具
└── tools/               # 工具代码
```

## 服务说明

### 后端微服务

| 服务 | 说明 | 状态 | 文档 |
|------|------|------|------|
| user | 用户服务：注册、登录、用户信息管理 | ⚠️ 需实现 | [查看](app/user/README.md) |
| goods | 商品服务：商品、分类、品牌、轮播图管理 | ⚠️ 需实现 | [查看](app/goods/README.md) |
| inventory | 库存服务：库存管理、扣减、归还 | ⚠️ 需实现 | [查看](app/inventory/README.md) |
| order | 订单服务：订单管理、购物车 | ⚠️ 需实现 | [查看](app/order/README.md) |

### API 网关

| 服务 | 说明 | 状态 | 文档 |
|------|------|------|------|
| api | C 端 API：用户、商品、订单接口 | ⚠️ 需实现 | [查看](app/mxshop/README.md) |
| admin | 管理后台 API：后台管理接口 | ⚠️ 需实现 | [查看](app/mxshop/README.md) |

## 技术栈

### 核心框架
- **Web 框架**: Gin
- **RPC 框架**: gRPC
- **ORM 框架**: GORM
- **配置管理**: Viper
- **日志**: Zap
- **依赖注入**: Wire

### 中间件
- **数据库**: MySQL
- **缓存**: Redis
- **搜索引擎**: Elasticsearch
- **服务注册**: Consul
- **配置中心**: Nacos
- **分布式事务**: DTM
- **链路追踪**: Jaeger/Zipkin
- **监控**: Prometheus + Grafana

### 其他
- **认证**: JWT
- **参数验证**: Validator
- **短信**: 阿里云 SMS

## 快速开始

### 1. 环境准备

确保已安装：
- Go 1.19+
- MySQL 5.7+
- Redis 5.0+
- Consul 1.12+
- Elasticsearch 7.x (可选，用于商品搜索)

### 2. 配置文件

每个服务的配置文件在 `configs/` 目录下：

```bash
# 复制配置模板
cp configs/shop/api.yaml.example configs/shop/api.yaml

# 编辑配置文件，填入真实的数据库、Redis、阿里云等配置
vim configs/shop/api.yaml
```

### 3. 启动服务

```bash
# 启动用户服务
go run cmd/user/user.go

# 启动商品服务
go run cmd/goods/goods.go

# 启动库存服务
go run cmd/inventory/inventory.go

# 启动订单服务
go run cmd/order/order.go

# 启动 API 网关
go run cmd/shop/api.go
```

## 学习路径

建议按以下顺序实现各个服务：

1. **用户服务 (User)** - 最简单，熟悉项目结构
2. **商品服务 (Goods)** - 学习 Elasticsearch 集成
3. **库存服务 (Inventory)** - 学习并发控制和分布式锁
4. **订单服务 (Order)** - 学习分布式事务
5. **API 网关** - 学习 HTTP 到 gRPC 的协议转换

## 项目特点

- ✅ **DDD 分层架构**: Controller -> Service -> Data
- ✅ **依赖注入**: 使用 Wire 进行依赖管理
- ✅ **服务发现**: 基于 Consul 的服务注册与发现
- ✅ **链路追踪**: OpenTelemetry 集成
- ✅ **监控告警**: Prometheus + Grafana
- ✅ **优雅退出**: 支持平滑关闭
- ✅ **健康检查**: 内置健康检查接口

## 注意事项

⚠️ **安全提醒**：
- 配置文件中的敏感信息（密钥、密码）已被 `.gitignore` 忽略
- 不要将真实的密钥提交到代码仓库
- 使用前请先配置好各项服务的连接信息

## 参考资料

- [gRPC 官方文档](https://grpc.io/docs/)
- [GORM 文档](https://gorm.io/zh_CN/)
- [Gin 文档](https://gin-gonic.com/zh-cn/docs/)
- [Consul 文档](https://www.consul.io/docs)
