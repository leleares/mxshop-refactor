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
查看容器端口映射情况：`docker port containerid xx` 查看容器的xx端口映射到主机的哪个端口上了
查看容器完整信息：`docker inspect containerid`

docker 挂载数据盘：
为什么docker容器要挂载外部数据？因为容器一旦被销毁，里面的状态全丢，将容器中的数据挂载到外部盘上不用担心容器销毁后数据丢失问题。
挂载数据盘 `docker run -v 本地绝对文件路径:容器中绝对文件路径` 可以是文件夹也可以是文件，注意：挂载目录的时候要注意：外部目录会完全覆盖容器中的目录
`docker run imageid -e param1=1 --name container1` -e是指定环境变量的意思，说白了就是全局参数，容器中可通过$param1获取，--name是给这个容器起名字
```

### 自制镜像 
k8s中的调度单位是pod，每一个pod里面可以认为是有容器，因此我们的项目必须要做成镜像，才能运行起来当作容器，才可以使用k8s进行管理。
#### Dockerfile 
Dockerfile  是一个用来构建镜像的文本文件，文本内容包含了一条条构建镜像所需的指令和说明。
Dockerfile 👇：
```
# 基于ngnix镜像构建镜像，向当前容器中的/usr/share/nginx/html/index.html目录中写入内容。
FROM nginx
RUN echo '这是一个本地构建的nginx镜像' > /usr/share/nginx/html/index.html
COPY⁄  home/srv/index.html /mydir/src/index.html
CMD ["/usr/myserver"]
EXPOSE 8080
```
Dockerfile 就是这个名字也没有后缀，一般放到项目的根目录，Dockerfile写完后如何生成image？
使用 `docker build` 可以加参数，常见参数如下：
- -t 指明要构建的镜像的名字
- -f Dockerfile 在哪个目录，若在当前路径则无需指定
#### Dockerfile 中的常见命令
#####  RUN
RUN 命令后面就相当于在shell终端执行的命令，RUN命令在build阶段运行，即构建镜像的阶段就会运行，与其类似的命令还有一个CMD命令，这个命令在image run的时候会执行
常见用法：RUN ["可执行文件","参数k","参数值v"] 等价于：./可执行文件 -k v

##### COPY
将本机文件copy到镜像文件中
常见用法：COPY home/srv /mydir/src 将本机中home/srv文件拷贝至镜像/mydir/src目录中，常见做法是将go项目build产物可执行文件copy到镜像中，这样镜像就可以将go项目run起来。

##### EXPOSE
告知docker需要将什么端口暴露出来，例如我的go项目端口号是8888，如果不使用这个命令，docker容器启动时可能开启的并不是8888这个端口而是80端口，因此需要使用EXPOSE命令告知docker我的项目端口是啥。这样他就会将8888端口映射到宿主机上的任意一个可用的端口。

##### ENTRYPOINT 和 CMD
CMD 命令设置容器启动后默认执行的命令及其参数，当Dockerfile中存在多个CMD命令，只有最后一个会被
执行，但CMD设置的命令能够被docker run后面的命令行参数覆盖替换
ENTRYPOINT配置容器启动时的执行命令，当运行 docker run时指定了其他命令，docker run时指定的命令
会追加到ENTRYPOINT配置命令行的参数一起执行。ENTRYPOINT 中的参数始终会被使用，与CMD不同，它不
会被替换。

##### ENV 和 ARG
```
FROM apline
ENV project="shop"
ENV name="rose"
```
定义构建镜像中的环境变量或者说常量，这个环境变量会被持久化到容器当中，当然可以在启动容器时修改这个环境变量，docker run -e name="jack"

```
FROM apline
ARG Version 1
```
arg参数就相当于变量，在构建镜像时可以由外部传入变量的值，例如docker build --build-args Version=2 这样来修改arg变量的值

##### WORKDIR
指定工作目录，用 workdir 指定的工作目录，会在构建镜像的每一层中都存在，比如我们想要构建一个基础镜像，以后各层的当前
目录就被改为workdir的目录 如用目录不存在 workdir会帮你创建目录

# k8s （kubernetes）
为什么需要k8s。
容器和服务不是一一对应的，一台服务器上可能运行多个容器，例如现在要对用户服务进行扩容，那么就启动一个用户服务的容器即可，但是放到哪台服务器上？这个不应该由人来手工维护，而应该有一个中间层叫做：容器编排。
容器编排本身记录了所有服务器上运行的实例，清楚这些服务的负载情况，他就可以将新的容器分配给合适的服务。k8s就是这个编排容器的实现，但k8s能力远强于此。
例如服务监控，k8s会监控每一个容器，当容器挂掉的时候会再重启一个容器实例。
k8s 管理多个服务器，例如服务器A、B、C。每个服务器中有若干个 pod（每个pod中含有一个或者多个docker实例），k8s核心能力就是调度这些pod在合适的服务器上运行，每个pod可以理解为是一个服务实例。

### kubesphere
账号：admin
密码：Kubesphere123
k8s是核心的容器编排服务的实现，但其未提供可视化操作界面，kubesphere就是基于k8s服务实现的可视化软件。
核心概念：
1. 工作负载（deployment）
主要工作是编写yml，yml可以理解为是指导k8s的核心文件，包括指定容器启动参数，以及部署。
部署分为三种类型：
- 无状态服务：普通后端服务，每一次启动一个实例本质上无差别。
- 有状态服务：本身有数据，例如sql、redis、nacos、es这些
- 守护服务：本身无需暴露端口，没人会调用他，例如负责上报日志到阿里云的后台服务
2. 服务（service）
k8s中负责服务调度的概念，例如多个容器组可以暴露为同一端口和name，本身也可以完成负载均衡等工作。
3. 容器组（pod）
一个pod中可以有一个或者多个container，之所以这么设计，主要是因为很多场景下一个container旁边需要有一个sidecar container，如果k8s的调度单位是container，那么复杂度会升高。
### RocketMQ 消息链路与关键名词

#### 消息发送和消费链路

假设 A 服务是 Producer，B 服务是 Consumer，双方都在进程中引入 RocketMQ Go SDK，并配置 NameServer 地址。

```text
A 服务（Producer）
  1. Go SDK 向 NameServer 查询 Topic `test` 的路由
  2. NameServer 返回该 Topic 所在的 Broker 和 MessageQueue
  3. Go SDK 选择一个 MessageQueue，直接把消息发送给 Broker
  4. Broker 持久化消息，并向 A 返回发送结果

