###
 # @Descripttion: 
 # @version: 
 # @Author: cm.d
 # @Date: 2021-11-12 21:49:50
 # @LastEditors: cm.d
 # @LastEditTime: 2021-11-13 01:24:48
### 

#!/bin/bash

./AlfheimDB --httpserver_addr=localhost:12345 --raft_addr=localhost:40000 --raft_id=id1 --raft_cluster=localhost:40000/id1,localhost:40001/id2,localhost:40002/id3 --respserver_addr=localhost:6379