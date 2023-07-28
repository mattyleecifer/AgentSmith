package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

var browserapi string = "http://127.0.0.1:8210/"

func main() {
	req := getrequest()
	// fmt.Println(req)

	switch req["command"] {
	case "launchbrowser":
		launchbrowser()
	case "browserget":
		browserget(req["args"])
	case "browsergettext":
		browsergettext(req["args"])
	case "browserclick":
		browserclick(req["args"])
	case "browsersendtext":
		browsersendtext(req["args"])
	case "getcurrenturl":
		launchbrowser()
	case "getbodytext":
		getbodytext()
	case "scrollup":
		scrollup()
	case "scrolldown":
		scrolldown()
	case "default":
	}
}

func launchbrowser() {
	apiURL := browserapi + "launchbrowser/"
	body := getbody(apiURL)
	printresponse(body)
}

func scrollup() {
	apiURL := browserapi + "scrollup/"
	body := getbody(apiURL)
	printresponse(body)
}

func scrolldown() {
	apiURL := browserapi + "scrolldown/"
	body := getbody(apiURL)
	printresponse(body)
}

func browserget(input string) {
	// link = url.QueryEscape(link)
	req := getsubrequest(input)
	link := req[0]
	link = strings.Replace(link, "http://", "", 1)
	link = strings.Replace(link, "https://", "", 1)
	apiURL := browserapi + "browserget/" + link + "/"
	body := getbody(apiURL)
	printresponse(body)
}

func browserclick(input string) {
	req := getsubrequest(input)
	// link = url.QueryEscape(link)
	switch req[0] {
	case "css":
		apiURL := browserapi + "browserclickcss/" + req[1] + "/"
		body := getbody(apiURL)
		printresponse(body)
	case "xpath":
		apiURL := browserapi + "browserclickxpath/" + req[1] + "/"
		body := getbody(apiURL)
		printresponse(body)
	}
}

func browsergettext(input string) {
	req := getsubrequest(input)
	// link = url.QueryEscape(link)
	switch req[0] {
	case "css":
		apiURL := browserapi + "browsergettextcss/" + req[1] + "/"
		body := getbody(apiURL)
		printresponse(body)
	case "xpath":
		apiURL := browserapi + "browsergettextxpath/" + req[1] + "/"
		body := getbody(apiURL)
		printresponse(body)
	}
}

func browsersendtext(input string) {
	req := getsubrequest(input)
	// link = url.QueryEscape(link)
	switch req[0] {
	case "css":
		apiURL := browserapi + "browsersendtextcss/" + req[1] + "/" + req[2] + "/"
		body := getbody(apiURL)
		printresponse(body)
	case "xpath":
		apiURL := browserapi + "browsersendtextxpath/" + req[1] + "/" + req[2] + "/"
		body := getbody(apiURL)
		printresponse(body)
	}
}

func getbodytext() {
	apiURL := browserapi + "getbodytext/"
	body := getbody(apiURL)
	printresponse(body)
}

func printresponse(body []byte) {
	var apiResponse map[string]string
	err := json.Unmarshal(body, &apiResponse)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(apiResponse["message"])
}

func getbody(apiURL string) []byte {
	res, err := http.Get(apiURL)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer res.Body.Close()

	// fmt.Println(res)

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return body
}
