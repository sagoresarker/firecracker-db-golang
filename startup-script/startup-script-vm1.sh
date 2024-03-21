#!/bin/bash

# Add an IP address to the eth0 interface
ip addr add 172.16.0.2/24 dev eth0

# Bring the eth0 interface up
ip link set eth0 up

# Add a default route via the gateway at 172.16.0.1
ip route add default via 172.16.0.1 dev eth0
