package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
)

func (agent *Agent) hchat(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Header   template.HTML
		Role     string
		Content  string
		Index    string
		Function template.HTML
	}

	if r.Method == http.MethodGet {
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
			for i, item := range messages[1:] {
				var content string
				lines := strings.Split(item.Content, "\n")
				for _, line := range lines {
					content += line + "<br>"
				}
				msg := message{
					Role:    item.Role,
					Content: item.Content,
					Index:   i + 1,
				}
				data.Messages = append(data.Messages, msg)
			}
			render(w, hchatpage, data)
		}
	}

	if r.Method == http.MethodPost {
		rawtext := r.FormValue("text")
		if strings.TrimSpace(rawtext) == "" {
			render(w, "", nil)
			return
		}
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

		render(w, hchatnewpage, data)
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
		render(w, hchatnewpage, data)
	}
}

func (agent *Agent) hchatedit(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimPrefix(r.URL.Path, "/chat/edit/")
	if r.Method == http.MethodGet {
		id, err := strconv.Atoi(query)
		if err != nil {
			fmt.Println(err)
		}
		data := struct {
			Edittext  string
			MessageID int
		}{
			Edittext:  agent.req.Messages[id].Content,
			MessageID: id,
		}
		render(w, hchatedit, data)
	}

	if r.Method == http.MethodPost {
		id, err := strconv.Atoi(query)
		if err != nil {
			fmt.Println(err)
		}
		edittext := r.FormValue("edittext")
		agent.req.Messages[id].Content = edittext
		newtext := `<pre style="white-space: pre-wrap; font-family: inherit;">` + edittext + `</pre>`
		render(w, newtext, nil)
	}

	if r.Method == http.MethodDelete {
		err := agent.deletelines(query)
		if err != nil {
			fmt.Println(err)
		}
		r.Method = http.MethodGet
		agent.hchat(w, r)
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
		render(w, hchatsavepage, data)
	}

	if r.Method == http.MethodPost {
		filename := r.FormValue("filename")
		agent.savefile(agent.req.Messages, "Chats", filename)
		render(w, "Chat Saved!", nil)
	}
}

func (agent *Agent) hchatfile(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimPrefix(r.URL.Path, "/chat/files/")
	if r.Method == http.MethodGet {
		if query == "" {
			var data struct {
				Filelist []string
			}
			filelist, err := getsavefilelist("Chats")
			if err != nil {
				fmt.Println(err)
			}
			data.Filelist = filelist
			render(w, hchatfilespage, data)
		} else {
			_, err := agent.loadfile("Chats", query)
			if err != nil {
				fmt.Println(err)
			}
			r.Method = http.MethodGet
			agent.hchat(w, r)
		}

	}

	if r.Method == http.MethodDelete {
		err := deletefile("Chats", query)
		if err != nil {
			fmt.Println(err)
		}
		render(w, "<p>Chat Deleted</p>", nil)
	}
}

func (agent *Agent) hchatclear(w http.ResponseWriter, r *http.Request) {
	agent.setprompt()
	r.Method = http.MethodGet
	agent.hchat(w, r)
}

func (agent *Agent) hreset(w http.ResponseWriter, r *http.Request) {
	agent.reset()
	r.Method = http.MethodGet
	agent.hchat(w, r)
}

func (agent *Agent) hsettings(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data := struct {
			Model             string
			Functionmodel     string
			Maxtokens         int
			Callcost          float64
			Autoclearfunction bool
		}{
			Model:             model,
			Functionmodel:     functionmodel,
			Maxtokens:         maxtokens,
			Callcost:          callcost,
			Autoclearfunction: autoclearfunction,
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

		r.Method = http.MethodGet
		agent.hchat(w, r)
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

func hautofunction(w http.ResponseWriter, r *http.Request) {
	buttonon := `<button class="buttonon" hx-put="/autofunction/">Autofunction</button>`
	buttonoff := `<button class="buttonoff" hx-delete="/autofunction/">Autofunction</button>`
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
	buttonon := `<button class="buttonon" hx-put="/autorequestfunction/">Autorequestfunction</button>`
	buttonoff := `<button class="buttonoff" hx-delete="/autorequestfunction/">Autorequestfunction</button>`
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
