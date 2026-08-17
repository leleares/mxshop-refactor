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

### 