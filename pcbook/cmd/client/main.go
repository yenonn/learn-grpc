package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/yenonn/pcbook/pb"
	"github.com/yenonn/pcbook/sample"
)

func createLaptop(laptopClient pb.LaptopServiceClient, laptop *pb.Laptop) {
	laptopRequest := &pb.CreateLaptopRequest{Laptop: laptop}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := laptopClient.CreateLaptop(ctx, laptopRequest)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.AlreadyExists {
			// laptop AlreadyExists
			log.Printf("laptop already exists %s", res.Id)
		} else {
			log.Fatal("cannot create laptop: ", err)
		}
		return
	}
	log.Printf("Created laptop with Id: %s", res.Id)
}

func searchLaptop(laptopClient pb.LaptopServiceClient, filter *pb.Filter) {
	log.Print("search filter: ", filter)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req := &pb.SearchLaptopRequest{Filter: filter}
	stream, err := laptopClient.SearchLaptop(ctx, req)
	if err != nil {
		log.Fatal("cannot search laptop: ", err)
	}
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Fatal("cannot receive response: ", err)
		}
		laptop := res.GetLaptop()
		log.Print(" - found: ", laptop.GetId())
		log.Print("  + brand: ", laptop.GetBrand())
		log.Print("  + name: ", laptop.GetName())
		log.Print("  + cpu cores: ", laptop.GetCpu().GetNumberCores())
		log.Print("  + cpu min ghz: ", laptop.GetCpu().GetMinGhz())
		log.Print("  + cpu max ghz: ", laptop.GetCpu().GetMaxGhz())
		log.Print("  + ram: ", laptop.GetRam().GetValue(), laptop.GetRam().GetUnit())
		log.Print("  + price: ", laptop.GetPriceUsd())
	}
}

func rateLaptop(laptopClient pb.LaptopServiceClient, laptopIDs []string, scores []float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	stream, err := laptopClient.RateLaptop(ctx)
	if err != nil {
		return fmt.Errorf("cannot rate laptop: %v", err)
	}

	// listening from the server
	waitResponse := make(chan error)
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				waitResponse <- nil
				return
			}
			if err != nil {
				waitResponse <- fmt.Errorf("cannot receive stream response: %v", err)
				return
			}
			log.Print("received response: ", res)
		}
	}()

	// start sending
	for i, laptopID := range laptopIDs {
		req := &pb.RateLaptopRequest{LaptopId: laptopID, Score: scores[i]}
		err := stream.Send(req)
		if err != nil {
			return fmt.Errorf("cannot send stream request: %v - %v", err, stream.RecvMsg(nil))
		}
		log.Print("sent request: ", req)
	}
	err = stream.CloseSend()
	if err != nil {
		return fmt.Errorf("cannot close send: %v", err)
	}
	err = <-waitResponse
	return err
}

func uploadImage(laptopClient pb.LaptopServiceClient, laptopID string, imagePath string) {
	file, err := os.Open(imagePath)
	if err != nil {
		log.Fatal("cannot open image file: ", err)
	}
	defer file.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	stream, err := laptopClient.UploadImage(ctx)
	if err != nil {
		log.Fatal("cannot upload image: ", err)
	}
	req := &pb.UploadImageRequest{
		Data: &pb.UploadImageRequest_Info{
			Info: &pb.ImageInfo{
				LaptopId:  laptopID,
				ImageType: filepath.Ext(imagePath),
			},
		},
	}
	err = stream.Send(req)
	if err != nil {
		log.Fatal("cannot send image info: ", err)
	}

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)
	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("cannot read chunk to buffer: ", err)
		}
		req := &pb.UploadImageRequest{
			Data: &pb.UploadImageRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}
		err = stream.Send(req)
		if err != nil {
			log.Fatal("cannot send chunk data to server: ", err)
		}
	}
	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("cannot receive response: ", err)
	}
	log.Printf("image uploaded with ids: %s, size: %d", res.GetId(), res.GetSize())
}

func testCreateLaptop(laptopClient pb.LaptopServiceClient) {
	createLaptop(laptopClient, sample.NewLaptop())
}

func testSearchLaptop(laptopClient pb.LaptopServiceClient) {
	for i := 0; i < 10; i++ {
		createLaptop(laptopClient, sample.NewLaptop())
	}

	filter := &pb.Filter{
		MaxPriceUsd: 3000,
		MinCpuCores: 4,
		MinCpuGhz:   2.5,
		MinRam:      &pb.Memory{Value: 8, Unit: pb.Memory_GIGABYTE},
	}
	searchLaptop(laptopClient, filter)
}

func testUploadImage(laptopClient pb.LaptopServiceClient) {
	laptop := sample.NewLaptop()
	createLaptop(laptopClient, laptop)
	uploadImage(laptopClient, laptop.GetId(), "tmp/laptop.jpg")
}

func testRateLaptop(laptopClient pb.LaptopServiceClient) {
	n := 3
	laptopIDs := make([]string, n)
	for i := 0; i < n; i++ {
		laptop := sample.NewLaptop()
		laptopIDs[i] = laptop.GetId()
		createLaptop(laptopClient, laptop)
	}
	scores := make([]float64, n)
	for {
		fmt.Print("rate laptop (y/n)? ")
		var ans string
		fmt.Scan(&ans)
		if strings.ToLower(ans) != "y" {
			break
		}
		for i := 0; i < n; i++ {
			scores[i] = sample.RandomLaptopScore()
		}
		err := rateLaptop(laptopClient, laptopIDs, scores)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func main() {
	fmt.Println("Hello Laptop Client!")
	serverAddress := flag.String("address", "", "the server address")
	flag.Parse()
	log.Printf("Dailing server %s", *serverAddress)

	conn, err := grpc.Dial(*serverAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatal("Cannot dail the server address.", err)
	}
	laptopClient := pb.NewLaptopServiceClient(conn)
	// testCreateLaptop(laptopClient)
	// testSearchLaptop(laptopClient)
	// testUploadImage(laptopClient)
	testRateLaptop(laptopClient)
}
