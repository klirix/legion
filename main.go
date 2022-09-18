package main

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/docker/distribution/uuid"
)

type LegionDeploymentConfig struct {
	Name   string            `json:"name"`
	Domain string            `json:"domain"`
	Env    map[string]string `json:"env"`
}

const uploadDir = `legion/uploads/`

func main() {

	// cli, err := client.NewClientWithOpts(client.FromEnv)

	http.HandleFunc("/hello", func(w http.ResponseWriter, req *http.Request) {
		uploadFile, err := grabTempFile(req)
		if err != nil {
			fmt.Fprintf(w, "Failed to persist temp file, "+err.Error())
			return
		}
		defer os.Remove(uploadFile)
		archive, err := zip.OpenReader(uploadFile)
		if err != nil {
			fmt.Fprintf(w, "failed to open zip")
			return
		}

		manifest, err := checkLegionManifest(archive)
		if err != nil {
			fmt.Fprint(w, err.Error())
			return
		}

		for _, file := range archive.File {
			os.MkdirAll("legion/builds", fs.ModePerm)
			ofile, erro := os.Create(path.Join("legion/builds", file.Name))
			if erro != nil {
				fmt.Fprint(w, "Failed to open files: "+erro.Error())
				return
			}
			ifile, erri := file.Open()
			if erri != nil {
				fmt.Fprint(w, "Failed to open files: "+erri.Error())
				return
			}
			io.Copy(ofile, ifile)
		}

		fmt.Printf("%#v", manifest)

		// cli.ImageBuild(context.TODO())

		fmt.Fprintf(w, "ok")
	})

	log.Println("Listing for requests at http://localhost:8000/hello")
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func grabTempFile(req *http.Request) (uploadFile string, err error) {
	uuid := uuid.Generate().String()
	uploadFile = ""
	reader, err := req.MultipartReader()
	if err != nil {
		return
	}
	form, err := reader.ReadForm(1 << 32)
	if err != nil {
		return
	}

	os.MkdirAll(uploadDir, fs.ModePerm)
	if err != nil {
		return
	}
	fileIn, err := form.File["file"][0].Open()
	if err != nil {
		return
	}

	uploadFile = path.Join(uploadDir, uuid+".zip")
	fileOut, err := os.Create(uploadFile)
	if err != nil {
		return
	}

	_, err = io.Copy(fileOut, fileIn)
	if err != nil {
		return
	}

	return
}

func checkLegionManifest(archive *zip.ReadCloser) (LegionDeploymentConfig, error) {
	manifest := LegionDeploymentConfig{}
	for _, file := range archive.File {
		if file.Name != "legion.json" {
			continue
		}
		manifestFile, err := file.Open()
		if err != nil {
			return manifest, err
		}
		decoder := json.NewDecoder(manifestFile)
		decoder.Decode(&manifest)
		return manifest, nil
	}
	return manifest, errors.New("legion manifest not found")
}
