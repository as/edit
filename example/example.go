// ssam
package main

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"os"
	"strings"

	"github.com/as/edit"
	"github.com/as/text"
)

func init() {
	log.SetFlags(0)
	log.SetPrefix("example: ")
}

func no(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

type Writer struct {
	text.Editor
}

func (w *Writer) Write(p []byte) (n int, err error) {
	return w.Editor.Insert(p, w.Len()), nil
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalln("usage: echo hello | example ,a,world,")
	}
	in, out := bufio.NewReader(os.Stdin), bufio.NewWriter(os.Stdout)

	ed, err := text.Open(text.NewBuffer())
	no(err)

	cmd, err := edit.Compile(strings.Join(os.Args[1:], " "))
	no(err)

	_, err = io.Copy(&Writer{ed}, in)
	no(err)

	cmd.Run(ed)
	_, err = io.Copy(out, bytes.NewReader(ed.Bytes()))
	no(err)

	out.Flush()
}
