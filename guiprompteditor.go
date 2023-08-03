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

func (agent *Agent) handlersprompteditor() {
	http.HandleFunc("/prompteditpage", RequireAuth(agent.hprompteditpage))
	http.HandleFunc("/promptdelete", RequireAuth(agent.hpromptdelete))
	http.HandleFunc("/promptload", RequireAuth(hpromptload))
	http.HandleFunc("/promptset", RequireAuth(agent.hpromptset))
	http.HandleFunc("/promptsave", RequireAuth(agent.hpromptsave))
}

func (agent *Agent) hprompteditpage(w http.ResponseWriter, r *http.Request) {
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
	deleteprompt(promptname)
	agent.hprompteditpage(w, r)
}

func hpromptload(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Name         string
		Description  string
		Parameters   string
		SavedPrompts template.HTML
	}

	promptname := r.FormValue("promptname")
	promptname += ".json"

	prompt, err := loadprompt(promptname)
	if err != nil {
		fmt.Println(err)
	}

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

	saveprompt(&newprompt)
	agent.hprompteditpage(w, r)
}

func saveprompt(f *promptDefinition) (string, error) {
	// saves to disk
	appDir := filepath.Join(homeDir, "Prompts")
	err := os.MkdirAll(appDir, os.ModePerm)
	if err != nil {
		fmt.Println("Failed to create app directory:", err)
		return "", err
	}

	jsonData, err := json.Marshal(f)
	if err != nil {
		return "", err
	}

	filename := f.Name + ".json"

	savepath := filepath.Join(appDir, filename)

	file, err := os.OpenFile(savepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = file.Write(jsonData)
	if err != nil {
		return "", err
	}

	fmt.Println("\nFile saved: ", savepath)

	return filename, nil
}

func deleteprompt(filename string) error {
	// removes from disk
	filepath := filepath.Join(homeDir, "Prompts", filename)

	err := os.Remove(filepath)
	if err != nil {
		fmt.Println("Error deleting file:", err)
		return err
	}

	fmt.Println("File deleted successfully.")

	return nil
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
			savedprompts += "<tr><td>" + name + "</td><td><form hx-post='/promptload' hx-target='#main-content' hx-swap='outerHTML'><button class='btn' name='promptname' value='" + name + "'>Load</button></form></td><td><form hx-post='/promptdelete' hx-target='#main-content' hx-swap='outerHTML' hx-confirm='Are you sure?'><button class='btn' name='promptname' value='" + name + "'>Delete</button></form></td></tr>"
		}
		savedprompts += `</table>`
	}

	tsavedprompts := template.HTML(savedprompts)
	return tsavedprompts
}
