# client  
  
import socket 
import time 
  
address = ('127.0.0.1', 6666)  
s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)  
s.connect(address)  
  
data = s.recv(512)  
print 'the data received is',data  
  
s.send('hihi')  
time.sleep(20)
s.close()  