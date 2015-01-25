//Docket Client
//Author: Sivamani Varun
//Gopher Gala
package main

//push
//pull
//-h[ost]
//-p[ort]

import (
	"errors"
	"fmt"
	"github.com/alecthomas/kingpin"
	"github.com/fsouza/go-dockerclient"
	"os/exec"
	"regexp"
	"strings"
)

var (
	host = kingpin.Flag("host", "Set host of docket registry.").Short('h').Default("127.0.0.1").IP()
	port = kingpin.Flag("port", "Set port of docket registry.").Short('p').Default("9090").Int()

	push      = kingpin.Command("push", "Push to the docket registry.")
	pushImage = push.Arg("push", "Image to push.").Required().String()

	pull      = kingpin.Command("pull", "pull to the docket registry.")
	pullImage = pull.Arg("pull", "Image to pull.").Required().String()
)

func applyPush(image string) error {
	reg, err := regexp.Compile("[^A-Za-z0-9]+")
	if err != nil {
		return err
	}

	endpoint := "unix:///var/run/docker.sock"
	client, _ := docker.NewClient(endpoint)
	imgs, _ := client.ListImages(docker.ListImagesOptions{All: false})

	found := false
	tagId := ""
	filePath := ""

	for _, img := range imgs {
		tags := img.RepoTags
		for _, tag := range tags {
			if tag == image {
				found = true
				tagId = img.ID
				fmt.Println("Found image: ", image)
				fmt.Println("ID: ", img.ID)
				fmt.Println("RepoTags: ", img.RepoTags)
				fmt.Println("Created: ", img.Created)
				fmt.Println("Size: ", img.Size)
				fmt.Println("VirtualSize: ", img.VirtualSize)
				fmt.Println("ParentId: ", img.ParentID)
				safeImageName := reg.ReplaceAllString(image, "_")
				s := []string{"/tmp/", tagId, "_", safeImageName, ".tar"}
				filePath = strings.Join(s, "")
				break
			}
		}
	}
	if !found {
		return errors.New("Sorry the image could not be found.")
	}

	//Run export command
	//command invocation
	//run docker command save to tar ball in /tmp
	fmt.Println("Exporting image to tarball...")
	cmd := fmt.Sprintf("docker save %s > %s", image, filePath)
	output, err1 := exec.Command("sh", "-c", cmd).Output()
	if err1 != nil {
		return err1
	}

	fmt.Println("Successively exported tarball...")
	//make post request with contents of tarball to docket registry

	req, err2 := http.NewRequest("POST", (*postURL).String(), nil)
	if err2 != nil {
		return err2
	}
	if len(*postData) > 0 {
		for key, value := range *postData {
			req.Form.Set(key, value)
		}
	} else if postBinaryFile != nil {
		if headers.Get("Content-Type") != "" {
			headers.Set("Content-Type", "application/octet-stream")
		}
		req.Body = *postBinaryFile
	} else {
		return errors.New("--data or --data-binary must be provided to POST")
	}

	req.Header = *headers
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("HTTP request failed: %s", resp.Status)
	}
	_, err = io.Copy(os.Stdout, resp.Body)
	return err

	return nil
}

func applyPull(image string) error {
	fmt.Println("image = ", image)
	fmt.Println("host = ", host)
	fmt.Println("port = ", port)
	return nil
}

func main() {
	kingpin.CommandLine.Help = "Docket Client"

	switch kingpin.Parse() {
	case "push":
		kingpin.FatalIfError(applyPush(*pushImage), "Pushing of image failed")
	case "pull":
		kingpin.FatalIfError(applyPush((*pushImage)), "Pushing of image failed")
	}
}
