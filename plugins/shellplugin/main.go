package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func main() {
	req := getrequest()

	// fmt.Println("req", req)

	switch req["command"] {
	case "python":
		code := strings.ReplaceAll(req["args"], "\\n", ";")
		python(code)
	case "shell":
		shell(req["args"])
	}
	// fmt.Println("code", code)
}

func python(code string) {
	cmd := "python3"
	arg1 := "-c"
	arg2 := code
	cmdArgs := []string{arg1, arg2}
	// fmt.Println(cmdArgs)

	exec := exec.Command(cmd, cmdArgs...)

	output, err := exec.CombinedOutput()
	if err != nil {
		// log.Println(err)
		output = []byte(err.Error())
	}

	fmt.Println(string(output))
}

func shell(code string) {
	cmd := "bash"
	arg1 := "-c"
	arg2 := code
	cmdArgs := []string{arg1, arg2}
	// fmt.Println(cmdArgs)

	exec := exec.Command(cmd, cmdArgs...)

	output, err := exec.CombinedOutput()
	if err != nil {
		// log.Println(err)
		output = []byte(err.Error())
	}

	fmt.Println(string(output))
}
