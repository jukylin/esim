package new

func ConfigInit() {
	fc1 := &FileContent{
		FileName: "conf.yaml",
		Dir:      "conf",
		Content: `
runmode : '${ENV}'
appname : {{service_name}}

#HTTP 服务
httpport : 8080

#服务端
grpc_server_tcp : 50055
grpc_server_kp_time : 60
grpc_server_kp_time_out : 5
#链接超时
grpc_server_conn_time_out : 3

#客户端
grpc_client_kp_time : 60
grpc_client_kp_time_out : 5
#链接超时 单位：ms
grpc_client_conn_time_out : 300
grpc_client_permit_without_stream: true

jaeger_disabled: '${JAEGER_DISABLED}'
jaeger_local_agent_host_port: '0.0.0.0:6831'

#mysql
dbs:
#- {db: 'test', dsn: 'root:123456@tcp(0.0.0.0:3306)/config?charset=utf8&parseTime=True&loc=Local',
#  maxidle: 10, maxopen: 100, maxlifetime : 10}
#- {db: 'test_slave', dsn: 'root:123456@tcp(0.0.0.0:3306)/config?charset=utf8&parseTime=True&loc=Local',


#mongodb
#mgos:
#- {db: 'test', uri: 'mongodb://0.0.0.0:27017/admin?connect=direct'}

#ms
mgo_connect_timeout : 500
#min
mgo_max_conn_idle_time : 10
mgo_max_pool_size : 100
mgo_min_pool_size : 10

# http请求 单位：s
http_client_time_out : 3


#redis
redis_max_active : 50
redis_max_idle : 100
redis_idle_time_out : 600
redis_etc1_host : '${REDIS_ETC1_HOST}'
redis_etc1_post : '${REDIS_ETC1_PORT}'
redis_etc1_password : '${REDIS_ETC1_PASSWORD}'

#redis 读超时 单位：ms
redis_read_time_out : 500
#redis 写超时 单位：ms
redis_write_time_out : 500
#redis 连接超时 单位：ms
redis_conn_time_out : 500


#prometheus http addr
prometheus_http_addr : 9002
`,
	}

	fc2 := &FileContent{
		FileName: "monitoring.yaml",
		Dir:      "conf",
		Content: `
#grpc 服务端
#开启慢检查 true/false
grpc_server_check_slow : {{bool}}
#单位: ms
grpc_server_slow_time : 500
#开启 tracer true/false
grpc_server_tracer : {{bool}}
#启动metrice true/false
grpc_server_metrics : {{bool}}
#启动debug true/false
grpc_server_debug: {{bool}}

#grpc 客户端
#开启慢检查 true/false
grpc_client_check_slow : {{bool}}
#单位: ms
grpc_client_slow_time : 500
#开启 tracer true/false
grpc_client_tracer : {{bool}}
#启动metrice true/false
grpc_client_metrics : {{bool}}
#启动debug true/false
grpc_client_debug: {{bool}}

#mysql
#开启慢检查 true/false
mysql_check_slow : {{bool}}
# 大于 slow_sql_time 为慢sql 单位：ms
mysql_slow_time : 500
#开启 tracer true/false
mysql_tracer : {{bool}}
#启动metrice true/false
mysql_metrics : {{bool}}

#mongodb
#开启慢检查 true/false
mgo_check_slow : {{bool}}
# 大于 slow_sql_time 为慢命令 单位：ms
mgo_slow_time : 250
#开启 tracer true/false
mgo_tracer : {{bool}}
#启动metrice true/false
mgo_metrics : {{bool}}


# http client
#开启慢检查 true/false
http_client_check_slow : {{bool}}
#  单位：ms
http_client_slow_time : 500
#开启 tracer true/false
http_client_tracer : {{bool}}
#启动metrice true/false
http_client_metrics : {{bool}}

# http server
#开启 tracer true/false
http_tracer: {{bool}}
#启动metrice true/false
http_metrics: {{bool}}

#redis
#开启慢检查 true/false
redis_check_slow : {{bool}}
#慢命令 单位 ms
redis_slow_time : 50
#开启 tracer true/false
redis_tracer : {{bool}}
#启动metrice true/false
redis_metrics : {{bool}}
`,
	}

	Files = append(Files, fc1, fc2,)
}
