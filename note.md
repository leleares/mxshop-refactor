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
5. consumer可以有多个，他们可以进行分组，叫做consumer组
6. 一个consumer可以消费多个partition，但同一个consumer组中的多个consumer不可以同时消费同一个partition
7. 假如现在有3个broker(分别为b1、b2、b3)，3个topic中有3个partition（分别为p1、p2、p3），则会产生3个master partition，三个master partition会分布在不同的broker中，这样当向每个partition中写入数据时都会将数据同步到其他的slave Partition，这样以来向每一个不同的Partition中写入数据时写入的都是不同的broker

### 云原生
何为云原生？
云生的概念是相对于部署项目时部署到自己购买的服务器上而言的，假设现在是2010年，如果要部署项目，那就必须要先购买实体服务器，然后将项目开完完毕后再上传到实体服务器上部署，那个时候大部分都是单体应用，项目部署后还非常担心服务器会随时挂掉，带来很多不便，对于有流量的场景会再加机器来解决，这个时候会造成很多成本的浪费，对于云原生来说，其实就是购买云厂商的服务，可以根据流量实时的弹性伸缩所需要的云服务，云原生状态下基本是微服务，完整的原原生还要结合docker技术和K8S技术来讲。

#### docker
采用go语言开发，go的兴起，某种意义上是因为docker的流行。
原理：Docker 使用 Google 公司推出的 Go 语言 进行开发实现，基于 Linux 内核的 cgroup,namespace，以及
OverlayFS 类的 Union FS 等技术，对进程进行封装隔离，属于操作系统层面的虚拟化技术。由于隔离的进程独
立于宿主和其它的隔离的进程，因此也称其为容器。

#### 容器化技术
注意，容器化技术并非docker独有，docker只是容器化技术其中一种
拿平常在电脑上装的linux系统举例（ubentu），ubentu是独立于电脑的操作系统，我们完全可以在这个上面部署服务，这其实就是容器化技术的体现，但是一旦项目多了起来，比如有多个java项目，多个nodejs项目，甚至java和nodejs又有多个版本，这就导致虚拟机操作系统非常重而且版本维护起来比较复杂。虚拟机也是相当占用电脑内存，启动慢。

因此，想到能不能有一种技术能够将具体语言以及它的操作系统和依赖环境完全打包成一个隔离的环境（容器），这就是docker的核心思想。
docer性能比较好的原因就是上面说的它是基于Linux内核的一些比较新的技术，导致它可以直接使用宿主的相关能力，并非像上面提到的ubentu那样直接是一个隔离的操作系统，docker多个容器之间也可以进行通讯，并且宿主也可以合理安排各个docker容器使用内存情况。

#### 为什么够远是云原生最佳语言？
对于使用docker技术来说，如果是部署Java项目，假如说有三个项目，那么对应每一个docker容器中不仅要装操作系统，还要装jvm编译器，这是很耗时的，对于go语言来说，他的交付产物只有一个EXE文件，它在docker容器中是很轻量的，只依赖操作系统，别的环境都不需要。

#### docker 核心概念
1. 镜像：类似于光盘，镜像中印有相关数据(或叫root文件系统，体积一般较大)（例如mysql的镜像肯定印有mysql的版本号以及相关配置），镜像是一个静态的概念，正如光盘不能被重新录入数据一样，镜像不包含任何动态数据，其内容在构建之后也不会被改变。
2. 容器：镜像运行起来就是容器，容器可以被：创建、启动，停止、删除、暂停等。容器的本质是进程，但它与直接在宿主运行的普通进程有所不同，容器进程有自己独立的命名空间，并且拥有自己独立的root文件系统，自己的网络配置。类似于一个独立的操作系统一样。镜像和容器都是分层存储，每个容器运行时是以镜像为基础层在其上创建一个当前容器的存储层，我们称这个为容器运行时读写而准备的存储层为容器存储层。但要注意这个容器存储层属于容器内的概念，一旦整个容器被销毁，其内部产生的数据也将一并被销毁，这个时候就需要将容器中的某些数据挂载到宿主的目录（数据卷），这个时候即使容器被销毁，那么数据也保存在宿主中。
3. 仓库：docker hub，里面有n多镜像，类似于node生态的npm
```
docker 基础操作：
镜像相关操作👇：
拉取镜像 `docker pull [options] [address:port]/仓库名:标签`
运行镜像 `docker run -it --rm ubuntu:18.04 bash` -it 是两个参数，i是interacive的意思指的是进入交互操作，t指的是terminal说明想进入bash终端交互；-rm 说明在容器退出后将容器删除
列出本地的镜像：`docker image ls`
删除镜像：`docker image rm id `
容器相关操作👇：
可以理解为容器就是一个微型操作系统。
运行一个新的容器：`docker run imageid` -d参数表示以守护进程方式来运行 -p参数指定端口映射，hostport:containerport，例如 -p 8081:8080意思是宿主机上的8081端口对应着容器的8080端口。服务间调用访问8081端口即可。也可以指定-P命令，大P的意思是随机在宿主机上找一个可用的端口做映射即可。
运行一个已有的容器：`docker start containerid`
重启一个已有的容器：`docker restart containerid` 内部会先进行stop随后再进行start
列出所有容器：`docker ps -a` -a的作用是列出所有的容器，包括已经停止的
停止：`docker stop containerid`
移除：`docker rm containerid`
容器在运行中怎么强制删除：`docker rm -f containerid`
进入容器：`docker exec -it containerid /bin/bash` 就是会直接进入容器的bash终端
进入容器查看日志：`docker logs containerid`

```