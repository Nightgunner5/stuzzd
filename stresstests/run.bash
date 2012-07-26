#!/bin/bash

echo "# ops.txt (creating a file that would cause warnings if it wasn't created yet)"
touch ops.txt

echo "# noop (makes sure the required packages are compiled before testing starts)"
time go run noop.go

echo "# GOMAXPROCS=1 chunks (10000 chunks in parallel with 1 thread)"
time GOMAXPROCS=1 go run chunks.go

echo "# GOMAXPROCS=2 chunks (10000 chunks in parallel with 2 threads)"
time GOMAXPROCS=2 go run chunks.go

echo "# GOMAXPROCS=4 chunks (10000 chunks in parallel with 4 threads)"
time GOMAXPROCS=4 go run chunks.go

echo "# GOMAXPROCS=8 chunks (10000 chunks in parallel with 8 threads)"
time GOMAXPROCS=8 go run chunks.go

echo "# GOMAXPROCS=16 chunks (10000 chunks in parallel with 16 threads)"
time GOMAXPROCS=16 go run chunks.go

echo "# GOMAXPROCS=32 chunks (10000 chunks in parallel with 32 threads)"
time GOMAXPROCS=32 go run chunks.go

