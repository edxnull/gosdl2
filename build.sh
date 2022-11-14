#!/bin/bash

execname=gosdl2
path=bin/$execname

go build -o bin/$execname ./
bin/./$execname
