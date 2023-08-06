package main

// create, edit, and make prompt calls - this will allow users to make commandline or api promptcalls.

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (agent *Agent) hprompt(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Name         string
		Description  string
		Parameters   string
		SavedPrompts template.HTML
	}

	data.Name = agent.prompt.Name
	data.Description = agent.prompt.Description
	data.Parameters = agent.prompt.Parameters
	data.SavedPrompts = rendersavedprompts()

	render(w, heditprompt, data)
}

func (agent *Agent) hpromptdelete(w http.ResponseWriter, r *http.Request) {
	promptname := r.FormValue("promptname")
	promptname += ".json"
	deletefile("Prompts", promptname)
	agent.hprompt(w, r)
}

func (agent *Agent) hpromptload(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Name         string
		Description  string
		Parameters   string
		SavedPrompts template.HTML
	}

	promptname := r.FormValue("promptname")
	promptname += ".json"

	prompt := promptDefinition{}

	loaddata, err := agent.loadfile("Prompts", promptname)
	if err != nil {
		fmt.Println(err)
	}

	_ = json.Unmarshal(loaddata, &prompt)

	data.Name = prompt.Name
	data.Description = prompt.Description
	data.Parameters = prompt.Parameters

	data.SavedPrompts = rendersavedprompts()

	render(w, heditprompt, data)
}

func (agent *Agent) hpromptset(w http.ResponseWriter, r *http.Request) {
	newprompt := promptDefinition{
		Name:        r.FormValue("promptname"),
		Description: r.FormValue("promptdescription"),
		Parameters:  r.FormValue("edittext"),
	}
	agent.prompt = newprompt
	agent.setprompt()

	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (agent *Agent) hpromptsave(w http.ResponseWriter, r *http.Request) {
	newprompt := promptDefinition{
		Name:        r.FormValue("promptname"),
		Description: r.FormValue("promptdescription"),
		Parameters:  r.FormValue("edittext"),
	}

	agent.savefile(newprompt, "Prompts", newprompt.Name)
	agent.hprompt(w, r)
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
		res = append(res, file.Name())
		fmt.Println(file.Name())
	}

	return res, nil
}

func rendersavedprompts() template.HTML {
	allsavedprompts, err := getsavepromptlist()
	if err != nil {
		fmt.Println(err)
	}

	var savedprompts string
	if allsavedprompts != nil {
		allsavedprompts, err := getsavepromptlist()
		if err != nil {
			fmt.Println(err)
		}
		savedprompts += `<table style="display: flex;" id="centertext">`
		for i := 0; i < len(allsavedprompts); i++ {
			name := strings.ReplaceAll(allsavedprompts[i], ".json", "")
			savedprompts += "<tr><td>" + name + "</td><td><form hx-post='/prompt/load/' hx-target='#main-content' hx-swap='outerHTML'><button class='btn' name='promptname' value='" + name + "'>Load</button></form></td><td><form hx-post='/prompt/delete/' hx-target='#main-content' hx-swap='outerHTML' hx-confirm='Are you sure?'><button class='btn' name='promptname' value='" + name + "'>Delete</button></form></td></tr>"
		}
		savedprompts += `</table>`
	}

	tsavedprompts := template.HTML(savedprompts)
	return tsavedprompts
}
