package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
)

func handler(w http.ResponseWriter, req *http.Request) {
	tmpfd, tmpfdErr := ioutil.TempFile("", "editsrv")
	if tmpfdErr != nil {
		fmt.Fprintln(os.Stderr, tmpfdErr)
		return
	}
	tmpName := tmpfd.Name()
	defer os.Remove(tmpName)

	io.CopyN(tmpfd, req.Body, req.ContentLength)
	tmpfd.Close()

	var editorName string
	if len(os.Args) >= 2 {
		editorName = os.Args[1]
	} else {
		editorName = "notepad.exe"
	}
	editorArgs := make([]string, 0, len(os.Args))
	for i := 2; i < len(os.Args); i++ {
		editorArgs = append(editorArgs, os.Args[i])
	}
	editorArgs = append(editorArgs, tmpName)
	cmd1 := exec.Command(editorName, editorArgs...)
	fmt.Fprintf(os.Stderr, "Call %s ", editorName)
	for _, arg1 := range editorArgs {
		fmt.Fprintf(os.Stderr, " %s", arg1)
	}
	fmt.Fprint(os.Stderr, "\n")
	if err := cmd1.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	tmpfd, tmpfdErr = os.Open(tmpName)
	if tmpfdErr != nil {
		fmt.Fprintln(os.Stderr, tmpfdErr)
		tmpfd.Close()
		return
	}
	fmt.Fprintf(os.Stderr, "Send '%s' to Chrome\n", tmpName)
	_, copyErr2 := io.Copy(w, tmpfd)
	if copyErr2 != nil {
		fmt.Fprintln(os.Stderr, copyErr2)
	}
	tmpfd.Close()
	fmt.Fprintln(os.Stderr, "Done")
}

func main() {
	fmt.Println("Any Editor Server for chrome-extension 'Edit with Emacs'")
	http.HandleFunc("/edit", handler)
	http.ListenAndServe(":9292", nil)
}
