package main

import (
	"bufio"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func HasHtml(headers map[string][]string) bool {
	xurl, ok := headers["X-Url"]
	if !ok || len(xurl) <= 0 {
		return false
	}
	return strings.HasPrefix(xurl[0], "https://twitter.com")
}

func typeHeaders(h map[string][]string, w io.Writer) {
	for key, vals := range h {
		for _, val := range vals {
			fmt.Fprintf(w, "%s: %s\n", key, val)
		}
	}
}

func html2text(out io.Writer, in io.Reader) {
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Fprintln(os.Stderr, line)
		line = strings.TrimSpace(line)
		line = strings.Replace(line, "<div><br></div>", "\n", -1)
		line = strings.Replace(line, "<div>", "", -1)
		line = strings.Replace(line, "</div>", "\n", -1)
		line = html.UnescapeString(line)
		io.WriteString(out, line)
	}
}

func text2html(out io.Writer, in io.Reader) {
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		line = html.EscapeString(line)
		if line == "" {
			line = "<br>"
		}
		fmt.Fprintf(out, "<div>%s</div>", line)
	}
}

func handler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(os.Stderr, "From: %s\n", req.RemoteAddr)
	typeHeaders(req.Header, os.Stderr)
	tmpfd, tmpfdErr := ioutil.TempFile("", "editsrv")
	if tmpfdErr != nil {
		fmt.Fprintln(os.Stderr, tmpfdErr)
		return
	}
	io.WriteString(tmpfd, "\xEF\xBB\xBF")
	tmpName := tmpfd.Name()
	defer os.Remove(tmpName)

	hasHtml := HasHtml(req.Header)
	if hasHtml {
		html2text(tmpfd, req.Body)
	} else {
		io.CopyN(tmpfd, req.Body, req.ContentLength)
	}
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
	if hasHtml {
		text2html(w, tmpfd)
	} else {
		_, copyErr2 := io.Copy(w, tmpfd)
		if copyErr2 != nil {
			fmt.Fprintln(os.Stderr, copyErr2)
		}
	}
	tmpfd.Close()
	fmt.Fprintln(os.Stderr, "Done")
}

func main() {
	fmt.Println("Any Editor Server for chrome-extension 'Edit with Emacs'")
	http.HandleFunc("/edit", handler)
	http.ListenAndServe(":9292", nil)
}
