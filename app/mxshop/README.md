# MXShop API Gateway - API 网关服务

## 服务说明

API 网关是面向前端的统一入口，负责接收 HTTP 请求并转发到后端的各个微服务。

## 包含的服务

### 1. API 服务 (app/mxshop/api)
面向 C 端用户的 RESTful API，提供：
- 用户注册、登录、信息管理
- 商品浏览、搜索
- 购物车管理
- 订单创建和查询
- 短信验证码

### 2. Admin 服务 (app/mxshop/admin)
面向管理后台的 API，提供：
- 用户管理
- 商品管理
- 订单管理
- 其他后台管理功能

## 需要实现的功能

### API 服务功能模块

#### 用户模块 (user)
- 用户注册
- 用户登录（JWT 认证）
- 获取用户信息
- 更新用户信息
- 图形验证码
- 短信验证码

#### 商品模块 (goods)
- 商品列表（分页、搜索、筛选）
- 商品详情
- 商品分类
- 品牌列表
- 轮播图

#### 订单模块 (order)
- 创建订单
- 订单列表
- 订单详情
- 购物车管理

#### 短信模块 (sms)
- 发送短信验证码

### Admin 服务功能模块
- 用户列表管理
- 其他管理功能

## 目录结构

```
app/mxshop/
├── api/                 # C 端 API 服务
│   ├── config/          # 配置定义（保留）
│   ├── app.go           # 服务启动入口（保留）
│   ├── http.go          # HTTP 服务器（保留）
│   ├── router.go        # 路由注册（保留）
│   ├── auth.go          # JWT 认证（保留）
│   └── internal/        # 内部实现（需实现）
│       ├── controller/  # 控制器层
│       │   ├── user/    # 用户相关接口
│       │   ├── goods/   # 商品相关接口
│       │   └── sms/     # 短信相关接口
│       ├── service/     # 业务逻辑层
│       ├── data/        # RPC 客户端封装
│       └── domain/      # 请求/响应模型
├── admin/               # 管理后台 API 服务
│   ├── config/          # 配置定义（保留）
│   ├── app.go           # 服务启动入口（保留）
│   ├── http.go          # HTTP 服务器（保留）
│   ├── router.go        # 路由注册（保留）
│   └── controller/      # 控制器实现（需实现）
└── README.md            # 本文件
```

## 架构说明

API 网关的职责：
- **协议转换**: HTTP -> gRPC
- **认证鉴权**: JWT Token 验证
- **参数校验**: 请求参数验证
- **聚合调用**: 调用多个后端服务
- **响应转换**: 统一的响应格式

## 使用的技术栈

- Gin - Web 框架
- JWT - 用户认证
- Validator - 参数验证
- gRPC Client - 调用后端服务
- Redis - Session/缓存
- Consul - 服务发现
