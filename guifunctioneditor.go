package main

// create, edit, and make function calls

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

func (agent *Agent) handlersfunctioneditor() {
	http.HandleFunc("/functions", RequireAuth(agent.hfunctions))
	http.HandleFunc("/functiondelete", RequireAuth(agent.hfunctiondelete))
	http.HandleFunc("/functionremove", RequireAuth(agent.hfunctionremove))
	http.HandleFunc("/functioneditcurrent", RequireAuth(agent.hfunctioneditcurrent))
	http.HandleFunc("/functioneditadd", RequireAuth(agent.hfunctioneditadd))
	http.HandleFunc("/functionload", RequireAuth(agent.hfunctionload))
	http.HandleFunc("/functionset", RequireAuth(agent.hfunctionset))
	http.HandleFunc("/functionsave", RequireAuth(agent.hfunctionsave))
}

func (agent *Agent) hfunctions(w http.ResponseWriter, r *http.Request) {
	var data struct {
		CurrentFunctions template.HTML
		SavedFunctions   template.HTML
	}

	var currentfunctions string
	if agent.req.Functions != nil {
		currentfunctions += `<table style="display: flex;" id="centertext">`
		for i := 0; i < len(agent.req.Functions); i++ {
			name := agent.req.Functions[i].Name
			description := agent.req.Functions[i].Description
			currentfunctions += "<tr style='text-align: left;'><td>" + name + ":<br>" + description + "<br></td><td><form hx-get='/functioneditcurrent' hx-target='#main-content' hx-swap='outerHTML'><input type='hidden' name='functionname' value='" + name + "'><button class='btn'>Edit</button></form><br></td><td><form hx-get='/functionremove' hx-target='#main-content' hx-swap='outerHTML'><input type='hidden' name='functionname' value='" + name + "'><button class='btn'>Remove</button></form><br></td></tr>"
		}
		currentfunctions += `</table>`
	}
	allsavedfunctions, err := getsavefunctionlist()
	if err != nil {
		fmt.Println(err)
	}

	var savedfunctions string
	if allsavedfunctions != nil {
		allsavedfunctions, err := getsavefunctionlist()
		if err != nil {
			fmt.Println(err)
		}
		savedfunctions += `<table style="display: flex;" id="centertext">`
		for i := 0; i < len(allsavedfunctions); i++ {
			name := strings.ReplaceAll(allsavedfunctions[i], ".json", "")
			savedfunctions += "<tr><td style='text-align: left;'>" + name + "</td><td><form hx-post='/functionload' hx-target='#main-content' hx-swap='outerHTML'><input type='hidden' name='functionname' value='" + name + "'><button class='btn'>Load</button></form></td><td><form><input type='hidden' name='functionname' value='" + name + "'><button class='btn' hx-post='/functiondelete' hx-target='#main-content' hx-swap='outerHTML' hx-confirm='Are you sure?'>Delete</button></form></td></tr>"
		}
		savedfunctions += `</table>`
	}
	tcurrentfunctions := template.HTML(currentfunctions)
	tsavedfunctions := template.HTML(savedfunctions)
	data.CurrentFunctions = tcurrentfunctions
	data.SavedFunctions = tsavedfunctions

	render(w, hfunctionspage, data)
}

func (agent *Agent) hfunctionremove(w http.ResponseWriter, r *http.Request) {
	functionname := r.FormValue("functionname")
	agent.removefunction(functionname)
	agent.hfunctions(w, r)
}

func (agent *Agent) hfunctiondelete(w http.ResponseWriter, r *http.Request) {
	functionname := r.FormValue("functionname")
	functionname += ".json"
	deletefunction(functionname)
	agent.hfunctions(w, r)
}

func (agent *Agent) hfunctionload(w http.ResponseWriter, r *http.Request) {
	functionname := r.FormValue("functionname")
	functionname += ".json"
	newfunction, err := loadfunction(functionname)
	if err != nil {
		fmt.Println(err)
	}
	agent.setfunction(newfunction)
	agent.hfunctions(w, r)
}

func (agent *Agent) hfunctionset(w http.ResponseWriter, r *http.Request) {
	functionname := r.FormValue("functionname")
	functiondescription := r.FormValue("functiondescription")
	parameters := r.FormValue("edittext")

	var jsonData map[string]interface{}
	_ = json.Unmarshal([]byte(parameters), &jsonData)

	newfunction := openai.FunctionDefinition{
		Name:        functionname,
		Description: functiondescription,
		Parameters:  jsonData,
	}

	agent.setfunction(newfunction)
	agent.hfunctions(w, r)
}

func (agent *Agent) hfunctionedit(w http.ResponseWriter, r *http.Request, f openai.FunctionDefinition) {
	data := openai.FunctionDefinition{}

	functiondata, err := json.MarshalIndent(f.Parameters, "", "    ")
	if err != nil {
		fmt.Println("Error:", err)
	}
	data.Name = f.Name
	data.Description = f.Description
	data.Parameters = string(functiondata)

	render(w, heditfunction, data)
}

func (agent *Agent) hfunctioneditcurrent(w http.ResponseWriter, r *http.Request) {
	functionname := r.FormValue("functionname")

	for _, function := range agent.req.Functions {
		if functionname == function.Name {
			agent.hfunctionedit(w, r, function)
			continue
		}
	}
}

func (agent *Agent) hfunctioneditadd(w http.ResponseWriter, r *http.Request) {
	f := openai.FunctionDefinition{
		Name:        "New",
		Description: "New",
		Parameters: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"Variable1": {
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"Variable2": {
							Type:        jsonschema.String,
							Description: "Description of variable",
						},
					},
				},
				"Variable3": {
					Type: jsonschema.String,
					Enum: []string{"Choice1", "Choice2"},
				},
			},
			Required: []string{"Variable1", "Variable3"},
		},
	}
	agent.hfunctionedit(w, r, f)
}

func (agent *Agent) hfunctionsave(w http.ResponseWriter, r *http.Request) {
	newfunction := openai.FunctionDefinition{
		Name:        r.FormValue("functionname"),
		Description: r.FormValue("functiondescription"),
		Parameters:  map[string]interface{}{},
	}

	edittext := r.FormValue("edittext")
	edittext = strings.ReplaceAll(edittext, "\n", "")
	edittext = strings.ReplaceAll(edittext, "    ", "")

	var jsonData map[string]interface{}
	err := json.Unmarshal([]byte(edittext), &jsonData)
	if err != nil {
		fmt.Println("Error:", err)
	}

	newfunction.Parameters = jsonData

	savefunction(&newfunction)

	agent.hfunctions(w, r)
}

func savefunction(f *openai.FunctionDefinition) (string, error) {
	// saves to disk
	appDir := filepath.Join(homeDir, "Functions")
	err := os.MkdirAll(appDir, os.ModePerm)
	if err != nil {
		fmt.Println("Failed to create app directory:", err)
		return "", err
	}

	jsonData, err := json.Marshal(f)
	if err != nil {
		return "", err
	}

	filename := strings.ReplaceAll(f.Name, " ", "")

	savepath := filepath.Join(appDir, filename+".json")

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

	return filename + ".json", nil
}

func deletefunction(filename string) error {
	// deletes from disk
	// Create a directory for your app
	filepath := filepath.Join(homeDir, "Functions", filename)

	err := os.Remove(filepath)
	if err != nil {
		fmt.Println("Error deleting file:", err)
		return err
	}

	fmt.Println("File deleted successfully.")

	return nil
}
