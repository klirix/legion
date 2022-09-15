package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/docker/distribution/uuid"
)

type LegionDeploymentConfig struct {
	Name   string            `json:"name"`
	Domain string            `json:"domain"`
	Env    map[string]string `json:"env"`
}

const uploadDir = `legion/uploads/`

func main() {

	// Hello world, the web server

	// ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	// defer cancel()

	// docker, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	// if err != nil {
	// 	panic(err)
	// }

	// services, err := docker.ServiceList(ctx, types.ServiceListOptions{})
	// if err != nil {
	// 	log.Print("Failed to fetch containers")
	// 	panic(err)
	// }

	// for _, service := range services {
	// 	log.Printf("%s %s", service.Spec.Name, service.ID)
	// }

	http.HandleFunc("/hello", func(w http.ResponseWriter, req *http.Request) {
		uuid := uuid.Generate().String()
		bytes, err := io.ReadAll(req.Body)
		print(len(bytes))
		if err != nil {
			fmt.Fprintf(w, "Failed to read body")
			return
		}
		os.MkdirAll(uploadDir, fs.ModePerm)
		uploadFile := fmt.Sprintf("%s/%s.zip", uploadDir, uuid)
		err = ioutil.WriteFile(uploadFile, bytes, fs.ModePerm)
		// defer os.Remove(uploadFile)
		if err != nil {
			fmt.Fprintf(w, "Failed to write to file")
			return
		}

		archive, err := zip.OpenReader(uploadFile)
		if err != nil {
			fmt.Fprintf(w, "failed to open zip")
			return
		}
		defer archive.Close()

		err = checkLegionManifest(archive)
		if err != nil {
			fmt.Fprintf(w, "legion manifest not found")
			return
		}

		fmt.Fprintf(w, "ok")
	})

	log.Println("Listing for requests at http://localhost:8000/hello")
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func checkLegionManifest(archive *zip.ReadCloser) error {
	for _, file := range archive.File {
		if file.Name == "legion.json" {
			return nil
		}
	}
	return errors.New("Legion manifest not found")
}
