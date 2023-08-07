package main

// create, edit, and make function calls

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

func (agent *Agent) hfunction(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		type Funcdef struct {
			Name        string
			Description string
		}

		var data struct {
			Currentfunctions []Funcdef
			Savedfunctions   []string
		}

		for _, item := range agent.req.Functions {
			newfunc := Funcdef{
				Name:        item.Name,
				Description: item.Description,
			}
			data.Currentfunctions = append(data.Currentfunctions, newfunc)
		}

		data.Savedfunctions, _ = getsavefunctionlist()
		render(w, hfunctionpage, data)
	}

	if r.Method == http.MethodPost {
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

		r.Method = http.MethodGet
		agent.hfunction(w, r)
	}

	query := strings.TrimPrefix(r.URL.Path, "/function/")

	if r.Method == http.MethodPatch {
		if query == "" {
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
		} else {
			for _, function := range agent.req.Functions {
				if query == function.Name {
					agent.hfunctionedit(w, r, function)
					continue
				}
			}

		}
	}

	if r.Method == http.MethodDelete {
		agent.removefunction(query)
		r.Method = http.MethodGet
		agent.hfunction(w, r)
	}
}

func (agent *Agent) hfunctionfiles(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimPrefix(r.URL.Path, "/function/files/")
	if !strings.HasSuffix(query, ".json") {
		query = query + ".json"
	}
	if r.Method == http.MethodGet {
		functionname := query
		newfunction, err := agent.functionload(functionname)
		if err != nil {
			fmt.Println(err)
		}
		agent.setfunction(newfunction)
		agent.hfunction(w, r)
	}

	if r.Method == http.MethodPost {
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

		r.Method = http.MethodGet
		agent.hfunction(w, r)
	}
	if r.Method == http.MethodDelete {
		functionname := query
		deletefile("Functions", functionname)
		agent.hfunction(w, r)
	}
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

	render(w, hfunctioneditpage, data)
}
