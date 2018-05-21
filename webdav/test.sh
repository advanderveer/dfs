#!/bin/bash

go test -v -race
if [ $? -ne 0 ]; then
    echo "--> UNIT TESTS FAILED"
    exit
fi

cd litmus
make URL=http://localhost:9999/ check
if [ $? -ne 0 ]; then
    echo "--> INTEGRATION TESTS FAILED"
    exit
fi

cd ..
go test -v -race -bench=.*
if [ $? -ne 0 ]; then
    echo "--> BENCHMARKS FAILED"
    exit
fi
