#!/bin/bash 
npm run build


go build -o gm_tools

tar cvzf gm_tools.tar.gz  gm_tools conf public README.md