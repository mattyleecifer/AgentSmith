package main

import (
	"fmt"
)

func main() {
	agent := newAgent()

	if serverFlag {
		fmt.Println("Running as server...")
		go agent.console()
		agent.gui()
	} else if consoleFlag {
		fmt.Println("Console only mode.")
		agent.console()
	} else {
		autofunction = true
		response, err := agent.getresponse()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(response.Message.Content)
		if savechatName != "" {
			agent.save(savechatName)
		}
	}
}
