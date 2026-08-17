# User 服务启动检查清单

## 当前状态对比

### ✅ 你已经有的文件（框架）
```
app/user/srv/
├── app.go              ✅ 应用入口
├── config/config.go    ✅ 配置定义
├── rpc.go              ✅ RPC 服务器配置
├── controller/user.go  ✅ Controller 框架（但缺少具体实现）
├── service/user.go     ✅ Service 框架（但缺少具体实现）
└── data/
    ├── user.go         ✅ Data 接口定义
    └── db/
        └── mysql.go    ✅ 数据库连接
```

### ❌ 你缺少的文件（业务实现）

#### 1. Controller 层 - 缺少具体的方法实现文件
```
❌ app/user/srv/controller/user/list.go        - 用户列表
❌ app/user/srv/controller/user/by_id.go       - 根据 ID 查询
❌ app/user/srv/controller/user/by_mobile.go   - 根据手机号查询
❌ app/user/srv/controller/user/create.go      - 创建用户
❌ app/user/srv/controller/user/update.go      - 更新用户
❌ app/user/srv/controller/user/password.go    - 密码相关
```

#### 2. Service 层 - 缺少具体实现
```
❌ app/user/srv/service/v1/service.go    - Service 接口定义
❌ app/user/srv/service/v1/user.go       - Service 实现
```

#### 3. Data 层 - 缺少具体实现
```
❌ app/user/srv/data/v1/db/user.go      - 数据访问实现
❌ app/user/srv/data/v1/user.go         - Data 接口和 DO 定义
```

#### 4. Wire 依赖注入
```
❌ app/user/srv/wire.go                 - Wire 定义文件
❌ app/user/srv/wire_gen.go             - Wire 生成的代码
```

#### 5. 启动入口
```
❌ cmd/user/user.go                     - main 函数
```

#### 6. 配置文件
```
❌ configs/user/srv.yaml                - 服务配置（数据库、端口等）
```

---

## 需要补充的完整内容

### 第一部分：Data 层（最底层，先实现）

#### 1.1 数据模型和接口 `app/user/srv/data/v1/user.go`

需要包含：
- `BaseModel` - 基础模型
- `UserDO` - 用户数据对象
- `UserStore` 接口 - 数据访问接口

#### 1.2 数据库实现 `app/user/srv/data/v1/db/user.go`

需要实现：
- `users` 结构体
- `NewUsers()` 构造函数
- `GetByID()` - 根据 ID 查询
- `GetByMobile()` - 根据手机号查询
- `List()` - 分页列表
- `Create()` - 创建用户
- `Update()` - 更新用户

#### 1.3 Wire ProviderSet `app/user/srv/data/v1/db/db.go`

需要添加：
```go
var ProviderSet = wire.NewSet(NewUsers, GetDBFactoryOr)
```

---

### 第二部分：Service 层（业务逻辑）

#### 2.1 Service 接口 `app/user/srv/service/v1/service.go`

需要定义：
```go
var ProviderSet = wire.NewSet(NewUserService)
```

#### 2.2 Service 实现 `app/user/srv/service/v1/user.go`

需要实现：
- `UserDTO` - 数据传输对象
- `UserSrv` 接口
- `userService` 实现
- `List()` - 列表查询
- `GetByID()` - ID 查询
- `GetByMobile()` - 手机号查询
- `Create()` - 创建用户（带业务校验）
- `Update()` - 更新用户

---

### 第三部分：Controller 层（gRPC 接口）

#### 3.1 Controller 主文件 `app/user/srv/controller/user/user.go`

需要：
- `userServer` 结构体
- `NewUserServer()` 构造函数
- `ProviderSet`

#### 3.2 各个方法实现

每个 gRPC 方法一个文件：
- `list.go` - GetUserList
- `by_id.go` - GetUserById
- `by_mobile.go` - GetUserByMobile
- `create.go` - CreateUser
- `update.go` - UpdateUser
- `password.go` - CheckPassWord

---

### 第四部分：Wire 依赖注入

#### 4.1 Wire 定义 `app/user/srv/wire.go`

```go
//go:build wireinject
// +build wireinject

package srv

import (
    "github.com/google/wire"
    "mxshop/app/pkg/options"
    "mxshop/app/user/srv/controller/user"
    "mxshop/app/user/srv/data/v1/db"
    v1 "mxshop/app/user/srv/service/v1"
    gapp "mxshop/gmicro/app"
    "mxshop/pkg/log"
)

func initApp(*options.NacosOptions, *log.Options, *options.ServerOptions, 
    *options.RegistryOptions, *options.TelemetryOptions, 
    *options.MySQLOptions) (*gapp.App, error) {
    wire.Build(
        ProviderSet,      // app.go 中的
        v1.ProviderSet,   // Service 层
        db.ProviderSet,   // Data 层
        user.ProviderSet, // Controller 层
    )
    return &gapp.App{}, nil
}
```

#### 4.2 生成 Wire 代码

运行命令：
```bash
cd app/user/srv
wire
```

会生成 `wire_gen.go`

---

### 第五部分：启动入口

#### 5.1 主函数 `cmd/user/user.go`

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
    rand.Seed(time.Now().UnixNano())
    if len(os.Getenv("GOMAXPROCS")) == 0 {
        runtime.GOMAXPROCS(runtime.NumCPU())
    }
    srv.NewApp("user-server").Run()
}
```

---

### 第六部分：配置文件

#### 6.1 服务配置 `configs/user/srv.yaml`

```yaml
log:
  name: mxshop-user-srv
  development: true
  level: debug
  format: console
  enable-color: true
  disable-caller: false
  disable-stacktrace: false
  output-paths: logs/mxshop-user-srv.log,stdout
  error-output-paths: logs/mxshop-user-srv.error.log

