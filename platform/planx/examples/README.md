## What is packet*.bin?
TODO: change to little endian
用于进行协议Packet测试和Unit Test的二进制包

在服务器启动的情况下可以通过如下命令进行发包
cat packet_1.bin | nc -nv 127.0.0.1 6666


### packet_chat.bin & packet_chab.bin

CHAT
cat packet_chat.bin | nc -nv 127.0.0.1 6666

CHAB
cat packet_chab.bin | nc -nv 127.0.0.1 6666
使用了msgpack的方式进行encode chat string.