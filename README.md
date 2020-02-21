# Esim文档

## 架构

> Esim的架构来源于《实现领域驱动设计》的六边形架构和阿里的COLA架构，这2个架构有一个共同点：业务与技术分离。也正是这点解决了微服务开发下业务依赖远程服务的问题。所以才决定由原来的三层架构转为四层架构。


![此处输入图片的描述][1]

## 分层

> Esim使用松散的分层架构，上层可以任意调用下层。

![此处输入图片的描述][2]

### 各层职责

目录 | 职责
---|---
controller | 负责显示信息和解析、校验请求，适配不同终端
application | 不包含业务规则，为下一层领域模型协调任务，分配工作
domain| 负责表达业务概念，业务状态信息和业务规则，是业务软件的核心
infrastructure|为各层提供技术支持，持久化，领域事件等

### 编码规范

- [Uber Style 编码规范](https://github.com/xxjwxc/uber_go_guide_cn)

- 函数第一个参数是ctx！函数第一个参数是ctx！函数第一个参数是ctx！

- 命名

&emsp; | Jaeger
---|---
目录名 |小写/中横线
函数名 |小驼峰
文件名 |下划线
变量 | 小驼峰
常量 | 小驼峰
包名 | 当前目录名
请求地址 | *小写
请求参数 | 小驼峰
返回参数 | 小驼峰


目录 | 定义 | 文件 | 类 | 接口
---|---|---|---|---
application/service |应用层|index.go | IndexService|无
domain/service|领域服务 | index.go | IndexService|无
domain/entity |实体| index.go | Index|无
infra/event |领域事件|index.go | couponEventPub | IndexEvent
infra/repo|资源库|index.go| IndexDbRepo |IndexRepo
infra/dao|数据访问对象| index.go| IndexDao |无


### 数据库设计规范小三样

```mysql
`create_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
`last_update_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
`is_deleted` TINYINT(1) UNSIGNED NOT NULL DEFAULT '0' COMMENT '删除标识',
```


## 特性

- 由三层架构演进为四层架构（DDD + 简洁架构）
- 面向接口编程
- 编译时的依赖注入
- 管控业务使用的网络io
- 融入log，opentracing，metrice提升服务可观察性
- 单元测试友好，面向TDD


## 依赖注入
> Esim 使用[wire](https://github.com/google/wire)实现编译时的依赖注入，它有以下优点：

- 当依赖关系图变得复杂时，运行时依赖注入很难跟踪和调试。 使用代码生成意味着在运行时执行的初始化代码是常规的，惯用的Go代码，易于理解和调试。不会因为框架的各种奇技淫巧而变得生涩难懂。特别重要的是，忘记依赖项等问题会成为编译时错误，而不是运行时错误。
- 与服务定位器不同，不需要费心编造名称来注册服务。 Wire使用Go语法中的类型将组件与其依赖项连接起来。
- 更容易防止依赖项变得臃肿。Wire生成的代码只会导入您需要的依赖项，因此您的二进制文件将不会有未使用的导入。 运行时依赖注入在运行之前无法识别未使用的依赖项。
- Wire的依赖图是静态可知的，这为工具化和可视化提供了可能。

> Esim将wire用于业务与基础设施之间。将基础设施的初始化从业务抽离出来，集中管理。

### Esim使用wire示例
> 基础设施的依赖和初始化都在 ```infra/infra.go``` 文件下。wire的使用主要分2步，以增加mysqlClient：

#### provide
##### before

```golang
type Infra struct {
	*container.Esim
}

var infraSet = wire.NewSet(
	wire.Struct(new(Infra), "*"),
	provideEsim,
)
```

##### after

```golang

type Infra struct {
	*container.Esim

	DB mysql.MysqlClient
}

var infraSet = wire.NewSet(
	wire.Struct(new(Infra), "*"),
	provideEsim,
	provideDb,
)

func provideDb(esim *container.Esim) mysql.MysqlClient {
    ......
	return mysqlClent
}
```

#### Inject

> 在当前目录下执行：```wire```命令，看到：

```linux
wire: projectPath/internal/infra: wrote projectPath/internal/infra/wire_gen.go
```
> 说明执行成功，就可以在项目中使用了。

```golang
infra.NewInfra().DB
```

## 依赖倒置
> 依赖倒置和依赖注入一样，都是应用于业务与基础设施之间。主要的作用是让业务与技术实现分离。
> 在实际的使用中我们把涉及io操作都放到了基础设施的资源库上。这样做的好处：

- 单元测试变简单，使用mock代替数据源
- 不用学习各种 mock sdk，只针对 app 和 domain写单元测试
- 不依赖远程，可以单机进行开发

## 工具

- esim db2entity -d db_name -d table_name

> 前置条件：
> 1. 在项目根目录下
> 2. 配置环境变量 

```linux
export ESIM_DB_HOST=127.0.0.1
export ESIM_DB_PORT=3306
export ESIM_DB_USER=root 
export ESIM_DB_PASSWORD=123456
```
> 由于DDD开发方式多了很多目录，文件，导致这部分工作变得很繁琐，所以```db2entity``` 从mysql数据库的表开始，自动建立实体，生成简单的CRUD语句和资源库的接口与实现，并把生成的资源库注入到基础设施。

- esim model -m modelname

> 前置条件:
> 1. 在模型目录下
> 2. 开启 module， ```export GO111MODULE=on```

> 当项目进入到调优阶段，由于DDD将模型和数据分离，可以单独对模型进行优化。```model``` 命令可以自动对模型进行初始化，内存对齐，生成临时对象池，reset和释放model。很大程度的减少调优花费的时间和心智负担。

## 安装

> 环境 go 1.3 及以上

> 使用 module 包管理工具

> go get github.com/jukylin/esim
> cd github.com/jukylin/esim
> go build -o esim ./tool

## 创建项目

```golang
esim new -s test
```

### 浏览器访问

#### 启动服务
```golang
cd test

go run main.go
```

#### 访问

> http://localhost:8080


### 使用组件测试 [推荐]

```
cd  test/internal/transports/http/component-test/
go test -v -tags="component_test"
```

## 配置
- 配置文件

> 配置文件在项目的conf目录下，分3个环境配置文件

> conf/dev.yaml

> conf/test.yaml

> conf/pro.yaml

- provide

```golang
func provideConf(){
    options := config.ViperConfOptions{}
    env := os.Getenv("ENV")
    if env == "" {
    	env = "dev"
    }

    file := []string{"conf/monitoring.yaml", "conf/" + env + ".yaml"}
    conf := config.NewViperConfig(options.WithConfigType("yaml"),
    	options.WithConfFile(file))

    return conf
}
```

- reference

```golang
service_name := infra.NewInfra().Conf.GetString("appname")
```

## 日志
> 日志会根据不同环境打印，开发和测试环境会把所以日志打印到终端，生产只会打印warn及以上的日志。
> Esim提供了2套日志接口，一套没有上下文，一套有。使用上下文是为了把分布式环境下的日志通过tracer_id串起来。

- provide

```golang
func provideLogger(conf config.Config) log.Logger {
	var loggerOptions log.LoggerOptions

	logger := log.NewLogger(
		loggerOptions.WithDebug(conf.GetBool("debug")),
	)
	return logger}
```

- reference

```golang
infra.NewInfra().Logger.Infof("info %s", "test")

infra.NewInfra().Logger.Infoc(ctx, "info %s", "test")
```



## HTTP
> 比官方接口多了（ctx）参数

- provide

```golang
func provideHttp(esim *container.Esim) *http.HttpClient {
	clientOptions := http.ClientOptions{}
	httpClent := http.NewHttpClient(
		clientOptions.WithTimeOut(esim.Conf.GetDuration("http_client_time_out")),
        clientOptions.WithProxy(
            func() interface {} {
                monitorProxyOptions := http.MonitorProxyOptions{}
                return http.NewMonitorProxy(
                    monitorProxyOptions.WithConf(esim.Conf),
                    monitorProxyOptions.WithLogger(esim.Logger))
            }),
	)

	return httpClent
}
```

- reference

```golang
resp, err := infra.NewInfra().Http.GetCtx(ctx, "http://www.baidu.com")
defer resp.Body.Close()

```


## Mongodb
> 文档 https://github.com/mongodb/mongo-go-driver

- provide

```golang
func provideMongodb(esim *container.Esim) mongodb.MgoClient {
	options := mongodb.MgoClientOptions{}
	mongo := mongodb.NewMongo(
		mgoClientOptions.WithLogger(esim.logger),
		mgoClientOptions.WithConf(esim.conf),
		mgoClientOptions.WithMonitorEvent(
			func() MonitorEvent {
				monitorEventOptions := MonitorEventOptions{}
				return NewMonitorEvent(
					monitorEventOptions.WithConf(esim.conf),
					monitorEventOptions.WithLogger(esim.logger),
				)
			},
		)
	)

	return mongo
}
```

- reference

```golang

import "go.mongodb.org/mongo-driver/bson"

type Info struct{
	Title string
}


info := Info{}

coll := infra.NewInfra().Mgo.GetColl("database", "coll")
filter := bson.M{"phone": "123456"}
res := coll.FindOne(inf.Mgo.GetCtx(c.Request.Context()), filter).Decode(&info)

```

## GRPC
> 文档 https://github.com/grpc/grpc-go

- provide

```golang
func provideGrpcClient(esim *container.Esim) *grpc.GrpcClient {

	clientOptional := grpc.ClientOptionals{}
	clientOptions := grpc.NewClientOptions(
		clientOptional.WithLogger(esim.Logger),
		clientOptional.WithConf(esim.Conf),
	)

	grpcClient := grpc.NewClient(clientOptions)

	return grpcClient
}
```

- reference

```golang
import (
    "pathto/protobuf/passport"
)

conn := infra.NewInfra().GrpcClient.DialContext(ctx, ":60080")
defer conn.Close()

client := passport.NewUserInfoClient(conn)

getUserByUserNameRequest := &passport.GetUserByUserNameRequest{}
getUserByUserNameRequest.Username = "123456"

replyData, err = client.GetUserByUserName(ctx, getUserByUserNameRequest)
```

## Redis
> 文档 https://github.com/gomodule/redigo

- provide

```golang
func provideRedis(esim *container.Esim) *redis.RedisClient {
	redisClientOptions := redis.RedisClientOptions{}
	redisClent := redis.NewRedisClient(
		redisClientOptions.WithConf(esim.Conf),
		redisClientOptions.WithLogger(esim.Logger),
		redisClientOptions.WithProxy(
			func() interface{} {
				monitorProxyOptions := redis.MonitorProxyOptions{}
				return redis.NewMonitorProxy(
					monitorProxyOptions.WithConf(esim.Conf),
					monitorProxyOptions.WithLogger(esim.Logger),
					monitorProxyOptions.WithTracer(esim.Tracer),
				)
			},
		),
	)

	return redisClent
}
```

- reference

```golang

"gitlab.etcchebao.cn/go_service/esim/pkg/redis"

conn := infra.NewInfra().Redis.GetCtxRedisConn()
defer conn.Close()
key := "username:"+username
exists, err := redis.Bool(conn.Do(ctx, "exists", key))
```

## Mysql
> 文档 https://gorm.io/docs/

- provide

```golang
func provideDb(esim *container.Esim) *mysql.MysqlClient {

    mysqlClientOptions := mysql.MysqlClientOptions{}
	mysqlClent := mysql.NewMysqlClient(
		mysqlClientOptions.WithConf(esim.Conf),
		mysqlClientOptions.WithLogger(esim.Logger),
		mysqlClientOptions.WithProxy(
			func() interface{} {
				monitorProxyOptions := mysql.MonitorProxyOptions{}
				return mysql.NewMonitorProxy(
					monitorProxyOptions.WithLogger(esim.Logger),
					monitorProxyOptions.WithConf(esim.Conf),
					monitorProxyOptions.WithTracer(esim.Tracer),
				)
			},
		),
	)

	return mysqlClent
}
```

- reference

```golang
var user model.User
infra.NewInfra().DB.GetDb(ctx, "db").Table("table").Where("username = ?", username).
	Select([]string{"id"}).First(&user)
```

## Opentracing
> Esim使用[Jaeger](https://www.jaegertracing.io/)实现分布式追踪，默认为关闭状态。

>开启需要使用jaeger-client-go自带的[环境变量](https://github.com/jaegertracing/jaeger-client-go#environment-variables)


  [1]: https://imgconvert.csdnimg.cn/aHR0cHM6Ly9hdGEyLWltZy5jbi1oYW5nemhvdS5vc3MtcHViLmFsaXl1bi1pbmMuY29tL2EzM2I4MGJjYWM1ZWM3M2QwZDEzNThkNmI0OWExMTljLnBuZw?x-oss-process=image/format,png
  [2]: https://upload.cc/i1/2019/12/26/86caKj.png
