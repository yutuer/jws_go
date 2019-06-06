#openssl genrsa -out root/ca.key 2048 
#openssl req -new -key root/ca.key -out root/ca.csr -subj "/C=CN/ST=GZ/L=ZH/O=KSS/OU=KSS/CN=taiyouxi"
#openssl x509 -req -days 3650 -in root/ca.csr -signkey root/ca.key -out root/ca.crt
#openssl ca -gencrl -out root/ca.crl -crldays 7  -config ./openssl.cnf


openssl genrsa -out server/server.key 1024
openssl req -new -key server/server.key -out server/server.csr -subj "/C=CN/ST=GZ/L=ZH/O=KSS/OU=KSS/CN=taiyouxi"
openssl ca -in server/server.csr -cert root/ca.crt -keyfile root/ca.key -out server/server.crt -days 3650  -config ./openssl.cnf

openssl genrsa -out client/client.key 1024
openssl req -new -key client/client.key -out client/client.csr -subj "/C=CN/ST=GZ/L=ZH/O=KSS/OU=KSS/CN=taiyouxi" -extensions v3_req
openssl ca -in client/client.csr -cert root/ca.crt -keyfile root/ca.key -out client/client.crt -days 3650 -extensions v3_req  -config ./openssl.cnf
openssl pkcs12 -export -inkey client/client.key -in client/client.crt -out client/client.pfx
openssl pkcs12 -in client/client.pfx -out client/client.pem -nodes


