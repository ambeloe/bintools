package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("no command specified")
		os.Exit(1)
	}
	cargs := os.Args[2:]
sw:
	switch os.Args[1] {
	case "?":
		fmt.Println("e: copy out chunk of file (both inclusive)\n" +
			"	Usage:\n" +
			"	|- infile startaddr:endaddr outfile\n" +
			"	|_ infile startaddr:endaddr")
		fmt.Println("ez: e but with leading and trailing zeroes cut off\n" +
			"	Usage: same as e")
	case "e":
		vargnum(cargs, 2, 3)

		f, err := ioutil.ReadFile(cargs[0])
		errExit(err, "error reading infile")

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
			fn = cargs[0] + "_" + cargs[1] + ".chunk"
		}

		err = ioutil.WriteFile(fn, f[start:end], 0644)
		errExit(err, "error writing to outfile")
	case "ez":
		vargnum(cargs, 2, 3)

		f, err := ioutil.ReadFile(cargs[0])
		errExit(err, "error opening file")

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
	default:
		fmt.Println("command not implemented:", os.Args[1])
		os.Exit(1)
	}
}

//parse range in format
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

func errExit(err error, message string) {
	if err != nil {
		fmt.Println(message)
		os.Exit(1)
	}
}
