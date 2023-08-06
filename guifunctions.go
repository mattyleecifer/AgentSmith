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
	if r.Method == http.MethodGet {
		w.Header().Set("HX-Trigger-After-Settle", `tokenupdate`)
		render(w, hsidebarpage, nil)
	}
	if r.Method == http.MethodDelete {
		button := `<div class="sidebar" id="sidebar" style="width: 0; background-color: transparent;"><button class="btn" id="floating-button" hx-get="/sidebar/" hx-target="#sidebar" hx-swap="outerHTML">Show Menu</button></div>`
		render(w, button, nil)
	}
}

func (agent *Agent) htokenupdate(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("htokenupdate")
	estcost := (float64(agent.tokencount) / 1000) * callcost
	tokencount := strconv.Itoa(agent.tokencount)
	estcoststr := strconv.FormatFloat(estcost, 'f', 6, 64)
	render(w, "#Tokens: "+tokencount+"<br>$Est: "+estcoststr, nil)
}

func hscroll(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("hscroll")
	render(w, "", nil)
}

func (agent *Agent) hchat(w http.ResponseWriter, r *http.Request) {
	rawquery := strings.TrimPrefix(r.URL.Path, "/chat/")
	query := strings.Split(rawquery, "/")
	switch query[0] {
	case "":
		if r.Method == http.MethodGet {
			agent.hloadchatscreen(w, r)
		} else {
			agent.hchatpostput(w, r)
		}
	case "clear":
		agent.setprompt()
		agent.hloadchatscreen(w, r)
	}
}

func (agent *Agent) hchatsave(w http.ResponseWriter, r *http.Request) {

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
		agent.savefile(agent.req.Messages, "Chats", filename)
		render(w, "Chat Saved!", nil)
	}

}

func (agent *Agent) hchatdelete(w http.ResponseWriter, r *http.Request) {
	rawquery := strings.TrimPrefix(r.URL.Path, "/chat/delete/")
	query := strings.Split(rawquery, "/")
	fmt.Println(rawquery, query)
	switch query[0] {
	case "line":
		err := agent.deletelines(query[1])
		if err != nil {
			fmt.Println(err)
		}
		agent.hloadchatscreen(w, r)
	case "savedchat":
		err := deletesave(query[1] + ".json")
		if err != nil {
			fmt.Println(err)
		}
		render(w, "<tr><td>Chat Deleted</td></tr>", nil)
	}
}

