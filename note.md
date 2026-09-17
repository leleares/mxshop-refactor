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

### SQL中的 DDL 和 DML
SQL 分两大类：DDL和DML
 DDL（定义结构  Data Definition Language）：例如：CREATE TABLE goods (...)、ALTER TABLE goods ADD COLUMN ...、 DROP TABLE goods、 TRUNCATE TABLE goods。说白了就是重构model的
 DML（操作数据 Data Manipulation Language）：—INSERT / SELECT / UPDATE / DELETE

 DML 的前提是表已经用 DDL 建好了


 ### saga 分布式事务
 核心思想：分为执行和补偿，拿一个下单操作来举例子，假如要跨服务调用扣减库存、扣减优惠券、扣减余额等服务，那就分别需要实现：扣减库存、归还库存；扣减优惠券、归还优惠券；扣减余额、添加余额，也就是说，每一步执行都有对应的补偿动作，这样一来整个跨微服务的调用过程，但凡有一个环节出现问题都会调用上一步微服务的补偿动作，这种解决分布式事务的方案叫做saga分布式解决方案。请注意saga只是一个理论，DTM是真正落地的开源框架。DTM在这里就是充当调度器的作用，我们可以去关心每个动作的执行和补偿接口，然后都交给DTM管理器去接管。

 dtm 支持http和grpc。
 对于http，需要借助dtm服务器的能力（本地可起，也可部署），将相关正向操作和补偿操作交给dtm服务器去完成就行。
 小结一句话：HTTP 模式 = 起一个 dtm 服务器 + 业务暴露「正向/补偿」接口 + 发起方用 dtmcli.NewSaga().Add(action, compensate, data).Submit() 提交，dtm 负责顺序执行、失败逆序补偿、断点续跑；业务侧用 barrier 保证幂等/空补偿/悬挂。
 基于grpc跟http比无非就是换了一种调用方式，本质原理相同。
 ```go
      // 通过saga dtm服务器来处理分布式事务，需要把需要进行操作的action（正向操作）和compensation（补偿操作）交给saga，saga服务会去调用具体的业务接口（调用库存服务、调用购物车服务..）
      // 开启一个gid
      saga.Add("inventory/busi.Bus/TransIn",     // 每个Add操作都是一个新的brand_id  
         "inventory/busi.Bus/TransInRevert", req).
      Add("cart/busi.Bus/TransIn", //新的brand_id  
          "cart/busi.Bus/TransInRevert", req)
 ```

saga 中的子事务屏障：所谓子事务屏障，指的是dtm在实际执行时会遇到很多问题，主要是由于网络是不可靠的：
1. 幂等：dtm调用业务接口，业务接口resp了OK，但由于网络抖动，resp丢失了，dtm于是进行重试，从而造成多进行扣减库存。
2. 空补偿：DTM调用库存服务，但在调用的过程当中，由于网络抖动这个请求丢掉了。于是DTM认为调用服务超时执行整体回滚操作这个时候库存服务会执行补偿动作，从而添加额外的库存。
3. 悬挂：类似的网络问题导致的业务错乱

子事务屏障的解决思路是：在全局完整事务的基础之上，业务方能否在本地来维护一个本地事务，同时业务方通过一张表来进行记录，当dtm调用本身的时候，通过查询表数据可以知道是否发生了以上三种情况从而避免。
表结构主要字段：gid、brand_id、op：gid 表示全局事务，一个saga操作许多业务的一串操作就是同一个gid，brand_id就是每一个分支，op：action、compensation。

### ioc 框架 wire
本质是运行脚本完成固化的初始化流程。
拿三层代码结构来举例子，data层service层和controller层对于这三层来说，data层需要向上暴露NewData方法，service层依赖NewData方法实例化后传入自己的NewService层，controller层依赖NewService方法实例化后传入NewController层，再交给上层，这样的过程将当繁琐wire就是帮助你完成这个事情，你可以将这函数所需的参数全部交给init的函数，然后再将所有的构造函数都交给wire脚本，执行wire命名就拿帮你初始化这些固定的流程。

### kafka 
分布式流消息处理系统，流消息决定了它可以像消息队列一样进行sub或者pub，分布式提供高并发支持与容错性。

1. broker：一个kafka服务器就是一个broker，多个kafka服务器组成broker集群。
2. topic：消息的标识，producer和consumer一般基于topic来进行pub sub
3. partition：一个topic会有多个partition，可以理解为topic中的分区，partition是队列结构
4. offset：真正的消息内容是存储于partition队列中的，在队列中的位置就是offset偏移量
5. 一个consumer可以消费多个partition，但多个consumer不可以消费同一个partition
6. 假如现在有3个broker(分别为b1、b2、b3)，3个topic中有3个partition（分别为p1、p2、p3），则会产生3个master partition，三个master partition会分布在不同的broker中，这样当向每个partition中写入数据时都会将数据同步到其他的slave Partition，这样以来向每一个不同的Partition中写入数据时写入的都是不同的broker