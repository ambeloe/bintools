package main

import (
	_ "embed"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

//easier to write this way; an extra 2kb of ram usage shouldn't be an issue
//go:embed help.txt
var helpText string

func main() {
	if len(os.Args) < 2 {
		fmt.Println("no command specified")
		os.Exit(1)
	}
	cargs := os.Args[2:]
sw:
	switch os.Args[1] {
	case "--help": //for compatibility with common help syntax
		fallthrough
	case "?":
		fmt.Println(helpText)
	case "e":
		vargnum(cargs, 2, 3)

		f, err := ioutil.ReadFile(cargs[0])
		exitErr(err, "error reading infile")

		//parse start and end addresses
		start, end, err := parseRange(cargs[1])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		end++

		//validate start and end addresses
		if start >= end {
			fmt.Println("start not before end")
			os.Exit(1)
		} else if int(end) > len(f) {
			fmt.Println("endaddr out of range")
			os.Exit(1)
		}

		//generate filename
		var fn string
		if len(cargs) == 3 {
			fn = cargs[2]
		} else {
			//chunk_startaddr:endaddr.chunk
			fn = filepath.Base(cargs[0]) + "_" + cargs[1] + ".chunk"
		}

		err = os.WriteFile(fn, f[start:end], 0644)
		exitErr(err, "error writing to outfile")
	case "ez":
		vargnum(cargs, 2, 3)

		f, err := os.ReadFile(cargs[0]) //yes this is inefficient
		exitErr(err, "error opening file")

		s, e, err := parseRange(cargs[1])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if s < e && int(e) <= len(f) { //range is valid
			//remove tailing zeroes
			for f[e] == 0 && e > s {
				e--
			}
			if e != s {
				//remove leading zeroes
				for f[s] == 0 && s < e {
					s++
				}
			}
		} else {
			fmt.Println("range invalid")
			os.Exit(1)
		}

		os.Args[1] = "e"
		cargs[1] = strconv.FormatUint(s, 16) + ":" + strconv.FormatUint(e, 16) //update range
		//fmt.Println(cargs)
		goto sw
	case "c":
		var off int64
		switch len(cargs) {
		case 2:
			off = 0
		case 3:
			e, err := strconv.ParseUint(cargs[2], 16, 0)
			exitErr(err, "error parsing offset")
			off = int64(e)
		default:
			fmt.Println("invalid arguments")
			os.Exit(1)
		}

		//open file, get stats, seek to desired offset
		f, err := os.Open(cargs[0])
		exitErr(err, "error opening infile")
		stat, err := f.Stat()
		exitErr(err, "error getting file stats")
		_, err = f.Seek(off, 0) //relative to start
		exitErr(err, "error seeking")

		//parse hex chunk size and create chunk buffer
		c, err := strconv.ParseUint(cargs[1], 10, 0)
		exitErr(err, "error parsing chunk size")
		if c > math.MaxInt64 {
			fmt.Println("chunk too fat")
			os.Exit(1)
		}
		chunk := make([]byte, c)

		var n int
		for i := 0; off < stat.Size(); i++ {
			//fmt.Println("chunk ", i)
			n, err = f.ReadAt(chunk, off)
			if err != nil && err != io.EOF {
				fmt.Println("error reading chunk from infile,", err)
			}

			//named infile_i_startaddr:endaddr.chunk
			err = os.WriteFile(filepath.Base(f.Name())+"_"+strconv.Itoa(i)+"_"+strconv.FormatInt(off, 16)+":"+strconv.FormatInt(off+int64(c), 16)+".chunk", chunk, 0644)
			exitErr(err, "error writing chunk")
			off += int64(c)
			if n != int(c) {
				break
			}
		}
	default:
		fmt.Println("command not implemented:", os.Args[1])
		os.Exit(1)
	}
}

//parse range in format startaddr:endaddr
func parseRange(ran string) (start uint64, end uint64, err error) {
	els := strings.Split(ran, ":")
	if len(els) != 2 {
		return 0, 0, errors.New("invalid format of address range")
	}
	start, err = strconv.ParseUint(els[0], 16, 0)
	if err != nil {
		return 0, 0, errors.New("error parsing startaddr")
	}
	end, err = strconv.ParseUint(els[1], 16, 0)
	if err != nil {
		return 0, 0, errors.New("error parsing endaddr")
	}
	return
}

//check if correct number of args is passed
func vargnum(a []string, min int, max int) {
	if len(a) < min || len(a) > max {
		fmt.Println("invalid arguments")
		os.Exit(1)
	}
}

func exitErr(err error, message ...interface{}) {
	if err != nil {
		fmt.Println(message)
		os.Exit(1)
	}
}
