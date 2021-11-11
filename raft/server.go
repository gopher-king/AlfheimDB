/*
 * @Descripttion:
 * @version:
 * @Author: cm.d
 * @Date: 2021-11-11 18:00:19
 * @LastEditors: cm.d
 * @LastEditTime: 2021-11-12 00:06:07
 */

package raft

import (
	"net"
	"os"
	"path/filepath"

	"github.com/AlfheimDB/config"
	"github.com/hashicorp/raft"
	boltdb "github.com/hashicorp/raft-boltdb"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/Jille/raft-grpc-leader-rpc/leaderhealth"
	transport "github.com/Jille/raft-grpc-transport"
	"github.com/Jille/raftadmin"
)

type AlfheimRaftServer struct {
	RaftId  string
	MyIP    string
	MyPort  string
	RaftDir string
	RaftFsm raft.FSM
	Raft    *raft.Raft
}

var RaftServer *AlfheimRaftServer

func Init() {
	logrus.Info("init raft server ")
	initRaft(config.Config.RaftAddr, config.Config.RaftDir, config.Config.RaftId)

}

func initRaft(address string, raftDir string, raftId string) {
	RaftServer = new(AlfheimRaftServer)
	ip, port, err := net.SplitHostPort(config.Config.RaftAddr)
	if err != nil {
		logrus.Fatal("Unknow ip and port", config.Config.RaftAddr)
	}
	logrus.Info("Bind address Ip: ", ip, " ,port: ", port)
	RaftServer.MyIP = ip
	RaftServer.MyPort = port
	RaftServer.RaftId = raftId
	RaftServer.RaftDir = raftDir
	sock, err := net.Listen("tcp", config.Config.RaftAddr)
	if err != nil {
		logrus.Fatal("Listen port error", err)
	}
	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = raft.ServerID(raftId)
	baseDir := filepath.Join(raftDir, raftId)
	ldb, err := boltdb.NewBoltStore(filepath.Join(baseDir, "logs.dat"))
	if err != nil {
		logrus.Fatal("Init log db error", err)
	}
	sdb, err := boltdb.NewBoltStore(filepath.Join(baseDir, "stable.dat"))
	if err != nil {
		logrus.Fatal("Init stable db error", err)
	}
	fss, err := raft.NewFileSnapshotStore(baseDir, 1, os.Stderr)
	if err != nil {
		logrus.Fatal("Init snapshot dir error", err)
	}

	fsm := AlfheimRaftFSMImpl{}
	RaftServer.RaftFsm = &fsm

	tm := transport.New(raft.ServerAddress(address), []grpc.DialOption{grpc.WithInsecure()})

	raftIns, err := raft.NewRaft(raftConfig, &fsm, ldb, sdb, fss, tm.Transport())
	if err != nil {
		logrus.Fatal("Init raft instance error", err)
	}
	RaftServer.Raft = raftIns
	// cfg := raft.Configuration{
	// 	Servers: []raft.Server{
	// 		{
	// 			Suffrage: raft.Voter,
	// 			ID:       raft.ServerID(raftId),
	// 			Address:  raft.ServerAddress(address),
	// 		},
	// 	},
	// }
	// raftFuture := raftIns.BootstrapCluster(cfg)
	// if err := raftFuture.Error(); err != nil {
	// 	logrus.Fatal("Bootstrap raft cluster error", err)
	// }
	grpcServer := grpc.NewServer()
	tm.Register(grpcServer)
	leaderhealth.Setup(raftIns, grpcServer, []string{"Example"})
	raftadmin.Register(grpcServer, raftIns)
	reflection.Register(grpcServer)
	logrus.Info("raft init success")
	if err := grpcServer.Serve(sock); err != nil {
		logrus.Fatal("Grpc serve sock error, ", err)
	}
}
