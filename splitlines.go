package main

import "log"
import "os"

//import "io"
import "bufio"
import "strconv"
import "fmt"
import "gomiscutils"
import "path/filepath"

var err error

func main() {
	if len(os.Args) == 1 || os.Args[1] == "-h" || os.Args[1] == "--help" {
		os.Stderr.WriteString("Usage: splitlines NUM FILE\n")
		os.Exit(0)
	}
	chunkSize, err := strconv.ParseUint(os.Args[1], 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	filer, err := os.Open(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}
	defer filer.Close()

	totalLines := gomiscutils.Countlines(filer)
	//println("totalLines ",totalLines)

	if chunkSize >= totalLines {
		return
	}

	chunks := totalLines / chunkSize
	if totalLines%chunkSize > 0 {
		chunks++
	}

	namePaddingLength := len(strconv.FormatUint(chunks, 10))

	basename, ext := gomiscutils.SplitExt(os.Args[2])
	//fmt.Println(basename,"--",ext)
	dirname := filepath.Dir(os.Args[2])

	fReader := bufio.NewReader(filer)
	var lines chan string = gomiscutils.Readline(fReader)

	var lnum uint64
	var suffixNum uint32
	var lineChan chan string = make(chan string)
	defer close(lineChan)
	var newFileSuffix chan uint32 = make(chan uint32)
	defer close(newFileSuffix)
	var quit chan bool = make(chan bool)
	defer close(quit)
	var fileClose chan bool = make(chan bool)
	defer close(fileClose)
	var wfiler *os.File
	var filename string
	go func() {
		for {
			select {
			case suffixNum := <-newFileSuffix:
				var suffix string = fmt.Sprintf("_%0*d", namePaddingLength, suffixNum)
				filename = fmt.Sprintf("%s%s%s", basename, suffix, ext)
				wfiler, err = os.Create(filepath.Join(dirname, filename))
				if err != nil {
					log.Fatal(err)
				}
				defer wfiler.Close()
			case l := <-lineChan:
				wfiler.WriteString(l)
			case <-fileClose:
				wfiler.Close()
			case <-quit:
				return
			}
		}
	}()
	for l := range lines {
		lnum++
		if lnum%chunkSize == 1 || chunkSize == 1 {
			suffixNum++
			newFileSuffix <- suffixNum
		}
		lineChan <- l
		if lnum%chunkSize == 0 {
			fileClose <- true
		}
	}
	quit <- true
}
