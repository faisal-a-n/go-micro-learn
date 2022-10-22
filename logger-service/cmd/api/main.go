package main

import (
	"context"
	"log"
	"logger-service/data"
	"net/http"
	"net/rpc"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	PORT     = "0.0.0.0:80"
	rpcPORT  = "0.0.0.0:5001"
	mongoURL = "mongodb://mongo:27017"
	grpcPORT = "0.0.0.0:50001"
)

var client *mongo.Client

type Config struct {
	Models data.Models
}

func main() {
	//Connect to mongo
	mongoClient, err := connectToMongoDB()
	if err != nil {
		log.Fatal("Error connecting to mongo", err)
	}
	client = mongoClient

	//Create context to disconnect
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	//close connection
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	app := Config{
		Models: data.New(client),
	}

	//Register RPC connection
	err = rpc.Register(new(RPCServer))
	if err != nil {
		log.Println("Cannot register rpc server", err)
	}
	go app.rpcListen()

	go app.gRPCListen()

	app.serve()
}

func (app *Config) serve() {
	server := &http.Server{
		Addr:    PORT,
		Handler: app.routes(),
	}
	log.Println("Service running in", PORT)
	err := server.ListenAndServe()

	if err != nil {
		log.Panic("Cannot start service", err)
	}
}

func connectToMongoDB() (*mongo.Client, error) {
	//connection options
	clientOptions := options.Client().ApplyURI(mongoURL)
	clientOptions.SetAuth(options.Credential{
		Username: "admin",
		Password: "password",
	})

	//connect
	c, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, err
	}
	log.Println("MongoDB connected!")
	return c, nil
}
