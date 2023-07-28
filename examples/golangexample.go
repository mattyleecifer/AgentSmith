package main

import (
	"fmt"
)

func main() {
	// Initiate new agent
	// You can also start a new agent with agent := newAgent("key")
	// see getflags() in core.go for full description of different starting options
	agent := newAgent()

	// Get response from agent
	response, err := agent.getresponse()
	if err != nil {
		fmt.Println(err)
	}

	// Extract and print response content from agent
	fmt.Println(response.Message.Content)

	// save response if -save flag is set
	if savechatName != "" {
		agent.save(savechatName)
	}
}

func (agent *Agent) examples() {
	// a function is an openai functiondefinition
	//There are two ways to load a function:
	// from homeDir/Functions
	functiondef, err := loadfunction("filename.json")
	if err != nil {
		// handle errors - i will omit error handling from here
		fmt.Println(err)
	}
	agent.setfunction(functiondef)
	// from string
	agent.setFunctionFromString("functiondefinitionjson")
}
