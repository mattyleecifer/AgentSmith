package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
)

func hsavedchats(w http.ResponseWriter, r *http.Request) {
	render(w, hsavedchatspage, nil)
}

func (agent *Agent) hsettings(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data := struct {
			Model         string
			Functionmodel string
			Maxtokens     string
			Callcost      string
		}{
			Model:         model,
			Functionmodel: functionmodel,
			Maxtokens:     strconv.Itoa(maxtokens),
			Callcost:      strconv.FormatFloat(callcost, 'f', -1, 64),
		}
		render(w, hsettingspage, data)
	}
	if r.Method == http.MethodPut {
		apikey := r.FormValue("apikey")
		if apikey != "" {
			c := openai.NewClient(apikey)
			agent.client = c
		}
		model = r.FormValue("chatmodel")
		functionmodel = r.FormValue("functionmodel")
		maxtokens, _ = strconv.Atoi(r.FormValue("maxtokens"))
		callcost, _ = strconv.ParseFloat(r.FormValue("callcost"), 64)
		autoclearfunction, _ = strconv.ParseBool(r.FormValue("autoclearfunction"))
		agent.hloadchatscreen(w, r)
	}
}

func hsidebar(w http.ResponseWriter, r *http.Request) {
	mode := strings.TrimPrefix(r.URL.Path, "/sidebar/")
	switch mode {
	case "on":
		render(w, hsidebarpage, nil)
	case "off":
		button := `<div class="sidebar" id="sidebar" style="width: 0; background-color: transparent;"><button class="btn" id="floating-button" hx-get="/sidebar/on" hx-target="#sidebar" hx-swap="outerHTML">Show Menu</button></div>`
		render(w, button, nil)
	}
}

func (agent *Agent) htokenupdate(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("htokenupdate")
	estcost := (float64(agent.tokencount) / 1000) * callcost
	data := struct {
		Tokencount string
		Estcost    string
	}{
		Tokencount: strconv.Itoa(agent.tokencount),
		Estcost:    strconv.FormatFloat(estcost, 'f', 6, 64),
	}
	render(w, htokencount, data)
}

func (agent *Agent) hgetresponse(w http.ResponseWriter, r *http.Request) {
	response, err := agent.getresponse()
	if err != nil {
		fmt.Println(err)
	}
	var data struct {
		Message   string
		MessageID int
		Run       template.HTML
	}
	if response.FunctionCall != nil {
		if autofunction {
			functionresponse := agent.callfunction(&response)
			data.Message = functionresponse.Message.Content
			data.MessageID = len(agent.req.Messages) - 1
		} else {
			data.Message = response.Message.Content
			data.MessageID = len(agent.req.Messages) - 1
			data.Run = template.HTML("<button class='btn' name='functionname' value='" + response.FunctionCall.Name + "'>Run</button>")
		}
	} else {
		data.Message = response.Message.Content
		data.MessageID = len(agent.req.Messages) - 1
	}

	render(w, hnewchat, data)

}

func (agent *Agent) hsubmit(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("hsubmit")
	text := agent.req.Messages[len(agent.req.Messages)-1].Content
	data := struct {
		Usertext  string
		MessageID string
	}{
		Usertext:  text,
		MessageID: strconv.Itoa(len(agent.req.Messages) - 1),
	}
	render(w, husermessage, data)
}

func hscroll(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("hscroll")
	render(w, "", nil)
}

func (agent *Agent) hchat(w http.ResponseWriter, r *http.Request) {

}

func (agent *Agent) hloadchatscreen(w http.ResponseWriter, r *http.Request) {
	type message struct {
		Role    string
		Content string
		Index   int
	}
	var data struct {
		Messages []message
	}

	messages := agent.req.Messages
	if len(messages) == 1 {
		render(w, hchatpage, data)
	} else {
		for i, item := range messages {
			var content string
			lines := strings.Split(item.Content, "\n")
			for _, line := range lines {
				content += line + "<br>"
			}
			msg := message{
				Role:    item.Role,
				Content: item.Content,
				Index:   i,
			}
			data.Messages = append(data.Messages, msg)
		}
		render(w, hchatpage, data)
	}
}

func (agent *Agent) hclearchat(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("hclearchat")
	rawtext := r.FormValue("text")
	if rawtext == "!" {
		agent.setprompt()
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	}
	query := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: rawtext,
	}
	agent.req.Messages = append(agent.req.Messages, query)
	render(w, hinputbox, nil)
}

