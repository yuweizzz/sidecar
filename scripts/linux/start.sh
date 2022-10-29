#!/bin/bash

PWD=`pwd`

./sidecar-server start -dir $PWD -conf $PWD/conf.toml
