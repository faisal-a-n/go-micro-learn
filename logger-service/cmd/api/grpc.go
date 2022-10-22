package main

import (
	"context"
	"log"
	"logger-service/data"
	"logger-service/logs"
	"net"
	"time"

	"google.golang.org/grpc"
)

type LogServer struct {
	logs.UnimplementedLogServiceServer
	Models data.Models
}

func (this *LogServer) WriteLog(ctx context.Context, req *logs.LogRequest) (*logs.LogResponse, error) {
	input := req.GetLogEntry()
	//write log
	err := this.Models.LogEntry.Insert(data.LogEntry{
		Name:      input.Name,
		Data:      input.Data,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		res := &logs.LogResponse{
			Result: "Failed",
		}
		return res, err
	}
	res := &logs.LogResponse{
		Result: "Inserted log",
	}
	return res, nil
}

func (app *Config) gRPCListen() {
	log.Println("Starting gRPC on ", grpcPORT)
	listen, err := net.Listen("tcp", grpcPORT)
	if err != nil {
		log.Println("Can't start gRPC ", err)
	}
	server := grpc.NewServer()

	//LogServer implements thte actions of the service
	logs.RegisterLogServiceServer(server, &LogServer{
		Models: app.Models,
	})

	log.Println("Started gRPC on ", grpcPORT)

	if err := server.Serve(listen); err != nil {
		log.Println("Can't start gRPC server", err)
	}
}