func (agent *Agent) hchatclear(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("hclear")
	agent.setprompt()
	agent.hloadchatscreen(w, r)
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

func (agent *Agent) hchatpostput(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Header   template.HTML
		Role     string
		Content  string
		Index    string
		Function template.HTML
	}
	if r.Method == http.MethodPost {
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
		// text := agent.req.Messages[len(agent.req.Messages)-1].Content

		data.Header = template.HTML(`<div id="message" class="message">`)
		data.Role = openai.ChatMessageRoleUser
		data.Content = rawtext
		data.Index = strconv.Itoa(len(agent.req.Messages) - 1)

		render(w, hnewmessage, data)
	}
	if r.Method == http.MethodPut {
		response, err := agent.getresponse()
		if err != nil {
			fmt.Println(err)
		}

		w.Header().Set("HX-Trigger-After-Settle", `tokenupdate`)

		data.Role = openai.ChatMessageRoleAssistant
		data.Header = template.HTML(`<div id="message" class="message" style="background-color: #393939">`)

		if response.FunctionCall != nil {
			if autofunction {
				functionresponse := agent.callfunction(&response)
				data.Content = functionresponse.Message.Content
				data.Index = strconv.Itoa(len(agent.req.Messages) - 1)
			} else {
				data.Content = response.Message.Content
				data.Index = strconv.Itoa(len(agent.req.Messages) - 1)
				data.Function = template.HTML(`<button hx-post="/function/run/` + response.FunctionCall.Name + `/" hx-indicator="#chatloading" hx-target="#chatloading" hx-swap="beforebegin scroll:#top-row:bottom" hx-select="#message">Run</button>`)
			}
		} else {
			data.Content = response.Message.Content
			data.Index = strconv.Itoa(len(agent.req.Messages) - 1)
		}

		render(w, hnewmessage, data)
	}
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
			html += "</div><td><form hx-post='/chat/load/' hx-target='#main-content' hx-swap='innerHTML'><button class='btn' name='data' value='" + filelist[i] + "'>Load</button></form></td><td><button class='btn' hx-delete='/chat/delete/savedchat/" + chatid + "/' hx-target='#savedchat" + chatid + "' hx-swap='outerHTML' hx-confirm='Are you sure?'>Delete</button></form></td>"
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

func (agent *Agent) hchatload(w http.ResponseWriter, r *http.Request) {
	filename := r.FormValue("data")
	_, err := agent.loadfile("Chats", filename)
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

func (agent *Agent) hfunctionrun(w http.ResponseWriter, r *http.Request) {
	rawquery := strings.TrimPrefix(r.URL.Path, "/function/run/")
	query := strings.Split(rawquery, "/")
	fmt.Println(rawquery, query)

	function := Response{
		FunctionCall: &openai.FunctionCall{
			Name:      query[0],
			Arguments: agent.req.Messages[len(agent.req.Messages)-1].Content,
		},
	}

	response := agent.callfunction(&function)

	w.Header().Set("HX-Trigger-After-Settle", `tokenupdate`)

	var data struct {
		Header   template.HTML
		Role     string
		Content  string
		Index    string
		Function string
	}
	data.Header = template.HTML(`<div id="message" class="message" style="background-color: #393939">`)
	data.Role = openai.ChatMessageRoleAssistant
	data.Content = response.Message.Content
	data.Index = strconv.Itoa(len(agent.req.Messages) - 1)
	render(w, hnewmessage, data)
}

func hautofunction(w http.ResponseWriter, r *http.Request) {
	buttonon := `<button class="menubtn" style="background-color: darkred;" hx-put="/autofunction/" hx-target="#autofunctiontoggle" hx-swap="innerHTML">Autofunction</button>`
	buttonoff := `<button class="menubtn" style="background-color: darkgreen;" hx-delete="/autofunction/" hx-target="#autofunctiontoggle" hx-swap="innerHTML">Autofunction</button>`
	if r.Method == http.MethodGet {
		if autofunction {
			render(w, buttonoff, nil)
		} else {
			render(w, buttonon, nil)
		}
	}
	if r.Method == http.MethodPut {
		autofunction = true
		render(w, buttonoff, nil)
	}
	if r.Method == http.MethodDelete {
		autofunction = false
		render(w, buttonon, nil)
	}
}

func (agent *Agent) hautorequestfunction(w http.ResponseWriter, r *http.Request) {
	buttonon := `<button class="menubtn" style="background-color: darkred;" hx-put="/autorequestfunction/" hx-target="#autorequestfunctiontoggle" hx-swap="innerHTML">Autorequestfunction</button>`
	buttonoff := `<button class="menubtn" style="background-color: darkgreen;" hx-delete="/autorequestfunction/" hx-target="#autorequestfunctiontoggle" hx-swap="innerHTML">Autorequestfunction</button>`
	if r.Method == http.MethodGet {
		if autorequestfunction {
			render(w, buttonoff, nil)
		} else {
			render(w, buttonon, nil)
		}
	}
	if r.Method == http.MethodPut {
		autorequestfunction = true
		agent.setAutoRequestFunction()
		render(w, buttonoff, nil)
	}
	if r.Method == http.MethodDelete {
		agent.removefunction("requestfunction")
		autorequestfunction = false
		render(w, buttonon, nil)
	}
}

func getsavefilelist() ([]string, error) {
	// Create a directory for your app
	savepath := filepath.Join(homeDir, "Chats")
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
		savepath = filepath.Join(homeDir, "Chats", filename)
	} else {
		savepath = filepath.Join(homeDir, "Chats", filename+".json")
	}

	err := os.Remove(savepath)
	if err != nil {
		fmt.Println("Error deleting file:", err)
		return err
	}

	fmt.Println("File deleted successfully.")

	return nil
}
