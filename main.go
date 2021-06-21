// get the file
// unzip the file
// execute npm install
// execute build command
// use terraform to create s3 bucket with random name
// take build file and upload to s3
// redirect a proxy with that domain name to the proxy
// delete the file
// any error ocurrs delete the file

package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	fileName    string
	fullURLFile string
)

func main() {

	fullURLFile = "https://github.com/LizzyKate/Todo/archive/refs/heads/master.zip"

	// Build fileName from fullPath
	fileURL, err := url.Parse(fullURLFile)
	if err != nil {
		log.Fatal(err)
	}
	path := fileURL.Path
	segments := strings.Split(path, "/")
	fileName = segments[len(segments)-1]

	// Create blank file
	file, err := os.Create(fileName)
	if err != nil {
		log.Fatal(err)
	}
	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
	// Put content on file
	resp, err := client.Get(fullURLFile)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	size, err := io.Copy(file, resp.Body)

	defer file.Close()

	fmt.Printf("Downloaded a file %s with size %d", fileName, size)

	Unzip(fileName, "./unzipped")

	exec.Command("rm", "-rf", fileName).Output()

	files, _err := exec.Command("ls", "./unzipped").Output()

	if _err != nil {
		fmt.Println("failed to list files in unzipped folder")
	}

	unzipped_name := strings.Split(string(files), "\n")

	os.Rename("unzipped/"+unzipped_name[0], "unzipped/master")

	// build, builderr := exec.Command("npm", "run", "build", "./unzipped/Todo-master/").Output()
	build, builderr := exec.Command("/bin/sh", "./list.sh").Output()

	if builderr != nil {
		fmt.Printf(builderr.Error())
	}

	fmt.Println(string(build))

	// take the zip file create via terraform and deploy
}

func Unzip(src string, dest string) ([]string, error) {

	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}
