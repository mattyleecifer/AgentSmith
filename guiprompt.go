package main

// create, edit, and make prompt calls - this will allow users to make commandline or api promptcalls.

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (agent *Agent) hprompt(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		var data struct {
			Name         string
			Description  string
			Parameters   string
			Savedprompts []string
		}

		data.Name = agent.prompt.Name
		data.Description = agent.prompt.Description
		data.Parameters = agent.prompt.Parameters
		data.Savedprompts, _ = getsavepromptlist()

		render(w, hpromptspage, data)
	}

	if r.Method == http.MethodPost {
		newprompt := promptDefinition{
			Name:        r.FormValue("promptname"),
			Description: r.FormValue("promptdescription"),
			Parameters:  r.FormValue("edittext"),
		}

		agent.prompt = newprompt
		agent.setprompt()

		r.Method = http.MethodGet
		agent.hchat(w, r)
	}
}

func (agent *Agent) hpromptfiles(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimPrefix(r.URL.Path, "/prompt/files/")

	if r.Method == http.MethodGet {
		var data struct {
			Name         string
			Description  string
			Parameters   string
			Savedprompts []string
		}

		prompt := promptDefinition{}

		loaddata, err := agent.loadfile("Prompts", query)
		if err != nil {
			fmt.Println(err)
		}

		_ = json.Unmarshal(loaddata, &prompt)

		data.Name = prompt.Name
		data.Description = prompt.Description
		data.Parameters = prompt.Parameters
		data.Savedprompts, _ = getsavepromptlist()

		render(w, hpromptspage, data)
	}

	if r.Method == http.MethodPost {
		newprompt := promptDefinition{
			Name:        r.FormValue("promptname"),
			Description: r.FormValue("promptdescription"),
			Parameters:  r.FormValue("edittext"),
		}

		agent.savefile(newprompt, "Prompts", newprompt.Name)

		r.Method = http.MethodGet
		agent.hprompt(w, r)
	}

	if r.Method == http.MethodDelete {
		deletefile("Prompts", query)

		r.Method = http.MethodGet
		agent.hprompt(w, r)
	}
}

func getsavepromptlist() ([]string, error) {
	// Create a directory for your app
	savepath := filepath.Join(homeDir, "Prompts")
	files, err := os.ReadDir(savepath)
	if err != nil {
		return nil, err
	}

	var res []string

	fmt.Println("\nFiles:")

	for _, file := range files {
		filename := strings.ReplaceAll(file.Name(), ".json", "")
		res = append(res, filename)
		fmt.Println(file.Name())
	}

	return res, nil
}
