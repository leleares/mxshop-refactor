# User 服务完整解读 - 从启动到运行

## 概述

这份文档会带你从 `cmd/user/user.go` 启动入口开始，一步步理解 user 服务的完整运行流程。

---

## 1. 启动入口：cmd/user/user.go

```go
package main

import (
	"math/rand"
	"mxshop/app/user/srv"
	"os"
	"runtime"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())                    // 初始化随机数种子
	if len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())            // 设置 CPU 核心数
	}
	srv.NewApp("user-server").Run()                     // 核心：创建并运行应用
}
```

**关键点**：
- `srv.NewApp("user-server")` - 创建应用实例
- `.Run()` - 启动应用

---

## 2. 应用创建：app/user/srv/app.go

### 2.1 NewApp 函数

```go
func NewApp(basename string) *app.App {
	cfg := config.New()                          // 1. 创建配置对象
	appl := app.NewApp("user",                   // 2. 创建应用
		"mxshop",
		app.WithOptions(cfg),                    // 3. 注入配置
		app.WithRunFunc(run(cfg)),               // 4. 注入运行函数
	)
	return appl
}
```

**执行流程**：
1. `config.New()` - 创建默认配置对象
2. `app.NewApp()` - 创建应用框架（来自 pkg/app）
3. `WithOptions(cfg)` - 将配置注入应用
4. `WithRunFunc(run(cfg))` - 注入真正的业务运行函数

### 2.2 run 函数 - 核心业务启动

```go
func run(cfg *config.Config) app.RunFunc {
	return func(baseName string) error {
		// 关键：通过 Wire 依赖注入初始化所有组件
		userApp, err := initApp(
			cfg.Nacos,        // Nacos 配置中心
			cfg.Log,          // 日志配置
			cfg.Server,       // 服务器配置
			cfg.Registry,     // 服务注册配置（Consul）
			cfg.Telemetry,    // 链路追踪配置
			cfg.MySQLOptions, // 数据库配置
		)
		if err != nil {
			return err
		}

		// 启动应用
		if err := userApp.Run(); err != nil {
			log.Errorf("run user app error: %s", err)
			return err
		}
		return nil
	}
}
```

**关键**：`initApp()` 是 Wire 自动生成的依赖注入函数

---

## 3. 配置管理：app/user/srv/config/config.go

### 3.1 Config 结构体

```go
type Config struct {
	Log          *log.Options              // 日志配置
	Nacos        *options.NacosOptions     // Nacos 配置中心
	Server       *options.ServerOptions    // 服务器配置（端口、名称等）
	Registry     *options.RegistryOptions  // 服务注册配置（Consul）
	Telemetry    *options.TelemetryOptions // 链路追踪配置
	MySQLOptions *options.MySQLOptions     // MySQL 数据库配置
}
```

### 3.2 配置初始化

```go
func New() *Config {
	return &Config{
		Log:          log.NewOptions(),          // 默认日志配置
		Server:       options.NewServerOptions(), // 默认服务器配置
		Registry:     options.NewRegistryOptions(), // 默认注册配置
		Telemetry:    options.NewTelemetryOptions(), // 默认链路追踪
		MySQLOptions: options.NewMySQLOptions(),     // 默认 MySQL 配置
		Nacos:        options.NewNacosOptions(),     // 默认 Nacos 配置
	}
}
```

**配置来源**：
1. 代码中的默认值
2. 配置文件（configs/user/srv.yaml）
3. 环境变量
4. 命令行参数

---

## 4. 依赖注入：Wire 工作原理

### 4.1 Wire 定义文件：wire.go

```go
//go:build wireinject
// +build wireinject

package srv

func initApp(
	*options.NacosOptions,
	*log.Options,
	*options.ServerOptions,
	*options.RegistryOptions,
	*options.TelemetryOptions,
	*options.MySQLOptions,
) (*gapp.App, error) {
	// Wire 会扫描这些 ProviderSet，自动生成依赖注入代码
	wire.Build(
		ProviderSet,      // app.go 中定义的提供者
		v1.ProviderSet,   // service 层的提供者
		db.ProviderSet,   // data 层的提供者
		user.ProviderSet, // controller 层的提供者
	)
	return &gapp.App{}, nil
}
```

### 4.2 Wire 生成的代码：wire_gen.go

