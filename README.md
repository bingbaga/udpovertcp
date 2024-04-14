# udpovertcp


#### client 模式为本地代理udp模式

example  
本地运行命令
test -local 0.0.0.0:22226 -mode client -remote "[2406::3103]:33366"
    本地监听udp端口22226  通过tcp 发往远程服务器 [2406::3103]:33366
vps运行命令
test -local 0.0.0.0:33366 -remote 162.159.195.28:2408 -mode server
      server监听tcp 33366端口  把收到的数据通过udp发送到162.159.195.28:2408