package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/hjhussaini/storage/proto"
	"github.com/hjhussaini/storage/rpc"
	"google.golang.org/grpc"
)

func main() {
	address := os.Getenv("ADDRESS")
	path := os.Getenv("STORAGE_PATH")

	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Cannot listen on %s: %s\n", address, err.Error())
	}

	options := []grpc.ServerOption{}
	server := grpc.NewServer(options...)

	storage := rpc.Storage{
		Path: path,
	}
	proto.RegisterStorageServer(server, &storage)

	errs := make(chan error, 2)
	go func() {
		log.Printf("Serving gRPC server on %s\n", address)
		if err = server.Serve(listener); err != nil {
			errs <- fmt.Errorf("Could not serve gRPC server: %s", err)
		}
		errs <- nil
	}()

	go func() {
		quit := make(chan os.Signal)
		signal.Notify(quit, os.Interrupt)
		signal.Notify(quit, os.Kill)

		<-quit
		server.Stop()
		listener.Close()
	}()

	if err := <-errs; err != nil {
		log.Fatal(err)
	}
}
