#!/bin/bash 

CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -o atcd_tools_linux_arm

tar cvzf atcd_tools.tar.gz  atcd_tools_linux_arm conf public README.md