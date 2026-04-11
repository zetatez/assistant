#!/usr/bin/env bash

mysql -uroot -pAAaa00__ -P3306 --ssl=0 -c -A -e "drop database assistant"

mysql -uroot -pAAaa00__ -P3306 --ssl=0 -c -A -e "create database assistant"
