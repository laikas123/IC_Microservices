
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
	pb "github.com/laikas123/IC_Microservices/ProtoFiles/"
	

	//these are just test credentials found online
	"google.golang.org/grpc/testdata"
)


var (
	tls                = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	caFile             = flag.String("ca_file", "", "The file containing the CA root cert file")
	serverAddr         = flag.String("server_addr", "127.0.0.1:10000", "The server address in the format of host:port")
	serverHostOverride = flag.String("server_host_override", "x.test.youtube.com", "The server name use to verify the hostname returned by TLS handshake")
)


func printFeature(client pb.RouteGuideClient, point *pb.Point) {
	log.Printf("Getting feature for point (%d, %d)", point.Latitude, point.Longitude)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	
	defer cancel()
	
	
	feature, err := client.GetFeature(ctx, point)
	if err != nil {
		log.Fatalf("%v.GetFeatures(_) = _, %v: ", client, err)
	}

	log.Println(feature)
}

// printFeatures lists all the features within the given bounding Rectangle.
func printFeatures(client pb.RouteGuideClient, rect *pb.Rectangle) {
	log.Printf("Looking for features within %v", rect)
	//another timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	stream, err := client.ListFeatures(ctx, rect)
	if err != nil {
		log.Fatalf("%v.ListFeatures(_) = _, %v", client, err)
	}
	for {
		//we receive each value from the server in a for loop
		feature, err := stream.Recv()
		//once the server is done writing break the loop
		//so we don't receive null
		if err == io.EOF {
			break
		}
		//if we aborted due to unexpected error list the features we 
		//got and print the client and error that occurred
		if err != nil {
			//note log.Fatalf causes the program to terminate
			log.Fatalf("%v.ListFeatures(_) = _, %v", client, err)
		}
		//print each feature as it comes in
		log.Println(feature)
	}
}

// runRecordRoute sends a sequence of points to server and expects to get a RouteSummary from server.
//in this case we stream to the server and expect to get a result back
func runRecordRoute(client pb.RouteGuideClient) {
	// Create a random number of random points
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	pointCount := int(r.Int31n(100)) + 2 // Traverse at least two points
	var points []*pb.Point
	for i := 0; i < pointCount; i++ {
		points = append(points, randomPoint(r))
	}
	log.Printf("Traversing %d points.", len(points))

	//create a context with timeout and cancel immediately
	//Note Background() returns a non-nil, empty Context.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//RecordRoute() takes a context and opens a new client to server stream
	//this returns to us the struct which contains the clientConn as a field
	//it also has Send() and CloseAndRecv() messages
	stream, err := client.RecordRoute(ctx)
	if err != nil {
		log.Fatalf("%v.RecordRoute(_) = _, %v", client, err)
	}
	//for the random points we created we loop through
	//we send each point down the wire to the server
	for _, point := range points {
		if err := stream.Send(point); err != nil {
			log.Fatalf("%v.Send(%v) = %v", stream, point, err)
		}
	}
	//we then call CloseAndRecv once we are done sending points
	//this method closes the sending direction so the client can't send(see our pb.go file)
	//and it also generates a routeSummary
	//we get our routeSummary as a pointer to a RouteSummary
	reply, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
	}
	log.Printf("Route summary: %v", reply)
}

// runRouteChat receives a sequence of route notes, while sending notes for various locations.
// the streams from server to client and client to server operate independently
// 
func runRouteChat(client pb.RouteGuideClient) {
	notes := []*pb.RouteNote{
		{Location: &pb.Point{Latitude: 0, Longitude: 1}, Message: "First message"},
		{Location: &pb.Point{Latitude: 0, Longitude: 2}, Message: "Second message"},
		{Location: &pb.Point{Latitude: 0, Longitude: 3}, Message: "Third message"},
		{Location: &pb.Point{Latitude: 0, Longitude: 1}, Message: "Fourth message"},
		{Location: &pb.Point{Latitude: 0, Longitude: 2}, Message: "Fifth message"},
		{Location: &pb.Point{Latitude: 0, Longitude: 3}, Message: "Sixth message"},
	}
	//new context and cancel
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	//call the method from our pb.go file 
	//this method will end up creating both streams
	//note although this is one variable in our pb.go file 
	//we define this to a be bidirectional stream object 
	//so this variable holds two streams 
	//they operate independently I believe this means that rather 
	//than have two stream objects that it just means that 
	//we can call send in either direction and receive in either 
	//direction and it will function without issue
	//the reason we even probably differentiate in the first place 
	//is because it's less room to hold for instance a one way stream
	//vs two independent streams so yet another small detail where we 
	//are being smart with our resources

	//this creates a new stream with bidirectional capability
	stream, err := client.RouteChat(ctx)
	if err != nil {
		log.Fatalf("%v.RouteChat(_) = _, %v", client, err)
	}

	//so we create a channel
	//then we run a goroutine
	waitc := make(chan struct{})


	go func() {

		//we Recv() the message from the server until it's done
		//and make sure to close the channel once it is to avoid leaving
		//it open
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				// read done.
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive a note : %v", err)
			}
			log.Printf("Got message %s at point(%d, %d)", in.Message, in.Location.Latitude, in.Location.Longitude)
		}
	}()
	//we also send notes to the server 
	for _, note := range notes {
		if err := stream.Send(note); err != nil {
			log.Fatalf("Failed to send a note: %v", err)
		}
	}

	//note we use CloseSend() which closes the send direction of the stream
	//for here that is client to server
	stream.CloseSend()

	//receives after the channel has been closed will return the zero
	//type for the channel
	//so what's good about this is that we keep this function running
	//until the channel is closed which means the server is done writing to us
	<-waitc
}

func randomPoint(r *rand.Rand) *pb.Point {
	lat := (r.Int31n(180) - 90) * 1e7
	long := (r.Int31n(360) - 180) * 1e7
	return &pb.Point{Latitude: lat, Longitude: long}
}

func main() {
	flag.Parse()
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

	// Looking for a valid feature
	printFeature(client, &pb.Point{Latitude: 409146138, Longitude: -746188906})

	// Feature missing.
	printFeature(client, &pb.Point{Latitude: 0, Longitude: 0})

	// Looking for features between 40, -75 and 42, -73.
	printFeatures(client, &pb.Rectangle{
		Lo: &pb.Point{Latitude: 400000000, Longitude: -750000000},
		Hi: &pb.Point{Latitude: 420000000, Longitude: -730000000},
	})

	// RecordRoute
	runRecordRoute(client)

	// RouteChat
	runRouteChat(client)
}
