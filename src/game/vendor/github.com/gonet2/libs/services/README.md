#services
a drop-in services discovery library based on etcd

# pb.go产生方式
protoc ./*.proto --go_out=plugins=grpc:./

#etcd目录结构
etcd目录结构采用 http://gliderlabs.com/registrator/latest/ 提供的结构:          

>    /backends/service_xxx/service_id ---> ip:port

调用 Init(..) 将服务发现限定在给定范围

![services](services.png)
