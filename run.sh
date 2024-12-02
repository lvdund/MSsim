#!/bin/bash

# sudo sysctl -w net.core.rmem_max=33554432
# sudo sysctl -w net.core.wmem_max=33554432
# sudo sysctl -w net.ipv4.tcp_rmem='4096 65536 33554432'
# sudo sysctl -w net.ipv4.tcp_wmem='4096 65536 33554432'
# sysctl -w net.core.somaxconn=10240
# sysctl -w net.core.netdev_max_backlog=10000

go build cmd/packetrusher.go ; sudo ./packetrusher multi-ue -n 1000 -tr 1 -nPdu 5