package main

import (
	"context"
	"io"
	"log"
	"net/http"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func main() {
	// Hello world, the web server
	print("Starting up!!!\n")

	ctx := context.Background()

	docker, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	services, err := docker.ServiceList(ctx, types.ServiceListOptions{})
	if err != nil {
		log.Print("Failed to fetch containers")
		panic(err)
	}

	for _, service := range services {
		log.Printf("%s %s", service.Spec.Name, service.ID)
	}

	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "Hello, world!\n")
	}

	http.HandleFunc("/hello", helloHandler)
	log.Println("Listing for requests at http://localhost:8000/hello")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