B 服务（Consumer）
  1. Go SDK 向 NameServer 查询 Topic `test` 的路由
  2. Go SDK 连接 Broker，通过长轮询拉取消息
  3. SDK 调用 B 注册的消费函数
  4. 处理成功则确认消费；处理失败则稍后重试
```

NameServer 只负责服务发现和路由查询，不转发业务消息。真正的数据链路是：

```text
A（Producer）──发送消息──> Broker ──拉取消息──> B（Consumer）
       │                                      │
       └──── 双方通过 NameServer 查询路由 ────┘
```

通常 A、B 只需配置 NameServer，查询路由、选择 Broker、建立连接和失败重试均由 Go SDK 完成。

#### 关键名词

- **Producer**：消息生产者，例如 A 服务。
- **Consumer**：消息消费者，例如 B 服务。
- **NameServer**：保存 Broker 地址和 Topic 路由；不保存、不转发业务消息。
- **Broker**：接收、持久化和提供消息的核心服务。
- **Topic**：消息的逻辑分类，例如 `order-created`；建议由运维提前创建。
- **MessageQueue**：Topic 内部的队列分片，用于并行收发；同一 Queue 内可保证顺序，不同 Queue 之间不保证全局顺序。
- **Producer Group**：具有相同发送职责的一组 Producer。
- **Consumer Group**：具有相同消费逻辑的一组 Consumer。同组多个实例共同分摊消息；不同 Group 各自消费一遍。
- **Tag**：Topic 内的二级分类，例如 `CREATE`、`CANCEL`。
- **Key**：消息的业务唯一标识，例如订单号，便于查询和排障。
- **Offset**：Consumer Group 在 MessageQueue 中的消费位置。
- **重试消息**：消费失败后由 RocketMQ 再次投递的消息。
- **死信消息（DLQ）**：超过最大重试次数仍失败的消息，需要人工排查或补偿。
- **客户端 SDK**：编译进 A、B 服务的 RocketMQ Go 依赖，负责路由查询、网络连接、消息收发、负载均衡和重试，不是独立部署的服务。

#### 重要语义

- A 收到发送成功，只表示 Broker 已接收消息，不表示 B 已处理成功。
- B 的消费逻辑应保证幂等，避免消息重复投递导致重复扣款、重复减库存等问题。
- 同一个 Consumer Group 内，一条消息通常只由一个实例处理；不同 Consumer Group 会分别收到消息。
- RocketMQ 的 PushConsumer 名为“Push”，底层实际通过长轮询从 Broker 拉取消息。

#### 常用端口

- `9876/TCP`：NameServer 路由查询和 Broker 注册。
- `10911/TCP`：Broker 与 Producer、Consumer 的主要通信端口。
- `10909/TCP`：Broker VIP Channel 端口，不是 HTTP；关闭 VIP Channel 后可只使用 `10911`。
- `10912/TCP`：Broker Master/Slave 数据复制端口；当前单 Broker 开发环境无需对业务客户端暴露。

#### Docker Desktop、Docker CLI、Docker Engine、containerd、runc、 CRI
  ## 一次 docker run 背后发生了什么
  假设执行：
  docker run -d --name web -p 8080:80 nginx
  大致调用链是：

  你
  │
  │ docker run / docker ps
  ▼
  Docker CLI
  │
  │ Docker API（Unix Socket 或网络）
  ▼
  Docker Engine / dockerd
  │
  ├── 管理镜像、容器、网络、Volume
  ├── 必要时从镜像仓库拉取镜像
  └── 调用 containerd
          │
          └── 调用低层运行时 runc
                  │
                  └── Linux 内核
                      ├── namespaces：进程、网络等隔离
                      ├── cgroups：CPU、内存等限制
                      └── capabilities / seccomp：权限控制

容器虽说是一个独立的环境，但本质上仍是服务器上的一个或者多个进程，对于docker容器来说，其原理就是基于linux os的 namespace、cgroup等底层能力实现进程、网络等隔离；CPU、内存等限制。
搞清概念：
所谓docker desktop，其实主要是供windows和mac使用的，windows和mac安装docker desktop后启动这个软件本质上就是启动了一个linux虚拟机，这才是运行docker容器的基石。注意docker desktop是docker的全家桶，内部包含了：Docker CLI、Docker Engine、containerd等。

docker cli 和 docker engine：
执行：docker pull nginx、docker ps等命令其实就是docker cli，但cli只是命令行工具，本质是交给Docker Engine去做事，因此docker engine是负责：理镜像、容器、网络、Volume、拉取镜像、调用containerd的。

containerd概念，对比docker engine做一些理解：
  Docker Engine = 面向用户的完整平台
  containerd = 更底层、更专注的运行时组件

  Docker Engine 在 containerd 之上增加了很多开发者友好的功能：

  • Docker API
  • docker 命令体验
  • 容器网络管理
  • Volume 管理
  • 镜像构建
  • Docker Compose
  • 登录镜像仓库
  • 端口映射等高级功能

  containerd 更聚焦于：

  • 镜像拉取、存储
  • 容器创建和生命周期
  • 容器快照
  • 调用 runc
  • 为 Kubernetes 等上层系统提供接口

k8s无需docker engine，因为k8s本身就存在很多对于容器的编排逻辑，k8s定义了CRI（container runtime interface），任何实现了这个接口的都可以被k8s当作容器进行调用，docker的containerd只是其中一种而已只不过比较主流，因此对于k8s管理的容器，无法使用docker ps 等命令查看容器，因此没有docker cli、docker engine这些上层内容，只有底层的containerd，只能通过k8s的 sudo crictl ps、sudo crictl pods、sudo crictl images 命令来查看容器或者pod。


#### 工作负载和服务的关系
首先对于一个服务，完整的部署流程是这样：
将服务打成镜像 -> 上传到自建镜像仓库（如harbor） -> 在k8s中创建工作负载，并指定此镜像 -> 工作负载状态变为运行中 -> 创建service指定上面的工作负载即可对内网或者公网（nodeport）暴露服务

底层：image -> container -> pod -> service
工作负载，其实就是管理pod的k8s对象，工作负载运行中即代表创建好了pod，可以对工作负载进行声明式配置指定想要几个pod。
工作负载虽然可以创建pod，但由于pod可能会被销毁重建，因此外部服务如果依赖pod作为服务的endpoint不可靠，因此加了一层service，service负责维护这些pod，调用service即可。

#### ingress
属于k8s中的概念，相当于网关或者nginx服务器的角色，外部不可能直接访问service的ip地址，例如现在有user service、good service，就可以设置一层ingress，ingress的域名设置为mxshop.com，外部直接访问mxshop.com即可，根据路由前缀匹配具体的服务，例如mxshop.com/user则打到user service上，如mxshop.com/good则打到good service上

### k8s中的持久卷
pod随时可以被销毁，数据如果存储在pod中，那么数据也会被销毁，因此pod中的数据通常都会挂载到外部的持久卷上，这样pod重启时，由于数据在持久卷上，状态不会丢失。
持久卷主要可以分为两种类型：本地磁盘和网盘，本地磁盘很容易理解，就是从所在机器上分配的存储空间中申领的。网盘就是存储到网络磁盘上的。
PV和PVC：PV（persist volume）、PVC（persist volume claim）
PV指的是静态的磁盘大小。PVC指的是申请多少磁盘，是一个动词的概念。

### k8s中的配置
配置文件，本质上就是在k8s中创建配置文件然后挂载到pod中所依赖的配置文件路径中。
配置文件分为一般配置文件和保密配置文件。

#### go项目的构建和部署
演化进程见下：
1. 最简单粗暴的部署，但性能差，本质就是把go项目复制到镜像中，然后在执行构建镜像时执行go build来构建出go的可执行文件，镜像run起来后项目就跑起来了，这种方式，发现镜像极大（依赖go环境），性能自然也就差。
```sh
FROM golang:1.19.4-alpine3.17 # 基于go环境的镜像
WORKDIR /build
COPY go.mod .
copy main.go .
RUN go build -o hello main.go
CMD ［"./hello"］
```

2. 如何对上一步进行瘦身？那就是在宿主机上将go项目build好，然后直接将可执行文件复制到镜像中，那么此镜像就不依赖go的环境，alpine 只是linux上运行镜像必要的文件系统极小。但这样做有个问题就是如果有不同版本的go项目，那么宿主机上go版本的管理就是个问题。
```sh
FROM alpine # 必要的镜像环境

