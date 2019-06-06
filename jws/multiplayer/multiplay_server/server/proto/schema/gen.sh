#!/bin/bash

#GVE protocols
pushd multiplayMsg
flatc -g -o ../../ *.fbs *.fb
flatc -n -o ../../client/ *.fbs *.fb
popd