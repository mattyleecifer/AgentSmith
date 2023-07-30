package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sashabaranov/go-openai"
)

func hloadchatpage(w http.ResponseWriter, r *http.Request) {
	render(w, hloadchat, nil)
}

func hloadsettings(w http.ResponseWriter, r *http.Request) {
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
	render(w, hsettings, data)
}

func (agent *Agent) hsetsettings(w http.ResponseWriter, r *http.Request) {
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
	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func hgetsidebar(w http.ResponseWriter, r *http.Request) {
	render(w, hsidebar, nil)
}

func hsidebaroff(w http.ResponseWriter, r *http.Request) {
	button := `<div class="sidebar" id="sidebar" style="flex: none;"><button class="btn" id="floating-button" hx-get="/getsidebar" hx-target="#sidebar" hx-swap="outerHTML">Show Menu</button></div>`
	render(w, button, nil)
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

func (agent *Agent) hchat(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("hchat")
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
			data.Run = template.HTML("<input type='hidden' name='functionname' value='" + response.FunctionCall.Name + "'><button class='btn'>Run</button>")
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

func (agent *Agent) hloadmessages(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("hloadmessages")
	messages := agent.req.Messages
	if len(messages) == 1 {
		messagelist := "<table><tr id='chattext'><td id='centertext'><div hx-get='/tokenupdate' hx-trigger='load' hx-target='#tokens' hx-swap='innerHTML'>Start asking questions!</div></td></tr></table>"
		render(w, messagelist, nil)
	} else {
		messagelist := "<table>"
		for i := 0; i < len(messages); i++ {
			if messages[i].Content == "" {
				continue
			}
			chatid := fmt.Sprint(i)

			messagelist += `<table><tr><td class="agent">
						` + messages[i].Role + `</td>
						<td id="reply-` + chatid + `" class="message">
						<div messageid="` + chatid + `">`

			lines := strings.Split(messages[i].Content, "\n")
			for _, line := range lines {
				messagelist += line + "<br>"
			}

			messagelist += `</td><td class="editbutton">
						<form hx-get="/edit" hx-target="#reply-` + chatid + `" hx-swap="outerHTML">
						<input type="hidden" name="messageid" value="` + chatid + `">
						<button class="btn">Edit</button>
						</form>
						<form hx-post="/delete" hx-target="#top-row" hx-swap="innerHTML">
						<input type="hidden" name="messageid" value="` + chatid + `">
						<button class="btn">Delete</button>
						</form>
						</td>
						</tr>`

		}
		messagelist += `<tr id="chattext" hx-get="/scroll" hx-trigger="load" hx-target="this" hx-swap="none, show:bottom"></tr></table>`
		render(w, messagelist, nil)
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
			chatid := strings.ReplaceAll(filelist[i], ".", "")
			html += "<tr id='savedchat" + chatid + "' style='text-align: left;'><td>"
			html += "<div class='savedchat'><div>"
			html += filelist[i]
			html += "</div><td><form hx-post='/load' hx-target='this' hx-swap='innerHTML'><input type='hidden' name='data' value='" + filelist[i] + "'><button class='btn'>Load</button></form></td><td><form hx-post='/deletechathistory' hx-target='#savedchat" + chatid + "' hx-swap='outerHTML' hx-confirm='Are you sure?'><input type='hidden' name='chatid' value='" + filelist[i] + "'><button class='btn'>Delete</button></form></td>"
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
	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (agent *Agent) hdelete(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("hdelete")
	messageid := r.FormValue("messageid")
	err := agent.deletelines(messageid)
	if err != nil {
		fmt.Println(err)
	}
	agent.hloadmessages(w, r)
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
	// fmt.Println("hsave")
	// filename, err := agent.save()
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// data := struct {
	// 	Chatid   string
	// 	Targetid string
	// }{
	// 	Chatid:   filename,
	// 	Targetid: strings.ReplaceAll(filename, ".", ""),
	// }
	agent.save()
	render(w, hsave, nil)
}

func (agent *Agent) hclear(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("hclear")
	agent.setprompt()
	agent.hloadmessages(w, r)
	// w.Header().Set("HX-Redirect", "/")
	// w.WriteHeader(http.StatusTemporaryRedirect)
}

func hdeletechathistory(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("hdeletechathistory")
	chatid := r.FormValue("chatid")
	err := deletesave(chatid)
	if err != nil {
		fmt.Println(err)
	}
	render(w, "", nil)
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
	// Create a directory for your app
	filepath := filepath.Join(homeDir, "Saves", filename)

	err := os.Remove(filepath)
	if err != nil {
		fmt.Println("Error deleting file:", err)
		return err
	}

	fmt.Println("File deleted successfully.")

	return nil
}
