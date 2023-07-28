package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/atotto/clipboard"
	"github.com/sashabaranov/go-openai"
)

func (agent *Agent) console() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	_, err := fmt.Println("Welcome")
	if err != nil {
		fmt.Println(err)
	} else {
		for {
			select {
			case <-interrupt:
				fmt.Println("Enter 'q' or 'quit' to exit!")
				continue
			default:
				fmt.Print("\nUser:\n")

				input := gettextinput()

				text := agent.process_text(input)

				if text == "" {
					continue
				}

				query := openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleUser,
					Content: text,
				}

				agent.req.Messages = append(agent.req.Messages, query)

				for {
					// retries response until it works
					response, err := agent.getresponse()
					if err != nil {
						fmt.Println(err)
						continue
					}

					estcost := (float64(agent.tokencount) / 1000) * callcost

					fmt.Println("\nAssistant:")
					fmt.Println(response.Message.Content)

					fmt.Println("\nTokencount: ", agent.tokencount, " Est. Cost: ", estcost)
					break
				}
			}
		}
	}
}

func (agent *Agent) process_text(text string) string {
	switch text {
	case "q", "quit":
		fmt.Println("\nQuitting...")
		os.Exit(0)
	case "del", "delete", "!":
		agent.setprompt()
		fmt.Println("\nChat cleared!")
		return ""
	case "reset":
		agent.reset()
		fmt.Println("\nChat reset!")
		return ""
	case "paste":
		text, err := clipboard.ReadAll()
		if err != nil {
			fmt.Println(err)
			return ""
		}
		fmt.Println("\nPasted text!")
		return text
	case "copy":
		response := agent.req.Messages[len(agent.req.Messages)].Content
		err := clipboard.WriteAll(response)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("\nCopied text!")
		return ""
	case "@", "sel", "select":
		agent.printnumberlines()
		fmt.Println("\nWhich lines would you like to delete?")
		editchoice := gettextinput()
		if editchoice == "" {
			return ""
		}
		agent.deletelines(editchoice)
		agent.printnumberlines()
		fmt.Println("Lines deleted!")
		return ""
	case "save":
		_, err := agent.save()
		if err != nil {
			fmt.Println(err)
		}
		return ""
	case "load":
		_, err := getsavefilelist()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("What file would you like to load?")
		filename := gettextinput()
		if filename == "" {
			return ""
		}
		err = agent.load(filename)
		if err != nil {
			fmt.Println(err)
		}
		return ""
	case "prompt":
		fmt.Println("\nEnter new prompt:")
		input := gettextinput()
		if input == "paste" {
			text, err := clipboard.ReadAll()
			if err != nil {
				fmt.Println(err)
				return ""
			}
			agent.prompt.Parameters = text
		} else {
			text := input
			agent.prompt.Parameters = text

		}
		agent.setprompt()
		fmt.Println("\nPrompt edited!")
		return ""
	case "help":
		fmt.Println("• Typing 'copy' will copy the last output from the bot\n• Typing 'paste' will paste your clipboard as a query - this way you can craft prompts in a text editor for multi-line queries\n• 'prompt' will let you enter in a new prompt ('paste' command works here)\n• 'save' will save the chat into a json file with the filename YYYYMMDDHHMM.txt\n• 'load <filename>' will load files\n• '@', 'sel', or 'select' will allow you to select lines to delete (handy if the chat is getting a bit long and you want to save on costs)\n• '!', 'del', or 'delete' will clear the chat log and start fresh\n'q' or 'quit' will quit the program")
		return ""
	default:
		// Nothing will happen
	}
	return text
}

func (agent *Agent) printnumberlines() {
	for i, msg := range agent.req.Messages {
		if msg.Role == openai.ChatMessageRoleUser {
			fmt.Printf("%d. User: %s\n", i, msg.Content)
		} else if msg.Role == openai.ChatMessageRoleAssistant {
			fmt.Printf("%d. Assistant: %s\n", i, msg.Content)
		}
	}
}
