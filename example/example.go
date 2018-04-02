// ssam
package main

import (
	"bufio"
	"bytes"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"runtime/pprof"
	"strings"

	"github.com/as/edit"
	"github.com/as/event"
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

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func trypprof() func() {
	if *cpuprofile == "" {
		return func() {}
	}
	f, err := os.Create(*cpuprofile)
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)
	return pprof.StopCPUProfile
}

func interp(buf text.Buffer, rec event.Record) {
	switch t := rec.(type) {
	case *event.Write:
		//fmt.Printf("%#v\n", t)
		buf.(io.WriterAt).WriteAt(t.P, t.Q0)
	case *event.Insert:
		//fmt.Printf("%#v\n", t)
		buf.Insert(t.P, t.Q0)
	case *event.Delete:
		//fmt.Printf("%#v\n", t)
		buf.Delete(t.Q0, t.Q1)
	}
}

func main() {
	flag.Parse()
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	cmd := edit.MustCompile(strings.Join(flag.Args(), " "))
	data, err := ioutil.ReadAll(in)
	if err != nil {
		log.Fatalf("edit: %s", err)
	}

	buf, _ := text.Open(text.BufferFrom(data))
	if err = cmd.Run(buf); err != nil {
		log.Fatalln("edit: %s", err)
	}

	io.Copy(out, bytes.NewReader(buf.Bytes()))
	out.Flush()
}

/*
func main() {
	if len(os.Args) < 2 {
		log.Fatalln("usage: echo hello | example ,a,world,")
	}
	in, out := bufio.NewReader(os.Stdin), bufio.NewWriter(os.Stdout)

	buf := text.NewBuffer()
	ed, err := text.Open(buf)
	no(err)

	cmd, err := edit.Compile(strings.Join(os.Args[1:], " "))
	no(err)

	_, err = io.Copy(&Writer{ed}, in)
	no(err)

	hist := worm.NewLogger()
	cor := text.NewCOR(ed, hist)
	cmd.Run(cor)
	cor.Flush()
	for i := int64(0); i < int64(hist.Len()); i++{
		e, err := hist.ReadAt(i)
		interp(buf, e)
		if err != nil{
			break
		}
	}
	_, err = io.Copy(out, bytes.NewReader(ed.Bytes()))
	out.Flush()
	no(err)
}
*/
