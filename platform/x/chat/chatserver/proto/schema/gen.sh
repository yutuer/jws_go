#!/bin/bash

flatc -g -o ../ RegReq.fb RegResp.fb ProtoWrap.fb *Notify.fb *Msg.fb
#flatc -g -o ../ *.fb

