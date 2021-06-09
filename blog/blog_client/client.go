package main

import (
	"context"
	"fmt"
	"grpc_go_boilerplate/blog/blogpb"
	"io"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	fmt.Println("Blog Client")

	tls := true

	opts := grpc.WithInsecure()

	if tls {
		// SSL Certificate
		certFile := "ssl/ca.crt"
		creds, sslErr := credentials.NewClientTLSFromFile(certFile, "")
		if sslErr != nil {
			log.Fatalf("Error with CA Trust certification: %v", sslErr)
		}
		opts = grpc.WithTransportCredentials(creds)
	} else {

	}

	conn, err := grpc.Dial("localhost:50051", opts)

	// Insecure
	// conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect grpc: %v", err)
	}
	defer conn.Close()

	c := blogpb.NewBlogServiceClient(conn)
	// fmt.Printf("Created Client %f\n", c)

	blog := &blogpb.Blog{
		AuthorId: "boms",
		Title:    "My first BLOG",
		Content:  "Contents of my first blog",
	}

	// why ?????
	// blog := blogpb.Blog{
	// 	AuthorId: "elbum",
	// 	Title:    "My first BLOG",
	// 	Content:  "Contents of my first blog",
	// }

	createBlogRes, err := c.CreateBlog(context.Background(), &blogpb.CreateBlogRequest{Blog: blog})
	if err != nil {
		log.Fatalf("Unexpected error :%v", err)
	}

	fmt.Printf("Blog has been created :%v\n", createBlogRes)
	blogID := createBlogRes.GetBlog().GetId()
	fmt.Printf("Blog id ===== : %v\n", blogID)

	// Read Blog
	_, err2 := c.ReadBlog(context.Background(), &blogpb.ReadBlogRequest{BlogId: "12345"})
	if err2 != nil {
		fmt.Printf("Error happened :%v\n\n", err2)
	}

	readBlogReq := &blogpb.ReadBlogRequest{BlogId: blogID}
	readBlogRes, readBlogErr := c.ReadBlog(context.Background(), readBlogReq)
	if readBlogErr != nil {
		fmt.Printf("Error happened :%v\n", err2)
	}
	fmt.Printf("Blog was read  :%v\n", readBlogRes)

	// Update Blog
	newblog := &blogpb.Blog{
		Id:       blogID,
		AuthorId: "Bens",
		Title:    "Edited My first BLOG",
		Content:  "Edited Contents of my first blog , some addition",
	}

	updateRes, updateErr := c.UpdateBlog(context.Background(), &blogpb.UpdateBlogRequest{Blog: newblog})
	if updateErr != nil {
		fmt.Printf("Error happened :%v\n", updateErr)
	}
	fmt.Printf("Blog was Updated  :%v\n", updateRes)

	// delete blog
	deleteRes, deleteErr := c.DeleteBlog(context.Background(), &blogpb.DeleteBlogRequest{BlogId: blogID})
	if deleteErr != nil {
		fmt.Printf("Error happened :%v\n", deleteErr)
	}
	fmt.Printf("Blog was Deleted  :%v\n", deleteRes)

	// list blog (server stream)
	stream, errs := c.ListBlog(context.Background(), &blogpb.ListBlogRequest{})
	if errs != nil {
		log.Fatalf("error while calling liststream RPC : %v", err)
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			// end of resStream
			break
		}
		if err != nil {
			log.Fatalf("error while getting stream :%v", err)
		}
		log.Printf("Response from BlogServer: %v", res.GetBlog())
	}

}
