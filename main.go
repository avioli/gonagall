package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/nfnt/resize"
	"html/template"
	"image/jpeg"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

const gonagallConfigFile = "gonagallconfig.json"

var gonagallConfig struct {
	BaseDir string
	TempDir string
}

func writeConfig() error {
	j, err := json.MarshalIndent(&gonagallConfig, "", "\t")
	if err != nil {
		fmt.Println("Error in writeConfig:", err)
		return err
	}
	outFile, err := os.Create(gonagallConfigFile)
	b := bytes.NewBuffer(j)
	if _, err := b.WriteTo(outFile); err != nil {
		fmt.Println("Error in writeConfig:", err)
		return err
	}
	return nil
}

func readConfig() error {
	s, _ := os.Getwd()
	fmt.Println("pwd", s)
	gonagallConfig.BaseDir = "."
	gonagallConfig.TempDir = "/tmp"
	inFile, err := os.Open(gonagallConfigFile)
	if err != nil {
		fmt.Println("Error in readConfig:", err)
		return writeConfig()
	}
	dec := json.NewDecoder(inFile)
	err = dec.Decode(&gonagallConfig)
	if err != nil {
		fmt.Println("Error in readConfig:", err)
		return err
	}
	return writeConfig()
}

type dummy struct {
	Subs []string
	Jpgs []string
}

func BrowseDirectory(w http.ResponseWriter, r *http.Request) {
	l := len("/")
	upath := r.URL.Path[l:]
	t, err := template.ParseFiles("template.browse.html")
	if err != nil {
		fmt.Fprintln(w, "1", err)
		return
	}
	entries, err := ioutil.ReadDir(gonagallConfig.BaseDir + "/" + upath)
	if err != nil {
		fmt.Fprintln(w, "2", err, gonagallConfig.BaseDir+"/"+upath)
		return
	}
	var d dummy
	for _, r := range entries {
		if r.IsDir() && !strings.HasPrefix(r.Name(), ".") {
			d.Subs = append(d.Subs, upath+"/"+r.Name())
		} else if !r.IsDir() && strings.HasSuffix(r.Name(), ".jpg") {
			d.Jpgs = append(d.Jpgs, upath+"/"+r.Name())
		}
	}
	err = t.Execute(w, d)
	if err != nil {
		fmt.Fprintln(w, "3", err)
		return
	}
	//http.ServeFile(w, r, gonagallConfig.BaseDir+upath)
}

func ServeThumb(w http.ResponseWriter, r *http.Request) {
	l := len("/thumb/")
	s := gonagallConfig.BaseDir + "/" + r.URL.Path[l:]
	fmt.Println("Trying to serve ", s)
	fImg1, _ := os.Open(s)
	img1, _ := jpeg.Decode(fImg1)
	fImg1.Close()
	fmt.Println("decoded", r.URL.Path)

	img2 := resize.Resize(100, 0, img1, resize.NearestNeighbor)

	jpeg.Encode(w, img2, nil)
	fmt.Println("encoded", r.URL.Path, r.URL.Path[l:])
}

func ServeSmall(w http.ResponseWriter, r *http.Request) {
	l := len("/view/")
	s := gonagallConfig.BaseDir + "/" + r.URL.Path[l:]
	fmt.Println("Trying to serve ", s)
	fImg1, _ := os.Open(s)
	img1, _ := jpeg.Decode(fImg1)
	fImg1.Close()
	fmt.Println("decoded", r.URL.Path)

	img2 := resize.Resize(480, 0, img1, resize.NearestNeighbor)

	jpeg.Encode(w, img2, nil)
	fmt.Println("encoded", r.URL.Path, r.URL.Path[l:])
}

func main() {
	readConfig()
	fmt.Println(gonagallConfig)
	http.HandleFunc("/", BrowseDirectory)
	http.HandleFunc("/thumb/", ServeThumb)
	http.HandleFunc("/view/", ServeSmall)
	//	http.HandleFunc("/original/", ServeOriginal)
	http.ListenAndServe(":8781", nil)
}
