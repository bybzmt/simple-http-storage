# Simple-http-Storage

这是一个简单的HTTP的文件存储服务，这个服务是为跨服务器保存文件而设计的。
仅仅是为了不在php服务器上通过mount挂载远程文件。

这样子做会有两点好处：
1. 运维部署会变得简单
2. 程序员会意识到他操作的文件是远程的，不会无意识的在网络上反复下载文件。

服务只支持4个操作
* HEAD 读取头信息，判断文件是否存在
* GET  得到文件
* PUT  上传文件
* DELETE 删除文件

# phpclient
这是一个简单http客户端
