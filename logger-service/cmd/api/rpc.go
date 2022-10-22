package main

import (
	"context"
	"fmt"
	"log"
	"logger-service/data"
	"net"
	"net/rpc"
	"time"
)

//RPCServer exposes it's function for RPC calls
type RPCServer struct {
}

type RPCPayload struct {
	Name string
	Data string
}

func (server *RPCServer) LogInfo(payload RPCPayload, res *string) error {
	collection := client.Database("logs").Collection("logs")
	_, err := collection.InsertOne(context.TODO(), data.LogEntry{
		Name:      payload.Name,
		Data:      payload.Data,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		fmt.Println("Error logging grpc", err)
		return err
	}
	*res = "Processed payload via RPC " + payload.Name
	return nil
}

func (app *Config) rpcListen() {
	log.Println("Starting RPC on ", rpcPORT)
	listen, err := net.Listen("tcp", rpcPORT)
	if err != nil {
		log.Println("Can't start RPC ", err)
	}
	defer listen.Close()
	for {
		rpcConn, err := listen.Accept()
		if err != nil {
			continue
		}
		go rpc.ServeConn(rpcConn)
	}

}
