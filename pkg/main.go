package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"gopkg.in/src-d/go-git.v4"
)

// Node is the repository type
type Node struct {
	Repository struct {
		Name           string
		URL            string
		CreatedAt      string
		UpdatedAt      string
		StargazerCount int
		ForkCount      int
		Owner          struct {
			Login string
		}
		PrimaryLanguage struct {
			Name string
		}
		Watchers struct {
			TotalCount int
		}
		Releases struct {
			TotalCount int
		}
	}
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() (err error) {
	r := make(chan [][]string, 2)
	r <- ReadCSV("java")
	r <- ReadCSV("python")
	jvdata, pydata := <-r, <-r
	close(r)

	java := make(chan string, 100)
	quitJv := make(chan string)
	go ForEachLanguage(java, quitJv, jvdata)

	python := make(chan string, 100)
	quitPy := make(chan string)
	go ForEachLanguage(python, quitPy, pydata)

	// var j, p int
	for {
		select {
		case <-java:
			fmt.Println(<-java)
		case <-python:
			fmt.Println(<-python)
		default:
			qj, qp := <-quitJv, <-quitPy
			if qj == "quit" && qp == "quit" {
				return
			}
		}
	}
}

// ReadCSV ...
func ReadCSV(f string) (d [][]string) {
	df, err := os.Open("./data/csv/" + f + ".csv")
	if err != nil {
		panic(err)
	}
	defer df.Close()

	r := csv.NewReader(df)
	d, err = r.ReadAll()
	if err != nil {
		panic(err)
	}
	return
}

// ForEachLanguage ...
func ForEachLanguage(ch, quit chan string, lang [][]string) {
	for _, r := range lang {
		repo, err := CloneRepository(r)
		if err != nil {
			fmt.Printf("Error while clone: %e", err)
		}
		ch <- repo
	}
	quit <- "quit"
	close(ch)
}

// CloneRepository ...
func CloneRepository(r []string) (repo string, err error) {
	repo = fmt.Sprintf("%s-%s", r[0], r[1])
	url := fmt.Sprintf("%s.git", r[2])
	// Tempdir to clone the repository
	dir, err := ioutil.TempDir("./repositories", repo)
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir) // clean up

	// Clones the repository into the given dir, just as a normal git clone does
	_, err = git.PlainClone(dir, false, &git.CloneOptions{
		URL: url,
	})
	if err != nil {
		log.Fatal(err)
	}
	if err = WriteData(dir, repo); err != nil {
		fmt.Printf("Error while writing data for repository: %s | %s", repo, url)
	}
	return repo, err
}

// WriteData ...
func WriteData(dir string, repo string) (err error) {
	ch := make(chan error, 3)
	ch <- ExecCommand("csv", dir, repo)
	ch <- ExecCommand("tabular", dir, repo)
	ch <- ExecCommand("html", dir, repo)
	if err = <-ch; err != nil {
		fmt.Println(err)
	}
	close(ch)

	return
}

// ExecCommand ...
func ExecCommand(ext string, dir string, repo string) (err error) {
	cmd := exec.Command("scc", "-f", ext, "-o", "./../../"+ext+"/"+repo+"."+ext)
	cmd.Dir = dir
	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput

	if err = cmd.Run(); err != nil {
		fmt.Printf("Error while clone: %e", err)
	}
	fmt.Print(string(cmdOutput.Bytes()))

	return
}

// func writeTXT() {
// 	cmd := exec.Command("scc", "-f", "html", "-o", "./../output.html")
// 	cmd.Dir = dir
// 	if err != nil {
// 		os.Stderr.WriteString(err.Error())
// 	}

// 	cmdOutput := &bytes.Buffer{}
// 	cmd.Stdout = cmdOutput
// 	err = cmd.Run()
// 	if err != nil {
// 		os.Stderr.WriteString(err.Error())
// 	}
// 	fmt.Print(string(cmdOutput.Bytes()))
// }

// func writeHTML() {
// 	cmd := exec.Command("scc", "-f", "html", "-o", "./../output.html")
// 	cmd.Dir = dir
// 	if err != nil {
// 		os.Stderr.WriteString(err.Error())
// 	}

// 	cmdOutput := &bytes.Buffer{}
// 	cmd.Stdout = cmdOutput
// 	err = cmd.Run()
// 	if err != nil {
// 		os.Stderr.WriteString(err.Error())
// 	}
// 	fmt.Print(string(cmdOutput.Bytes()))
// }

// func writeCSV() {
// 	cmd := exec.Command("scc", "-f", "html", "-o", "./../output.html")
// 	cmd.Dir = dir
// 	if err != nil {
// 		os.Stderr.WriteString(err.Error())
// 	}

// 	cmdOutput := &bytes.Buffer{}
// 	cmd.Stdout = cmdOutput
// 	err = cmd.Run()
// 	if err != nil {
// 		os.Stderr.WriteString(err.Error())
// 	}
// 	fmt.Print(string(cmdOutput.Bytes()))
// }
