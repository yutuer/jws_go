#!/bin/bash
#export GO15VENDOREXPERIMENT=0
godep save -t $(go list ./... | grep -vf buildignore)