```go
func initApp(...各种配置...) (*app.App, error) {
	// 1. 创建服务注册器（Consul）
	registrar := NewRegistrar(registryOptions)
	
	// 2. 创建数据库连接
	gormDB, err := db.GetDBFactoryOr(mySQLOptions)
	if err != nil {
		return nil, err
	}
	
	// 3. 创建数据访问层（Data Layer）
	userStore := db.NewUsers(gormDB)
	
	// 4. 创建业务逻辑层（Service Layer）
	userSrv := v1.NewUserService(userStore)
	
	// 5. 创建控制器层（Controller Layer）
	userServer := user.NewUserServer(userSrv)
	
	// 6. 创建 Nacos 数据源
	nacosDataSource, err := NewNacosDataSource(nacosOptions)
	
	// 7. 创建 RPC 服务器
	server, err := NewUserRPCServer(
		telemetryOptions,
		serverOptions,
		userServer,
		nacosDataSource,
	)
	
	// 8. 创建应用实例
	appApp, err := NewUserApp(
		logOptions,
		registrar,
		serverOptions,
		server,
	)
	
	return appApp, nil
}
```

**Wire 的作用**：
- 自动解决依赖关系
- 按正确的顺序创建对象
- 传递依赖给构造函数
- 编译时检查，避免运行时错误

---

## 5. 三层架构详解

### 5.1 架构图

```
┌─────────────────────────────────────────────┐
│         gRPC Client (其他服务调用)            │
└─────────────────┬───────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────┐
│  Controller 层 (app/user/srv/controller)    │
│  - 接收 gRPC 请求                            │
│  - 参数验证                                  │
│  - 调用 Service 层                           │
│  - 返回响应                                  │
└─────────────────┬───────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────┐
│  Service 层 (app/user/srv/service)          │
│  - 核心业务逻辑                              │
│  - 事务管理                                  │
│  - 业务规则验证                              │
│  - 调用 Data 层                              │
└─────────────────┬───────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────┐
│  Data 层 (app/user/srv/data)                │
│  - 数据库操作（GORM）                        │
│  - 数据模型转换                              │
│  - 缓存操作                                  │
└─────────────────┬───────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────┐
│           MySQL Database                     │
└─────────────────────────────────────────────┘
```

### 5.2 为什么要分层？

**单一职责原则**：
- Controller：只负责接收请求和返回响应
- Service：只负责业务逻辑
- Data：只负责数据访问

**好处**：
- 代码清晰易维护
- 便于单元测试
- 可以独立修改某一层而不影响其他层

---

## 6. 接下来要查看的代码

为了完整理解 user 服务，我们需要按顺序查看：

### 第一步：Data 层（数据访问）
1. `app/user/srv/data/v1/db/mysql.go` - 数据库连接
2. `app/user/srv/data/v1/db/user.go` - 用户数据模型
3. `app/user/srv/data/v1/user.go` - 用户数据访问接口

### 第二步：Service 层（业务逻辑）
4. `app/user/srv/service/v1/service.go` - 服务接口定义
5. `app/user/srv/service/v1/user.go` - 用户业务逻辑实现

### 第三步：Controller 层（请求处理）
6. `app/user/srv/controller/user/user.go` - gRPC 控制器

### 第四步：RPC 服务器
7. `app/user/srv/rpc.go` - gRPC 服务器配置

---

## 7. 完整启动流程总结

```
main()
  └─> srv.NewApp("user-server")
       ├─> config.New()                    // 创建配置
       └─> app.NewApp()                     // 创建应用框架
            └─> WithRunFunc(run(cfg))       // 注入运行函数
  
  └─> app.Run()
       └─> run(cfg)()
            └─> initApp(...)                // Wire 依赖注入
                 ├─> NewRegistrar()         // 创建 Consul 注册器
                 ├─> db.GetDBFactoryOr()    // 连接数据库
                 ├─> db.NewUsers()          // 创建 Data 层
                 ├─> v1.NewUserService()    // 创建 Service 层
                 ├─> user.NewUserServer()   // 创建 Controller 层
                 ├─> NewUserRPCServer()     // 创建 gRPC 服务器
                 └─> NewUserApp()           // 创建微服务应用
            
            └─> userApp.Run()               // 启动服务
                 ├─> 启动 gRPC 服务器
                 ├─> 注册到 Consul
                 ├─> 启动健康检查
                 └─> 等待停止信号
```

---

## 下一步

准备好了吗？我们现在开始深入查看每一层的具体实现代码！

我会从最底层的 Data 层开始，带你一步步理解每个函数的作用。