COPY hello /hello

CMD ["/hello"]
```

3. 能否有一种办法，既能隔离go环境，又能使得构建出的镜像不依赖go环境，这就是docker的多阶段构建。第一阶段只负责“加工”，最终发布的镜像不会包含第一阶段的 Go 环境。
```sh
FROM golang:1.19.4-alpine3.17 AS builder # 给这个镜像起个alias，方便后面的阶段引用
WORKDIR /build
COPY go.mod .
COPY main.go .
RUN go build -o hello main.go .
# 可以认为，从第二个FROM开始已经是第二个阶段了，第二个阶段基于第一个阶段的构建结果来构建
FROM alpine
WORKDIR /build
COPY --from=builder /build/hello .
CMD ["./hello"］
```
### devops、ci、cd、gitops
#### devops（dev:development； ops:operations）
比较宏观的理论，典型的代表就是jenkins。

#### ci（continous integration）
对于go项目，一般的流程就是: 代码上传 -> （代码检测、扫描等，一般公司没这个环节）-> go build -> build image -> image push

使用 jenkins 构建镜像
1. 从 gitlab 拉取代码 -> 构建打包 -> build image -> push 到镜像仓库
   注意打镜像时要加上版本号 

jenkins 通过插件拉取到分支代码后，会读取项目中Jenkins文件，这个文件就是指导jenkins工作的核心文件，Jenkins文件在项目中被维护，可以用来指导如何进行打包部署，这就是gitops思想（使用git来管理配置文件，配置文件指导一系列cicd过程）。
项目中应写Dockerfile用于指导如何进行将本项目打包成镜像（对于go项目就是采用多阶段构建将可执行文件打包到镜像当中）
写Jenkinsfile，本质上是写pipeline脚本，分为多阶段，例如阶段一：执行sh登录到镜像仓库harbor 阶段二：打包构建镜像 阶段三：推送镜像到harbor
部署：然后去k8s上操作，修改工作负载指定的镜像地址，修改完成后，pod会逐步替换为新的容器内容，这时候就重新部署成功了。

#### cd （continuous deployment）
流程：拉取image -> 使用k8s进行部署

