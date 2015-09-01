# game(逻辑)
[![Build Status](https://travis-ci.org/gonet2/game.svg?branch=master)](https://travis-ci.org/gonet2/game)

## 设计理念
游戏服务器对agent只提供一个接口， 即:

> rpc Stream(stream Game.Frame) returns (stream Game.Frame);

接收来自agent的Frame的双向流

来自设备的数据包，通过agent后直接透传到game server, Frame大体分为两类：  

1. 链路控制（register, kick)     
2. 来自设备的经过agent解密后的数据包 (message)       

数据包(message)格式为:      

> 协议号＋数据

        +----------------------------------+     
        | PROTO(2) | PAYLOAD(n)            |     
        +----------------------------------+     

在client_handler目录中绑定对应函数进行处理，协议生成和绑定通过tools目录中的脚本进行。

协议的绑定参考 https://github.com/gonet2/tools/tree/master/proto_scripts


## 使用
参考测试用例以及game.proto文件

## 安装
参考Dockerfile

# 环境变量
> NSQD_HOST: eg : http://172.17.42.1:4151
