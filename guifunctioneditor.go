package main

// create, edit, and make function calls

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

func (agent *Agent) hfunction(w http.ResponseWriter, r *http.Request) {
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
			currentfunctions += "<tr style='text-align: left;'><td>" + name + ":<br>" + description + "<br></td><td><form hx-get='/functioneditcurrent' hx-target='#main-content' hx-swap='innerHTML'><button class='btn' name='functionname' value='" + name + "'>Edit</button></form><br></td><td><form hx-get='/functionremove' hx-target='#main-content' hx-swap='innerHTML'><button class='btn' name='functionname' value='" + name + "'>Remove</button></form><br></td></tr>"
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
			savedfunctions += "<tr><td style='text-align: left;'>" + name + "</td><td><form hx-post='/functionload' hx-target='#main-content' hx-swap='innerHTML'><button class='btn' name='functionname' value='" + name + "'>Load</button></form></td><td><form hx-post='/functiondelete' hx-target='#main-content' hx-swap='innerHTML' hx-confirm='Are you sure?'><button class='btn' name='functionname' value='" + name + "'>Delete</button></form></td></tr>"
		}
		savedfunctions += `</table>`
	}
	tcurrentfunctions := template.HTML(currentfunctions)
	tsavedfunctions := template.HTML(savedfunctions)
	data.CurrentFunctions = tcurrentfunctions
	data.SavedFunctions = tsavedfunctions

	render(w, hfunctionpage, data)
}

func (agent *Agent) hfunctionremove(w http.ResponseWriter, r *http.Request) {
	functionname := r.FormValue("functionname")
	agent.removefunction(functionname)
	agent.hfunction(w, r)
}

func (agent *Agent) hfunctiondelete(w http.ResponseWriter, r *http.Request) {
	functionname := r.FormValue("functionname")
	functionname += ".json"
	deletefile("Functions", functionname)
	agent.hfunction(w, r)
}

func (agent *Agent) hfunctionload(w http.ResponseWriter, r *http.Request) {
	functionname := r.FormValue("functionname")
	functionname += ".json"
	newfunction, err := agent.functionload(functionname)
	if err != nil {
		fmt.Println(err)
	}
	agent.setfunction(newfunction)
	agent.hfunction(w, r)
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
	agent.hfunction(w, r)
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

	agent.savefile(newfunction, "Functions", newfunction.Name)

	agent.hfunction(w, r)
}