func hgetchathistory(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("hgetchathistory")
	filelist, err := getsavefilelist()
	if err != nil {
		fmt.Println(err)
	}
	if filelist == nil {
		html := "<div id='addchat'></div>"
		render(w, html, nil)
		return
	} else {
		html := `<table style="display: flex;" id="centertext">`
		for i := 0; i < len(filelist); i++ {
			chatid := strings.ReplaceAll(filelist[i], ".json", "")
			html += "<tr id='savedchat" + chatid + "' style='text-align: left;'><td>"
			html += "<div class='savedchat'><div>"
			html += filelist[i]
			html += "</div><td><form hx-post='/load' hx-target='#main-content' hx-swap='innerHTML'><button class='btn' name='data' value='" + filelist[i] + "'>Load</button></form></td><td><form hx-delete='/delete/chat/" + chatid + "' hx-target='#savedchat" + chatid + "' hx-swap='outerHTML' hx-confirm='Are you sure?'><button class='btn'>Delete</button></form></td>"
			html += `</tr>`
		}
		html += "</table><div id='addchat'></div>"
		render(w, html, nil)
	}
}

func (agent *Agent) hreset(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("hreset")
	agent.reset()
	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (agent *Agent) hload(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("hload")
	data := r.FormValue("data")
	agent.load(data)
	agent.hloadchatscreen(w, r)
}

func (agent *Agent) hdeletelines(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("hdelete")
	messageid := r.FormValue("messageid")
	err := agent.deletelines(messageid)
	if err != nil {
		fmt.Println(err)
	}
	agent.hloadchatscreen(w, r)
}

func (agent *Agent) hedit(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("hedit")
	if r.Method == http.MethodGet {
		messageid := r.FormValue("messageid")
		id, err := strconv.Atoi(messageid)
		if err != nil {
			fmt.Println(err)
		}
		edittext := agent.req.Messages[id].Content
		data := struct {
			Edittext  string
			MessageID string
		}{
			Edittext:  edittext,
			MessageID: messageid,
		}
		render(w, hedit, data)
	} else if r.Method == http.MethodPost {
		messageid := r.FormValue("messageid")
		id, err := strconv.Atoi(messageid)
		if err != nil {
			fmt.Println(err)
		}
		edittext := r.FormValue("edittext")
		agent.req.Messages[id].Content = edittext
		newtext := strings.Split(edittext, "\n")
		data := struct {
			Edittext  []string
			MessageID string
		}{
			Edittext:  newtext,
			MessageID: messageid,
		}
		render(w, hedited, data)
	}
}

func (agent *Agent) hsave(w http.ResponseWriter, r *http.Request) {
	rawquery := strings.TrimPrefix(r.URL.Path, "/save/")
	query := strings.Split(rawquery, "/")
	switch query[0] {
	case "chat":
		if r.Method == http.MethodGet {
			currentTime := time.Now()
			filename := currentTime.Format("20060102150405")
			data := struct {
				Filename string
			}{
				Filename: filename,
			}
			render(w, hsave, data)
		}

		if r.Method == http.MethodPost {
			filename := r.FormValue("filename")
			agent.save(filename)
			render(w, "Chat Saved!", nil)
		}
		if r.Method == http.MethodDelete {
			chatid := query[1]
			err := deletesave(chatid + ".json")
			if err != nil {
				fmt.Println(err)
			}
			render(w, "<tr><td>Chat Deleted</td></tr>", nil)
		}
	}
}

func (agent *Agent) hdelete(w http.ResponseWriter, r *http.Request) {
	rawquery := strings.TrimPrefix(r.URL.Path, "/delete/")
	query := strings.Split(rawquery, "/")
	switch query[0] {
	case "chat":
		chatid := query[1]
		err := deletesave(chatid + ".json")
		if err != nil {
			fmt.Println(err)
		}
		render(w, "<tr><td>Chat Deleted</td></tr>", nil)
	}
}

func (agent *Agent) hclear(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("hclear")
	agent.setprompt()
	agent.hloadchatscreen(w, r)
}

func (agent *Agent) hrunfunction(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("hrunfunction")
	// parameters := r.FormValue("functioncall")
	// parameters = strings.ReplaceAll(parameters, "\n", "")
	// parameters = strings.ReplaceAll(parameters, "  ", "")
	// function := openai.FunctionDefinition{
	// 	Name:       r.FormValue("functionname"),
	// 	Parameters: parameters,
	// }
	function := Response{
		FunctionCall: &openai.FunctionCall{
			Name:      r.FormValue("functionname"),
			Arguments: agent.req.Messages[len(agent.req.Messages)-1].Content,
		},
	}

	response := agent.callfunction(&function)

	// agent.req.Messages = append(agent.req.Messages, response.Message)

	var data struct {
		Message   string
		MessageID int
		Run       string
	}

	data.Message = response.Message.Content
	data.MessageID = len(agent.req.Messages) - 1
	render(w, hnewchat, data)
}

func hfunctionloading(w http.ResponseWriter, r *http.Request) {
	fmt.Println("hfunctionloading")
	var data struct {
		Function template.HTML
	}

	data.Function = template.HTML("<input type='hidden' name='functionname' value='" + r.FormValue("functionname") + "'>")

	render(w, hfunctionloadingtemplate, data)
}

func (agent *Agent) hautorequestfunctionoff(w http.ResponseWriter, r *http.Request) {
	// remove autorequestfunction
	agent.removefunction("requestfunction")

	autorequestfunction = false
	button := `<button class="menubtn" style="background-color: darkred;" hx-post="/autorequestfunctionon" hx-target="#autorequestfunctiontoggle" hx-swap="innerHTML">Autorequestfunction</button>`
	render(w, button, nil)
}

func (agent *Agent) hautorequestfunctionon(w http.ResponseWriter, r *http.Request) {
	autorequestfunction = true
	agent.setAutoRequestFunction()
	button := `<button class="menubtn" style="background-color: darkgreen;" hx-post="/autorequestfunctionoff" hx-target="#autorequestfunctiontoggle" hx-swap="innerHTML">Autorequestfunction</button>`
	render(w, button, nil)
}

func hautorequestfunctionstatus(w http.ResponseWriter, r *http.Request) {
	if autorequestfunction {
		button := `<button class="menubtn" style="background-color: darkgreen;" hx-post="/autorequestfunctionoff" hx-target="#autorequestfunctiontoggle" hx-swap="innerHTML">Autorequestfunction</button>`
		render(w, button, nil)
	} else {
		button := `<button class="menubtn" style="background-color: darkred;" hx-post="/autorequestfunctionon" hx-target="#autorequestfunctiontoggle" hx-swap="innerHTML">Autorequestfunction</button>`
		render(w, button, nil)
	}
}

func hautofunctionoff(w http.ResponseWriter, r *http.Request) {
	autofunction = false
	button := `<button class="menubtn" style="background-color: darkred;" hx-post="/autofunctionon" hx-target="#autofunctiontoggle" hx-swap="innerHTML">Autofunction</button>`
	render(w, button, nil)
}

func hautofunctionon(w http.ResponseWriter, r *http.Request) {
	autofunction = true
	button := `<button class="menubtn" style="background-color: darkgreen;" hx-post="/autofunctionoff" hx-target="#autofunctiontoggle" hx-swap="innerHTML">Autofunction</button>`
	render(w, button, nil)
}

func hautofunctionstatus(w http.ResponseWriter, r *http.Request) {
	if autofunction {
		button := `<button class="menubtn" style="background-color: darkgreen;" hx-post="/autofunctionoff" hx-target="#autofunctiontoggle" hx-swap="innerHTML">Autofunction</button>`
		render(w, button, nil)
	} else {
		button := `<button class="menubtn" style="background-color: darkred;" hx-post="/autofunctionon" hx-target="#autofunctiontoggle" hx-swap="innerHTML">Autofunction</button>`
		render(w, button, nil)
	}

}

func getsavefilelist() ([]string, error) {
	// Create a directory for your app
	savepath := filepath.Join(homeDir, "Saves")
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

func deletesave(filename string) error {
	var savepath string
	if strings.HasSuffix(filename, ".json") {
		savepath = filepath.Join(homeDir, "Saves", filename)
	} else {
		savepath = filepath.Join(homeDir, "Saves", filename+".json")
	}

	err := os.Remove(savepath)
	if err != nil {
		fmt.Println("Error deleting file:", err)
		return err
	}

	fmt.Println("File deleted successfully.")

	return nil
}
