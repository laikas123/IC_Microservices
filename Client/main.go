
package main

import (
	"context"
	"flag"
	"io"
	"log"
	"math/rand"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	//again we provide an alias for our directory
	pb "github.com/laikas123/IC_Microservices/ProtoFiles"
	

	//these are just test credentials found online
	"google.golang.org/grpc/testdata"
)


//small thing to point out the Client name is uppercase beginning for 
//an interface and is lowercase for a struct 
func getLocationData(client pb.ICCalculatorServiceClient, status *pb.LocationStatus) {
	log.Printf("Getting feature for point (%d, %d)", point.Latitude, point.Longitude)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	
	defer cancel()
	
	
	statusReturned, err := client.QueryLocations(ctx, status)
	if err != nil {
		log.Fatalf("%v.QueryLocations(_) = _, %v: ", client, err)
	}

	log.Println(statusReturned)
}



func main() {
	var opts []grpc.DialOption
	if *tls {
		if *caFile == "" {
			*caFile = testdata.Path("ca.pem")
		}
		creds, err := credentials.NewClientTLSFromFile(*caFile, *serverHostOverride)
		if err != nil {
			log.Fatalf("Failed to create TLS credentials %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()

	//this is where we create our inital RouteGuideClient 
	//which in turn calls gets passed as RouteGuideClient to all methods
	client := pb.NewRouteGuideClient(conn)

	pointLo := pb.Point{
		Latitude: 0,
		Longitude: 0
	}
	pointHi := pb.Point{
		Latitude: 10,
		Longitude: 10
	}

	newRectangle := pb.Rectangle{
		lo: pointLo,
		hi: pointHi
	}

	// Looking for a valid feature
	getLocationData(client, &pb.LocationStatus{usersonline: 10, locationtoserve: newRectangle})

	
}