registry:
  address: 127.0.0.1:8500
  scheme: http

server:
  name: mxshop-user-srv
  host: 127.0.0.1
  port: 50051
  healthz: true
  enable-metrics: true
  profiling: true

mysql:
  host: 127.0.0.1
  port: 3306
  username: root
  password: your_password
  database: mxshop_user_srv
  max-idle-connections: 10
  max-open-connections: 100
  max-connection-lifetime: 10s

nacos:
  host: 127.0.0.1
  port: 8848
  namespace: ""
  group: "DEFAULT_GROUP"
  data-id: "user-sentinel"

telemetry:
  name: mxshop-user-srv
  endpoint: http://127.0.0.1:14268/api/traces
  sampler: 1.0
  batcher: jaeger
```

---

## 外部依赖检查

### ✅ 必须运行的服务

1. **MySQL 数据库**
   ```bash
   # 检查是否运行
   mysql -h127.0.0.1 -uroot -p -e "SHOW DATABASES;"
   
   # 需要创建数据库和表
   CREATE DATABASE mxshop_user_srv;
   USE mxshop_user_srv;
   
   CREATE TABLE `user` (
     `id` int(11) NOT NULL AUTO_INCREMENT,
     `mobile` varchar(11) NOT NULL,
     `password` varchar(100) NOT NULL,
     `nick_name` varchar(20) DEFAULT NULL,
     `birthday` datetime DEFAULT NULL,
     `gender` varchar(6) DEFAULT 'male',
     `role` int(11) DEFAULT 1,
     `add_time` datetime DEFAULT NULL,
     `update_time` datetime DEFAULT NULL,
     `deleted_at` datetime DEFAULT NULL,
     `is_deleted` tinyint(1) DEFAULT 0,
     PRIMARY KEY (`id`),
     UNIQUE KEY `idx_mobile` (`mobile`)
   ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
   ```

2. **Consul 服务注册中心**
   ```bash
   # 检查是否运行
   curl http://127.0.0.1:8500/v1/status/leader
   
   # 如果没运行，启动 Consul
   consul agent -dev
   ```

3. **Nacos 配置中心**（可选，如果启用限流）
   ```bash
   # 检查是否运行
   curl http://127.0.0.1:8848/nacos
   ```

4. **Jaeger 链路追踪**（可选）
   ```bash
   # Docker 启动
   docker run -d --name jaeger \
     -p 5775:5775/udp \
     -p 6831:6831/udp \
     -p 6832:6832/udp \
     -p 5778:5778 \
     -p 16686:16686 \
     -p 14268:14268 \
     -p 9411:9411 \
     jaegertracing/all-in-one:latest
   ```

---

## 启动步骤（完成所有代码后）

### 1. 安装依赖
```bash
cd /Users/luojilab/wanglele/projects/mxshop_refactor
go mod tidy
```

### 2. 生成 Wire 代码
```bash
cd app/user/srv
go install github.com/google/wire/cmd/wire@latest
wire
```

### 3. 启动外部服务
```bash
# 启动 MySQL（如果没运行）
# 启动 Consul
consul agent -dev
```

### 4. 创建数据库和表
```bash
mysql -h127.0.0.1 -uroot -p < scripts/user_db.sql
```

### 5. 修改配置文件
```bash
# 编辑 configs/user/srv.yaml
# 填入正确的数据库密码等信息
```

### 6. 启动 user 服务
```bash
go run cmd/user/user.go -c configs/user/srv.yaml
```

### 7. 测试服务
```bash
# 运行客户端测试
go run app/user/client/client.go
```

---

## 当前进度评估

### 完成度：约 20%

✅ **已完成**：
- 框架结构搭建
- 配置定义
- 基本的文件结构

❌ **待完成**：
- Data 层具体实现（数据库操作）
- Service 层具体实现（业务逻辑）
- Controller 层具体实现（gRPC 方法）
- Wire 依赖注入
- 启动入口
- 配置文件
- 数据库表结构

---

## 建议的实现顺序

### 第一阶段：Data 层（1-2 小时）
1. 完善 `data/v1/user.go` - 定义 DO 和接口
2. 实现 `data/v1/db/user.go` - 数据库操作
3. 添加 ProviderSet

### 第二阶段：Service 层（1 小时）
1. 定义 Service 接口
2. 实现业务逻辑
3. 添加 ProviderSet

### 第三阶段：Controller 层（1 小时）
1. 实现各个 gRPC 方法
2. 数据转换
3. 添加 ProviderSet

### 第四阶段：Wire 和配置（30 分钟）
1. 编写 wire.go
2. 生成 wire_gen.go
3. 编写配置文件

### 第五阶段：测试启动（30 分钟）
1. 创建数据库表
2. 启动 Consul
3. 启动服务
4. 测试调用

---

## 下一步建议

我可以帮你：

1. **逐层实现代码** - 从 Data 层开始，一步步完成
2. **生成完整文件** - 我可以直接为你生成所有缺失的文件
3. **调试启动问题** - 遇到错误时帮你排查

你想选择哪种方式？我建议：
- 如果想**快速看到效果**：我直接生成所有文件
- 如果想**深入理解**：我们一层层实现，边写边讲解

你的选择是？🚀
