#!/bin/bash

set -e
protoc -I./pb --go-grpc_out="." --go_out="." test.proto


