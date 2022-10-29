#!/bin/bash

PWD=`pwd`

./sidecar-server start -daemon -dir $PWD -conf $PWD/conf.toml
