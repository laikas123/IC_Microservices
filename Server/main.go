
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/testdata"

	"github.com/golang/protobuf/proto"


    //we have given our routeguide directory an alias pb
	pb "github.com/laikas123/IC_Microservices/ProtoFiles"
)


type routeGuideServer struct {

	mu         sync.Mutex // protects routeNotes
	
}


func (s *routeGuideServer) QueryLocation(ctx context.Context, point *pb.LocationStatus) (*pb.Feature, error) {
	
	return &pb.Feature{Location: point}, nil
}


func (s *routeGuideServer) CalculateDistance(ctx context.Context, req *TwoPoints) (*Number, error) {
    return nil, status.Errorf(codes.Unimplemented, "method CalculateDistance not implemented")
}


func (s *routeGuideServer) CalculateGasLoss(ctx context.Context, req *Number) (*Number, error) {
    return nil, status.Errorf(codes.Unimplemented, "method CalculateGasLoss not implemented")
}


func (s *routeGuideServer) CalculateLocationProfit(ctx context.Context, req *Number) (*Number, error) {
    return nil, status.Errorf(codes.Unimplemented, "method CalculateLocationProfit not implemented")
}


func newServer() *routeGuideServer {
	s := &routeGuideServer{}}
	
    
	return s
}

func main() {
    //this is where the entire thing starts so parse the flags
	flag.Parse()
    //listen via tcp on the port specified
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
    //A ServerOption sets options such as credentials, codec and keepalive parameters, etc.
    //here we have a variable for a slice of all our ServerOptions
	var opts []grpc.ServerOption
    //note we are referencing the variable tls which is 
    //defined by our command line flags
	if *tls {
		if *certFile == "" {
            //Path just gets the filepath
			*certFile = testdata.Path("server1.pem")
		}
       
		if *keyFile == "" {
            //path just gets the filepath
			*keyFile = testdata.Path("server1.key")
		}
		creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
		if err != nil {

			log.Fatalf("Failed to generate credentials %v", err)
		}
     
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}

	grpcServer := grpc.NewServer(opts...)
 
	pb.RegisterRouteGuideServer(grpcServer, newServer())
  
	grpcServer.Serve(lis)
}

