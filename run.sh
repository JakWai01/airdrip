#!/bin/bash

case $(uname -s) in
Linux)
STR=$(ip a | grep "inet " | grep -Fv 127.0.0.1 | awk '{print $2}')
IFS='/ ' read -r -a array <<< $STR 
STR="${array[0]}"
;;
Darwin)
STR=$(ifconfig | grep "inet " | grep -Fv 127.0.0.1 | awk '{print $2}')
;;
esac

export LADDR=$STR
echo "$LADDR"
