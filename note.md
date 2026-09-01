### 三层代码结构
指的是：data -> service -> controller 三层
data 层服务负责与数据库进行交互
service 层服务调用data层接口，并对来自controller层的请求做业务处理
controller 层负责接收请求，并将数据转换成最终的resp

data层对于interface的相关使用：
在data层，会有一个interface来约定此data层对外应提供哪些接口能力（例如就叫UserStore），与此相对应会有一个struct实现这些接口（例如就叫User）
```go
interface UserStore {
    GetUserList
    GetUserByID
}

// 该 struct 绑定的方法需要 DB 连接，则其应该添加 DB 属性，项目启动时统一装配
type User struct {
    db *gorm.DB
}

func (u *User)GetUserList(){
    u.db.xxxx
}

func (u *User)GetUserByID(){
    u.db.xxxx
}

// 至此 User 这个 struct 实现了 UserStore 这个 interface
// 由于需要供service层调用，因此需要暴露构造方法，注意这个方法返回的是一个 interface
func NewUser(db *gorm.DB) *UserStore{
    return &User{
        db
    }
}

```

来到service层，同样会有 interface 相关使用，应有一个interface约定此service对外提供的能力（例如就叫UserSrv），与此相对应会有一个struct实现这些接口（例如就叫UserService）
```go
interface UserSrv {
    GetUserListSrv
    GetUserByIDSrv 
}

// 该 struct 绑定的方法需要 data 层提供数据服务，上面的 store 已经实现了这些能力，因此此struct中注入，此struct绑定的方法可以随意使用
type UserService struct {
    userStore *UserStore
}

func (u *UserService)GetUserListSrv(){
    u.userStore.GetUserList and u.userStore.GetUserByID ...
}

func (u *UserService)GetUserByIDSrv(){
     u.userStore.GetUserByID
}

// 至此 UserService 这个 struct 实现了 UserSrv 这个 interface
// 由于需要供controller层调用，因此需要暴露构造方法，注意这个方法返回的是一个 interface
func NewUserService(userStore *UserStore) *UserSrv{
    return &UserService{
        userStore
    }
}
```
来到controller层，由于我们是grpc服务，protobuf文件中定义了interface，因此我们无需再定义interface，只需定义struct，往struct中绑定方法使其实现protobuf中interface即可，因此绑定struct的这些方法吧不可以像data层和contoller层那样随意起名字了，必须严格按照protobuf中interface中的函数名来。
```go
// protobuf:
type UserServer interface {
	GetUserList(context.Context, *PageInfo) (*UserListResponse, error)
	GetUserById(context.Context, *IdRequest) (*UserInfoResponse, error)
}

// 因此我们需要在controller层定义一个struct，假如就叫 UserServerStruct
type UserServerStruct struct {
    // 需要service层提供的能力
    userSrv UserSrv
}

func (u *UserServerStruct)GetUserList(){
    u.userSrv.GetUserListSrv xxx
}

func (u *UserServerStruct)GetUserById(){
    u.userSrv.GetUserByIDSrv xxx
}
```

现在有个最大的问题就是这些struct都是何时被初始化的？例如 db 连接是怎么来的，service 层需要的store是怎么被注入的， controller层需要的srv是怎么被注入的，这就是项目初始化时【装配】做的事。
```go
// 装配：
	gormDB := db.GetDBFactoryOr(mySQLOptions)
    // 装配db
	userStore := db.NewUsers(gormDB)
    // 装配store
	userSrv := v1.NewUserService(userStore)
    // 装配srv
	userServer := user.NewUserServer(userSrv)
```
### gmicro
为什么项目中有两层 server，既有 grpc server 还有 gmicro 这一层 server。
原因在于 grpc 仅仅为一种服务间的调用方式， gmicro 是微服务框架，负责 grpc 服务的启动、服务注册、健康检查、优雅退出等一系列工作，换句话说，微服务里究竟是 grpc 服务还是http服务这都是不重要的，这都是被微服务框架所约束和进行统一管理的。

### 创建商品入数据库和es如何保证一致性
由于查询商品列表时需要使用到es进行过滤查询，因此对于商品crud都需要入数据库和es。
对于创建商品时的入库和入es，怎么保证入库成功并且也入es？
1. 使用事务：入库前开启事务，入库操作和入es操作都放到事务当中，任一失败都将导致rollback。最后进行commit。有问题：假如入es成功了，但是在接受es入库成功信号时受网络抖动影响竟然是超时，这时会导致rollback操作，但事实就是入库操作取消了，但是es已经写入了。
2. 使用canal（阿里开源组件），本质上是想实现基于可靠消息实现一致性，基本原理：sql的操作会有binlog的产生（数据库原生的机制），当有binlog的产生时，binlog就会通知canal，canal就负责往消息队列中写入消息（注意这时候是有topic的）（例如kafka），那我们的业务逻辑只需要写消费者即可，消费这个topic做任何我们想做的事，例如入es，甚至可以写多个消费者来进行其他的业务处理。

### 非线程安全
slice和map都是线程不安全的
也就是说当并发向slice中append元素时，可能会导致元素丢失。map当并发向map中写入数据时会导致kv丢失。