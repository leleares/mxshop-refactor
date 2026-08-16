# User 服务完整代码解读 - 从数据库到 gRPC

## 目录
1. [启动流程总览](#1-启动流程总览)
2. [Data 层详解](#2-data-层详解)
3. [Service 层详解](#3-service-层详解)
4. [Controller 层详解](#4-controller-层详解)
5. [RPC 服务器详解](#5-rpc-服务器详解)
6. [完整请求流程示例](#6-完整请求流程示例)
7. [关键概念解释](#7-关键概念解释)

---

## 1. 启动流程总览

### 1.1 启动顺序

```
main() 
  ↓
srv.NewApp("user-server")
  ↓
initApp() [Wire 依赖注入]
  ↓
创建各层组件（从下到上）:
  1. 数据库连接 (db.GetDBFactoryOr)
  2. Data 层 (db.NewUsers)
  3. Service 层 (v1.NewUserService)
  4. Controller 层 (user.NewUserServer)
  5. RPC 服务器 (NewUserRPCServer)
  6. 微服务应用 (NewUserApp)
  ↓
userApp.Run()
  ↓
启动 gRPC 服务器，注册到 Consul
```

### 1.2 Wire 依赖注入详解

**wire_gen.go 中生成的完整流程**：

```go
func initApp(...) (*app.App, error) {
    // 步骤 1: 创建 Consul 服务注册器
    registrar := NewRegistrar(registryOptions)
    
    // 步骤 2: 创建数据库连接（单例模式）
    gormDB, err := db.GetDBFactoryOr(mySQLOptions)
    
    // 步骤 3: 创建 Data 层 - 数据访问对象
    userStore := db.NewUsers(gormDB)
    
    // 步骤 4: 创建 Service 层 - 注入 Data 层
    userSrv := v1.NewUserService(userStore)
    
    // 步骤 5: 创建 Controller 层 - 注入 Service 层
    userServer := user.NewUserServer(userSrv)
    
    // 步骤 6: 创建 Nacos 配置中心客户端
    nacosDataSource, err := NewNacosDataSource(nacosOptions)
    
    // 步骤 7: 创建 gRPC 服务器 - 注入 Controller
    server, err := NewUserRPCServer(
        telemetryOptions,
        serverOptions,
        userServer,
        nacosDataSource,
    )
    
    // 步骤 8: 创建微服务应用 - 整合所有组件
    appApp, err := NewUserApp(
        logOptions,
        registrar,
        serverOptions,
        server,
    )
    
    return appApp, nil
}
```

**依赖关系图**：

```
                    微服务应用 (App)
                         ↑
                         |
         ┌───────────────┼───────────────┐
         |               |               |
    服务注册器        RPC服务器         日志配置
    (Consul)             ↑
                         |
                    Controller层
                    (userServer)
                         ↑
                         |
                    Service层
                    (userService)
                         ↑
                         |
                     Data层
                    (userStore)
                         ↑
                         |
                    数据库连接
                    (gormDB)
```

---

## 2. Data 层详解

### 2.1 数据库连接管理 (mysql.go)

```go
var (
    dbFactory *gorm.DB    // 全局数据库连接（单例）
    once      sync.Once   // 确保只初始化一次
)

func GetDBFactoryOr(mysqlOpts *options.MySQLOptions) (*gorm.DB, error) {
    var err error
    // sync.Once 确保这段代码只执行一次
    once.Do(func() {
        // 构建 MySQL DSN 连接字符串
        dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
            mysqlOpts.Username,  // root
            mysqlOpts.Password,  // 密码
            mysqlOpts.Host,      // 127.0.0.1
            mysqlOpts.Port,      // 3306
            mysqlOpts.Database)  // mxshop_user_srv
        
        // 使用 GORM 连接数据库
        dbFactory, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
        
        // 获取底层的 *sql.DB
        sqlDB, _ := dbFactory.DB()
        
        // 设置连接池参数
        sqlDB.SetMaxOpenConns(mysqlOpts.MaxOpenConnections)  // 最大打开连接数
        sqlDB.SetMaxIdleConns(mysqlOpts.MaxIdleConnections)  // 最大空闲连接数
        sqlDB.SetConnMaxLifetime(mysqlOpts.MaxConnectionLifetime) // 连接最大生命周期
    })
    
    return dbFactory, nil
}
```

**为什么使用单例模式？**
- 数据库连接是昂贵的资源
- 避免重复创建连接
- 复用连接池，提高性能

### 2.2 数据模型 (user.go - Data/v1)

```go
// 基础模型 - 所有表都有的公共字段
type BaseModel struct {
    ID        int32          `gorm:"primarykey"`
    CreatedAt time.Time      `gorm:"column:add_time"`
    UpdatedAt time.Time      `gorm:"column:update_time"`
    DeletedAt gorm.DeletedAt // 软删除
    IsDeleted bool
}

// 用户数据对象 (DO - Data Object)
type UserDO struct {
    BaseModel
    Mobile   string     `gorm:"index:idx_mobile;unique;type:varchar(11);not null"`
    Password string     `gorm:"type:varchar(100);not null"`
    NickName string     `gorm:"type:varchar(20)"`
    Birthday *time.Time `gorm:"type:datetime"`
    Gender   string     `gorm:"column:gender;default:male;type:varchar(6)"`
    Role     int        `gorm:"column:role;default:1;type:int"`
}

// 指定表名
func (u *UserDO) TableName() string {
    return "user"
}
```

**GORM 标签说明**：
- `gorm:"primarykey"` - 主键
- `index:idx_mobile` - 创建索引
- `unique` - 唯一约束
- `type:varchar(11)` - 字段类型
- `not null` - 非空约束
- `default:male` - 默认值

### 2.3 数据访问接口 (user.go - Data/v1)

```go
// UserStore 定义了数据访问的接口（面向接口编程）
type UserStore interface {
    List(ctx context.Context, orderby []string, opts metav1.ListMeta) (*UserDOList, error)
    GetByMobile(ctx context.Context, mobile string) (*UserDO, error)
    GetByID(ctx context.Context, id uint64) (*UserDO, error)
    Create(ctx context.Context, user *UserDO) error
    Update(ctx context.Context, user *UserDO) error
}
```

**为什么定义接口？**
- **解耦**：上层只依赖接口，不依赖具体实现
- **可测试**：可以 mock 接口进行单元测试
- **可替换**：可以轻松切换不同的数据存储实现（MySQL → MongoDB）

### 2.4 数据访问实现 (user.go - Data/v1/db)

```go
// users 是 UserStore 接口的具体实现
type users struct {
    db *gorm.DB  // 持有数据库连接
}

// NewUsers 是构造函数（工厂方法）
func NewUsers(db *gorm.DB) dv1.UserStore {
    return &users{db: db}
}

// 确保 users 实现了 UserStore 接口（编译时检查）
var _ dv1.UserStore = &users{}
```

#### 2.4.1 根据手机号查询用户

```go
func (u *users) GetByMobile(ctx context.Context, mobile string) (*dv1.UserDO, error) {
    user := dv1.UserDO{}
    
    // GORM 查询：SELECT * FROM user WHERE mobile=? LIMIT 1
    err := u.db.Where("mobile=?", mobile).First(&user).Error
    
    if err != nil {
        // 区分错误类型
        if errors.Is(err, gorm.ErrRecordNotFound) {
            // 用户不存在 - 业务错误
            return nil, errors.WithCode(code.ErrUserNotFound, err.Error())
        }
        // 数据库错误 - 系统错误
        return nil, errors.WithCode(code2.ErrDatabase, err.Error())
    }
    
    return &user, nil
}
```

**关键点**：
1. 使用 `errors.WithCode()` 包装错误，添加错误码
2. 区分业务错误（用户不存在）和系统错误（数据库故障）
3. 不直接返回 GORM 的错误，而是包装成统一的错误格式

#### 2.4.2 创建用户

```go
func (u *users) Create(ctx context.Context, user *dv1.UserDO) error {
    // GORM 插入：INSERT INTO user (...) VALUES (...)
    tx := u.db.Create(user)
    
    if tx.Error != nil {
        return errors.WithCode(code2.ErrDatabase, tx.Error.Error())
    }
    
    // 注意：user.ID 会被 GORM 自动填充为插入后的自增 ID
    return nil
}
```

#### 2.4.3 用户列表查询（分页）

```go
func (u *users) List(ctx context.Context, orderby []string, opts metav1.ListMeta) (*dv1.UserDOList, error) {
    ret := &dv1.UserDOList{}
    
    // 1. 处理分页参数
    var limit, offset int
    if opts.PageSize == 0 {
        limit = 10  // 默认每页 10 条
    } else {
        limit = opts.PageSize
    }
    
    if opts.Page > 0 {
        offset = (opts.Page - 1) * limit  // 计算偏移量
    }
    
    // 2. 构建查询
    query := u.db
    
    // 3. 添加排序
    for _, value := range orderby {
        query = query.Order(value)  // ORDER BY xxx
    }
    
    // 4. 执行查询并统计总数
    // SELECT * FROM user ORDER BY xxx LIMIT ? OFFSET ?
    // SELECT COUNT(*) FROM user
    d := query.Offset(offset).Limit(limit).Find(&ret.Items).Count(&ret.TotalCount)
    
    if d.Error != nil {
        return nil, errors.WithCode(code2.ErrDatabase, d.Error.Error())
    }
    
    return ret, nil
}
```

**分页逻辑**：
- Page = 1, PageSize = 10 → offset = 0, limit = 10 (第 1-10 条)
- Page = 2, PageSize = 10 → offset = 10, limit = 10 (第 11-20 条)

---

## 3. Service 层详解

### 3.1 Service 层的作用

**Service 层是业务逻辑层**，负责：
1. **业务规则校验**：如用户是否已存在
2. **事务管理**：跨多个数据操作的事务
3. **数据转换**：DO (Data Object) → DTO (Data Transfer Object)
4. **调用 Data 层**：但不直接操作数据库

### 3.2 DTO 数据传输对象

```go
// UserDTO 继承了 UserDO，但可以添加额外的业务字段
type UserDTO struct {
    dv1.UserDO
    // 可以添加业务相关的字段，如：
    // Token string
    // Permissions []string
}
```

**DO vs DTO**：
- **DO (Data Object)**：数据库表的映射，包含 GORM 标签
- **DTO (Data Transfer Object)**：业务层的数据对象，可能包含计算字段

### 3.3 Service 接口定义

```go
type UserSrv interface {
    List(ctx context.Context, orderby []string, opts metav1.ListMeta) (*UserDTOList, error)
    Create(ctx context.Context, user *UserDTO) error
    Update(ctx context.Context, user *UserDTO) error
    GetByID(ctx context.Context, ID uint64) (*UserDTO, error)
    GetByMobile(ctx context.Context, mobile string) (*UserDTO, error)
}
```

### 3.4 Service 实现

```go
type userService struct {
    userStrore dv1.UserStore  // 依赖 Data 层接口
}

// 构造函数 - Wire 会调用这个函数
func NewUserService(us dv1.UserStore) UserSrv {
    return &userService{
        userStrore: us,
    }
}
```

#### 3.4.1 创建用户 - 业务逻辑

```go
func (u *userService) Create(ctx context.Context, user *UserDTO) error {
    // ===== 业务逻辑：先检查用户是否存在 =====
    _, err := u.userStrore.GetByMobile(ctx, user.Mobile)
    
    // 如果查询出错，且错误是"用户不存在"
    if err != nil && errors.IsCode(err, code.ErrUserNotFound) {
        // 用户不存在，可以创建
        return u.userStrore.Create(ctx, &user.UserDO)
    }
    
    // 如果没有错误，说明用户已存在
    if err == nil {
        return errors.WithCode(code.ErrUserAlreadyExists, "用户已经存在")
    }
    
    // 其他错误（如数据库错误）直接返回
    return err
}
```

**业务逻辑流程**：
```
1. 查询手机号是否存在
   ├─ 用户不存在 → 继续创建 ✅
   ├─ 用户已存在 → 返回错误 ❌
   └─ 数据库错误 → 返回错误 ❌
```

#### 3.4.2 更新用户

```go
func (u *userService) Update(ctx context.Context, user *UserDTO) error {
    // 先查询用户是否存在
    _, err := u.userStrore.GetByID(ctx, uint64(user.ID))
    if err != nil {
        return err  // 用户不存在或数据库错误
    }
    
    // 用户存在，执行更新
    return u.userStrore.Update(ctx, &user.UserDO)
}
```

#### 3.4.3 查询用户列表

```go
func (u *userService) List(ctx context.Context, orderby []string, opts metav1.ListMeta) (*UserDTOList, error) {
    // 1. 调用 Data 层获取数据
    doList, err := u.userStrore.List(ctx, orderby, opts)
    if err != nil {
        return nil, err
    }
    
    // 2. 将 DO 转换为 DTO
    var userDTOList UserDTOList
    userDTOList.TotalCount = doList.TotalCount
    
    for _, value := range doList.Items {
        userDTO := UserDTO{*value}
        userDTOList.Items = append(userDTOList.Items, &userDTO)
    }
    
    // 3. 可以在这里添加额外的业务逻辑
    // 例如：过滤某些敏感字段、添加计算字段等
    
    return &userDTOList, nil
}
```

---

## 4. Controller 层详解

### 4.1 Controller 层的职责

Controller 层（也叫 Handler 层）是 **gRPC 请求的入口**，负责：
1. **接收 gRPC 请求**
2. **参数验证**：检查请求参数是否合法
3. **调用 Service 层**：执行业务逻辑
4. **数据转换**：DTO → Protobuf Message
5. **错误处理**：将错误转换为 gRPC 错误码
6. **日志记录**：记录请求日志

### 4.2 Controller 结构体

```go
type userServer struct {
    v1.UnimplementedUserServer  // gRPC 要求嵌入
    srv srv1.UserSrv            // 依赖 Service 层接口
}

// 构造函数 - Wire 会调用
func NewUserServer(srv srv1.UserSrv) v1.UserServer {
    return &userServer{srv: srv}
}

// 确保实现了 gRPC 接口
var _ v1.UserServer = &userServer{}
```

### 4.3 获取用户列表

```go
func (us *userServer) GetUserList(ctx context.Context, info *upbv1.PageInfo) (*upbv1.UserListResponse, error) {
    // 1. 记录日志
    log.Info("GetUserList is called")
    
    // 2. 转换 Protobuf 请求 → Service 层参数
    srvOpts := metav1.ListMeta{
        Page:     int(info.Pn),      // 页码
        PageSize: int(info.PSize),   // 每页大小
    }
    
    // 3. 调用 Service 层
    dtoList, err := us.srv.List(ctx, []string{}, srvOpts)
    if err != nil {
        return nil, err
    }
    
    // 4. 转换 DTO → Protobuf 响应
    var rsp upbv1.UserListResponse
    for _, value := range dtoList.Items {
        userRsp := DTOToResponse(*value)  // DTO → Protobuf
        rsp.Data = append(rsp.Data, userRsp)
    }
    
    return &rsp, nil
}
```

**数据流转**：
```
Protobuf Request (upbv1.PageInfo)
   ↓ 转换
Service 参数 (metav1.ListMeta)
   ↓ 调用 Service
Service 响应 (UserDTOList)
   ↓ 转换
Protobuf Response (upbv1.UserListResponse)
```

### 4.4 创建用户

```go
func (u *userServer) CreateUser(ctx context.Context, request *upbv1.CreateUserInfo) (*upbv1.UserInfoResponse, error) {
    // 1. 日志
    log.Infof("create user function called.")
    
    // 2. Protobuf → DO
    userDO := v12.UserDO{
        Mobile:   request.Mobile,
        NickName: request.NickName,
        Password: request.PassWord,
    }
    
    // 3. DO → DTO
    userDTO := v1.UserDTO{userDO}
    
    // 4. 调用 Service 层
    err := u.srv.Create(ctx, &userDTO)
    if err != nil {
        log.Errorf("create user: %v, error: %v", userDTO, err)
        return nil, err
    }
    
    // 5. DTO → Protobuf 响应
    userInfoRsp := DTOToResponse(userDTO)
    return userInfoRsp, nil
}
```

### 4.5 数据转换辅助函数

```go
func DTOToResponse(userDTO srvv1.UserDTO) *upbv1.UserInfoResponse {
    userInfoRsp := upbv1.UserInfoResponse{
        Id:       userDTO.ID,
        PassWord: userDTO.Password,
        NickName: userDTO.NickName,
        Gender:   userDTO.Gender,
        Role:     int32(userDTO.Role),
        Mobile:   userDTO.Mobile,
    }
    
    // 处理可能为 nil 的字段
    if userDTO.Birthday != nil {
        userInfoRsp.BirthDay = uint64(userDTO.Birthday.Unix())
    }
    
    return &userInfoRsp
}
```

---

## 5. RPC 服务器详解

### 5.1 RPC 服务器创建 (rpc.go)

```go
func NewUserRPCServer(
    telemetry *options.TelemetryOptions,    // 链路追踪配置
    serverOpts *options.ServerOptions,      // 服务器配置
    userver upb.UserServer,                 // Controller 层实例
    dataNacos *nacos.NacosDataSource,       // Nacos 配置中心
) (*rpcserver.Server, error) {
    
    // 1. 初始化链路追踪（Jaeger/Zipkin）
    trace.InitAgent(trace.Options{
        telemetry.Name,      // 服务名
        telemetry.Endpoint,  // 收集器地址
        telemetry.Sampler,   // 采样率
        telemetry.Batcher,   // 批次大小
    })
    
    // 2. 构建服务器地址
    rpcAddr := fmt.Sprintf("%s:%d", serverOpts.Host, serverOpts.Port)
    // 例如："127.0.0.1:50051"
    
    // 3. 配置服务器选项
    var opts []rpcserver.ServerOption
    opts = append(opts, rpcserver.WithAddress(rpcAddr))
    
    // 4. 如果启用限流，添加 Sentinel 拦截器
    if serverOpts.EnableLimit {
        opts = append(opts, 
            rpcserver.WithUnaryInterceptor(grpc.NewUnaryServerInterceptor()))
        
        // 初始化 Nacos 数据源（读取限流规则）
        err := dataNacos.Initialize()
        if err != nil {
            return nil, err
        }
    }
    
    // 5. 创建 gRPC 服务器
    urpcServer := rpcserver.NewServer(opts...)
    
    // 6. 注册用户服务
    upb.RegisterUserServer(urpcServer.Server, userver)
    
    return urpcServer, nil
}
```

**关键组件**：
1. **链路追踪**：记录请求调用链
2. **限流拦截器**：使用 Sentinel 进行流量控制
3. **服务注册**：将 Controller 注册到 gRPC 服务器

### 5.2 Nacos 配置中心

```go
func NewNacosDataSource(opts *options.NacosOptions) (*nacos.NacosDataSource, error) {
    // 1. Nacos 服务器配置
    sc := []constant.ServerConfig{
        {
            ContextPath: "/nacos",
            Port:        opts.Port,    // 8848
            IpAddr:      opts.Host,    // 127.0.0.1
        },
    }
    
    // 2. Nacos 客户端配置
    cc := constant.ClientConfig{
        NamespaceId: opts.Namespace,
        TimeoutMs:   5000,
    }
    
    // 3. 创建配置客户端
    client, err := clients.CreateConfigClient(map[string]interface{}{
        "serverConfigs": sc,
        "clientConfig":  cc,
    })
    
    // 4. 注册流控规则处理器
    h := datasource.NewFlowRulesHandler(datasource.FlowRuleJsonArrayParser)
    
    // 5. 创建 Nacos 数据源
    nds, err := nacos.NewNacosDataSource(client, opts.Group, opts.DataId, h)
    
    return nds, nil
}
```

**Nacos 的作用**：
- 配置中心：集中管理配置
- 限流规则：动态更新限流规则，无需重启服务

---

## 6. 完整请求流程示例

### 6.1 场景：获取用户列表

**客户端调用**：
```go
client.GetUserList(ctx, &pb.PageInfo{
    Pn: 1,      // 第 1 页
    PSize: 10,  // 每页 10 条
})
```

**服务端处理流程**：

```
1. gRPC 服务器接收请求
   ↓
2. Controller 层 (userServer.GetUserList)
   ├─ 记录日志
   ├─ 转换参数：Protobuf → Service 层参数
   └─ 调用 Service 层
   
3. Service 层 (userService.List)
   ├─ 调用 Data 层
   └─ 转换数据：DO → DTO
   
4. Data 层 (users.List)
   ├─ 构建 SQL 查询
   ├─ 执行 GORM 查询
   └─ 返回 UserDOList
   
5. 数据回流
   Data → Service → Controller
   
6. Controller 转换响应
   DTO → Protobuf Response
   
7. gRPC 返回响应给客户端
```

**SQL 执行**：
```sql
SELECT * FROM user ORDER BY id LIMIT 10 OFFSET 0;
SELECT COUNT(*) FROM user;
```

### 6.2 场景：创建用户

**请求数据**：
```go
client.CreateUser(ctx, &pb.CreateUserInfo{
    Mobile:   "13800138000",
    NickName: "张三",
    PassWord: "123456",
})
```

**处理流程**：

```
1. Controller 层 (userServer.CreateUser)
   ├─ Protobuf → UserDO
   ├─ UserDO → UserDTO
   └─ 调用 Service.Create
   
2. Service 层 (userService.Create)
   ├─ 检查手机号是否存在
   │  └─ 调用 Data.GetByMobile
   ├─ 如果不存在，调用 Data.Create
   └─ 如果已存在，返回错误
   
3. Data 层 (users.Create)
   └─ 执行 GORM 插入
   
4. 响应返回
   ├─ user.ID 被填充（自增ID）
   ├─ DTO → Protobuf Response
   └─ 返回给客户端
```

**SQL 执行**：
```sql
-- 步骤 1: 检查用户是否存在
SELECT * FROM user WHERE mobile='13800138000' LIMIT 1;

-- 步骤 2: 如果不存在，创建用户
INSERT INTO user (mobile, nick_name, password, add_time, update_time) 
VALUES ('13800138000', '张三', '123456', NOW(), NOW());
```

---

## 7. 关键概念解释

### 7.1 DO、DTO、VO 的区别

| 类型 | 全称 | 所在层 | 作用 |
|------|------|--------|------|
| **DO** | Data Object | Data 层 | 数据库表的映射，包含 GORM 标签 |
| **DTO** | Data Transfer Object | Service 层 | 业务层数据传输，可能包含计算字段 |
| **VO** | Value Object | Controller 层 | Protobuf Message，用于 RPC 通信 |

**转换流程**：
```
数据库 ←→ DO ←→ DTO ←→ VO (Protobuf) ←→ 客户端
       Data层   Service层   Controller层
```

### 7.2 为什么要分层？

**单一职责原则**：
- **Data 层**：只负责数据访问，不关心业务逻辑
- **Service 层**：只负责业务逻辑，不关心数据如何存储
- **Controller 层**：只负责协议转换，不关心业务如何实现

**好处**：
1. **代码清晰**：每层职责明确
2. **易于测试**：可以 mock 接口进行单元测试
3. **易于维护**：修改某一层不影响其他层
4. **易于扩展**：如替换数据库、添加缓存等

### 7.3 依赖注入（DI）与控制反转（IoC）

**传统方式**（依赖具体实现）：
```go
type UserService struct {
    userStore *MySQLUserStore  // 直接依赖具体实现
}

func NewUserService() *UserService {
    return &UserService{
        userStore: &MySQLUserStore{},  // 硬编码
    }
}
```

**依赖注入方式**（依赖接口）：
```go
type UserService struct {
    userStore UserStore  // 依赖接口
}

func NewUserService(us UserStore) *UserService {
    return &UserService{
        userStore: us,  // 从外部注入
    }
}
```

**Wire 自动注入**：
```go
// Wire 自动分析依赖关系，按正确顺序创建对象
userStore := db.NewUsers(gormDB)
userService := v1.NewUserService(userStore)
userController := user.NewUserServer(userService)
```

### 7.4 接口的重要性

**面向接口编程的好处**：

1. **解耦**：上层不依赖下层的具体实现
2. **可测试**：可以 mock 接口
3. **可替换**：轻松切换实现

**示例**：
```go
// Service 依赖 Data 层接口
type UserService struct {
    userStore UserStore  // 接口
}

// 可以轻松切换实现
// 实现 1: MySQL
type MySQLUserStore struct {}

// 实现 2: MongoDB
type MongoUserStore struct {}

// 实现 3: Mock（用于测试）
type MockUserStore struct {}
```

### 7.5 错误处理最佳实践

**分层错误处理**：

```go
// Data 层：包装数据库错误
if err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return errors.WithCode(code.ErrUserNotFound, err.Error())
    }
    return errors.WithCode(code.ErrDatabase, err.Error())
}

// Service 层：添加业务错误
if userExists {
    return errors.WithCode(code.ErrUserAlreadyExists, "用户已存在")
}

// Controller 层：转换为 gRPC 错误
if err != nil {
    log.Errorf("create user failed: %v", err)
    return nil, status.Error(codes.Internal, err.Error())
}
```

**错误码的作用**：
- 客户端可以根据错误码做不同的处理
- 便于监控和告警
- 便于日志分析

---

## 总结

### User 服务的核心设计

1. **三层架构**：Data → Service → Controller
2. **依赖注入**：使用 Wire 自动管理依赖
3. **面向接口**：各层之间依赖接口，不依赖具体实现
4. **单例模式**：数据库连接使用单例
5. **错误处理**：统一的错误码和错误包装
6. **分页查询**：使用 Offset/Limit 实现
7. **业务校验**：在 Service 层进行业务规则校验

### 启动到运行的完整链路

```
1. main() 启动
2. 读取配置（配置文件 + 环境变量）
3. Wire 依赖注入（创建所有组件）
4. 连接 MySQL
5. 创建 gRPC 服务器
6. 注册到 Consul
7. 启动健康检查
8. 监听端口，等待请求
9. 接收请求 → Controller → Service → Data → 数据库
10. 响应返回 → 客户端
```

---

## 下一步学习建议

1. **动手实践**：按照这个文档，在你的项目中一步步实现 user 服务
2. **查看配置文件**：理解 `configs/user/srv.yaml` 的配置项
3. **运行服务**：启动 MySQL、Consul，然后运行 user 服务
4. **编写客户端**：创建一个 gRPC 客户端调用 user 服务
5. **查看日志**：理解服务运行时的日志输出
6. **学习其他服务**：goods、order、inventory 服务的结构类似

准备好开始实践了吗？

---

## 附录：gRPC 客户端如何精准命中服务端方法

### 核心机制

**答案**：通过 **Protobuf 定义** + **代码生成** + **方法路径字符串** 实现精准匹配。

### 完整流程

#### 1. Protobuf 定义 (user.proto)

```protobuf
service User {
    rpc GetUserList(PageInfo) returns (UserListResponse);
}
```

#### 2. protoc 生成客户端代码

**客户端生成的代码**：

```go
// 客户端接口
type UserClient interface {
    GetUserList(ctx context.Context, in *PageInfo, opts ...grpc.CallOption) (*UserListResponse, error)
}

// 客户端实现
type userClient struct {
    cc grpc.ClientConnInterface  // 持有连接
}

// 关键：客户端调用方法
func (c *userClient) GetUserList(ctx context.Context, in *PageInfo, opts ...grpc.CallOption) (*UserListResponse, error) {
    out := new(UserListResponse)
    // ⭐ 注意这个字符串："/User/GetUserList"
    // 格式："/服务名/方法名"
    err := c.cc.Invoke(ctx, "/User/GetUserList", in, out, opts...)
    if err != nil {
        return nil, err
    }
    return out, nil
}
```

**关键点**：`"/User/GetUserList"` - 这是 gRPC 的方法路径

#### 3. protoc 生成服务端代码

**服务端接口**：

```go
// 服务端接口（你的 Controller 需要实现这个接口）
type UserServer interface {
    GetUserList(context.Context, *PageInfo) (*UserListResponse, error)
    GetUserByMobile(context.Context, *MobileRequest) (*UserInfoResponse, error)
    // ... 其他方法
}
```

**服务端注册方法**：

```go
// 注册服务到 gRPC 服务器
func RegisterUserServer(s grpc.ServiceRegistrar, srv UserServer) {
    s.RegisterService(&User_ServiceDesc, srv)
}

// 服务描述（关键！）
var User_ServiceDesc = grpc.ServiceDesc{
    ServiceName: "User",  // 服务名
    HandlerType: (*UserServer)(nil),
    Methods: []grpc.MethodDesc{
        {
            MethodName: "GetUserList",  // ⭐ 方法名
            Handler:    _User_GetUserList_Handler,  // ⭐ 处理函数
        },
        {
            MethodName: "GetUserByMobile",
            Handler:    _User_GetUserByMobile_Handler,
        },
        // ... 其他方法
    },
}

// 实际的处理函数（自动生成的）
func _User_GetUserList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
    in := new(PageInfo)
    if err := dec(in); err != nil {
        return nil, err
    }
    if interceptor == nil {
        // ⭐ 直接调用你实现的 GetUserList 方法
        return srv.(UserServer).GetUserList(ctx, in)
    }
    info := &grpc.UnaryServerInfo{
        Server:     srv,
        FullMethod: "/User/GetUserList",  // ⭐ 完整方法路径
    }
    handler := func(ctx context.Context, req interface{}) (interface{}, error) {
        return srv.(UserServer).GetUserList(ctx, req.(*PageInfo))
    }
    return interceptor(ctx, in, info, handler)
}
```

#### 4. 你的 Controller 实现接口

```go
// 你的 Controller 结构体
type userServer struct {
    v1.UnimplementedUserServer  // 嵌入未实现的方法
    srv srv1.UserSrv             // 依赖 Service 层
}

// 实现 UserServer 接口的 GetUserList 方法
func (us *userServer) GetUserList(ctx context.Context, info *upbv1.PageInfo) (*upbv1.UserListResponse, error) {
    log.Info("GetUserList is called")
    // ... 你的业务逻辑
    return &rsp, nil
}
```

#### 5. 注册到 gRPC 服务器 (rpc.go)

```go
func NewUserRPCServer(..., userver upb.UserServer, ...) (*rpcserver.Server, error) {
    urpcServer := rpcserver.NewServer(opts...)
    
    // ⭐ 关键：注册服务
    // 此时 gRPC 服务器内部建立了路径映射：
    // "/User/GetUserList" -> _User_GetUserList_Handler -> userver.GetUserList
    upb.RegisterUserServer(urpcServer.Server, userver)
    
    return urpcServer, nil
}
```

### 完整调用链路

```
客户端调用:
uc.GetUserList(ctx, &v1.PageInfo{})
    ↓
客户端生成代码:
c.cc.Invoke(ctx, "/User/GetUserList", in, out)
    ↓
gRPC 框架:
通过网络发送请求，携带方法路径 "/User/GetUserList"
    ↓
服务端 gRPC 框架:
接收请求，解析方法路径 "/User/GetUserList"
    ↓
查找路由表:
在 User_ServiceDesc 中找到 "GetUserList" 对应的 Handler
    ↓
调用 Handler:
_User_GetUserList_Handler(srv, ctx, ...)
    ↓
调用你的实现:
srv.(UserServer).GetUserList(ctx, in)
    ↓
你的 Controller:
userServer.GetUserList(ctx, info)
```

### 路由映射表（内部结构）

gRPC 服务器内部维护了一个映射表：

```go
map[string]Handler{
    "/User/GetUserList":      _User_GetUserList_Handler,
    "/User/GetUserByMobile":  _User_GetUserByMobile_Handler,
    "/User/GetUserById":      _User_GetUserById_Handler,
    "/User/CreateUser":       _User_CreateUser_Handler,
    "/User/UpdateUser":       _User_UpdateUser_Handler,
    "/User/CheckPassWord":    _User_CheckPassWord_Handler,
}
```

### 关键要点

1. **方法路径是唯一标识**：`"/服务名/方法名"`
2. **编译时确定**：不是运行时反射，而是编译时生成的代码
3. **类型安全**：客户端和服务端都有明确的类型定义
4. **接口约束**：服务端必须实现 `UserServer` 接口

### 为什么需要 Protobuf？

**Protobuf 是契约**：
- 客户端和服务端都根据同一个 `.proto` 文件生成代码
- 保证了方法签名的一致性
- 保证了数据序列化的兼容性

**如果没有 Protobuf**：
- 客户端不知道服务端有哪些方法
- 不知道参数类型是什么
- 不知道如何序列化数据

### 实际例子

**客户端代码**：
```go
// 1. 创建客户端
uc := v1.NewUserClient(conn)

// 2. 调用方法（看起来像本地调用）
resp, err := uc.GetUserList(context.Background(), &v1.PageInfo{
    Pn: 1,
    PSize: 10,
})
```

**实际发生的事情**：
1. `GetUserList` 方法内部调用 `c.cc.Invoke(ctx, "/User/GetUserList", in, out)`
2. 请求通过网络发送到服务端
3. 请求头包含方法路径：`"/User/GetUserList"`
4. 服务端 gRPC 框架根据路径找到对应的 Handler
5. Handler 调用你的 `userServer.GetUserList` 方法
6. 响应原路返回

### 总结

**精准匹配的秘密**：
1. **Protobuf 定义了契约**：服务名 + 方法名
2. **代码生成器生成了桥接代码**：客户端和服务端
3. **方法路径字符串是唯一标识**：`"/User/GetUserList"`
4. **gRPC 框架维护路由表**：路径 → Handler → 你的实现

这就是为什么客户端调用 `GetUserList` 能够精准命中服务端的 `GetUserList` 方法！🎯
