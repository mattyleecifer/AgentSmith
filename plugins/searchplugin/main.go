package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/chromedp/chromedp"
	"github.com/msoap/html2data"
	"github.com/sashabaranov/go-openai"
	gowiki "github.com/trietmn/go-wiki"
)

type scrapedata struct {
	name           string
	googleCardName string
	googleAddress  string
	googlePhone    string
	siteURL        string
	searchURL      string
}

type slicesd []scrapedata

func main() {
	req := getrequest()
	// fmt.Print("req", req)

	switch req["command"] {
	case "searchwiki":
		searchwiki(req["args"])
	case "getwiki":
		getwiki(req["args"])
	case "getwikisectionlist":
		getwikisectionlist(req["args"])
	case "getwikisections":
		getwikisections(req["args"])
	case "getwikisummary":
		getwikisummary(req["args"])
	case "google":
		google(req["args"])
	}
}

func searchwiki(query string) {
	parsedquery, _ := url.Parse(query)
	apiURL := "https://en.wikipedia.org/w/api.php?action=query&format=json&list=search&utf8=1&srsearch=" + parsedquery.String()

	res, err := http.Get(apiURL)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	// fmt.Println(res)

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	var searchResponse struct {
		Query struct {
			Search []struct {
				Title   string `json:"title"`
				Snippet string `json:"snippet"`
				PageID  int    `json:"pageid"`
			} `json:"search"`
		} `json:"query"`
	}

	err = json.Unmarshal(body, &searchResponse)
	if err != nil {
		fmt.Println(err)
		return
	}

	var searchResults string
	for _, result := range searchResponse.Query.Search {
		searchResults += "\nPageID: " + strconv.Itoa(result.PageID) + "\nTitle: " + result.Title + "\nSnippet: " + result.Snippet + "\n-----"
	}
	searchResults = strings.ReplaceAll(searchResults, "<span class=\"searchmatch\">", "")
	searchResults = strings.ReplaceAll(searchResults, "</span>", "")

	fmt.Println(searchResults)
}

func getwiki(title string) {
	page, err := gowiki.GetPage(title, -1, false, true)
	if err != nil {
		fmt.Println(err)
	}

	content, _ := page.GetContent()
	// fmt.Print(content)

	fmt.Println(content)
}

func getwikisectionlist(title string) {
	page, err := gowiki.GetPage(title, -1, false, true)
	if err != nil {
		fmt.Println(err)
	}

	contentlist, _ := page.GetSectionList()

	var content string
	for _, str := range contentlist {
		content += "\n" + str + "\n-----"
	}

	// fmt.Print(content)

	fmt.Println(content)
}

func getwikisummary(title string) {
	page, err := gowiki.GetPage(title, -1, false, true)
	if err != nil {
		fmt.Println(err)
	}

	content, _ := page.GetSummary()
	// fmt.Print(content)

	fmt.Println(content)
}

func getwikisections(data string) {
	datalist := strings.Split(data, "|")

	title := datalist[0]

	page, err := gowiki.GetPage(title, -1, false, true)
	if err != nil {
		fmt.Println(err)
	}

	var content string
	for _, str := range datalist[1:] {
		sectiondata, _ := page.GetSection(str)
		// fmt.Println(err)
		content += "\n" + str + ":\n" + sectiondata + "\n-----"
	}
	// fmt.Print(content)

	fmt.Println(content)
}

func google(query string) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
	)
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	var content string
	var html string

	err := chromedp.Run(ctx,
		chromedp.Navigate("https://www.google.com/search?q="+query),
		chromedp.WaitVisible(`div[id=search]`),
		chromedp.OuterHTML(`div[id=search]`, &html),
		chromedp.Text(`div[id=search]`, &content),
	)
	if err != nil {
		log.Fatal(err)
	}

	// get links/titles and then match them in the response text
	reader := bytes.NewReader([]byte(html))
	doc := html2data.FromReader(reader)
	data, _ := doc.GetData(map[string]string{"links": `a:has(h3):attr(href)`, "titles": `a h3`})
	// fmt.Println(data)
	for i, result := range data["titles"] {
		// fmt.Println(result, i)
		titlefind := regexp.MustCompile(result)
		title := "\n------\nTitle: " + result + "\nLink: " + data["links"][i] + "\nRecord data:\n"
		content = titlefind.ReplaceAllString(content, title)
	}
	contentlist := strings.Split(content, "------")
	content = ""

	for _, str := range contentlist[1:6] {
		content += str + "\n------\n"
	}
	// fmt.Println(content)

	// get ai summary
	response := summarizeGoogleContent(content)
	// fmt.Println(response)

	fmt.Println(response.Message.Content)
}

func summarizeGoogleContent(content string) Response {
	var response Response

	prompt := "You are Summarize-Content-GPT, an AI that is an expert and looking at Google result summaries entering them into records using the createsummary function"

	createsummaryfunction := `{"name":"createsummary","description":"This function accepts a series of records and it returns an array containing a title, link, and a summary of each record.","parameters":{"properties":{"Results":{"description":"An array of the summaries with a Title, Link, and Summary of the data in the record.","items":{"Link":{"description":"Record link","type":"string"},"Summary":{"description":"Summary of the record data","type":"string"},"Title":{"description":"Title of record","type":"string"}},"required":["Title","Link","Summary"],"type":"array"}},"required":["Results"],"type":"object"}}`

	agent := newAgent()
	agent.prompt.Parameters = prompt
	agent.setprompt()
	agent.setmessage(openai.ChatMessageRoleUser, content, "")

	agent.setFunctionFromString(createsummaryfunction)

	agent.req.FunctionCall = &openai.FunctionCall{
		Name:      "createsummary",
		Arguments: content,
	}

	response, _ = agent.getresponse()

	return response
}

func searchwebsite(url string) {
	res, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	// fmt.Println(res)

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(body)

}

// func (sd scrapedata) scrapeFacebook() scrapedata {
// 	doc := html2data.FromURL(API + sd.fbLink)
// 	if doc.Err != nil {
// 		fmt.Println(sd.searchURL, "failed!\nError:", doc.Err) // maybe add to log
// 	}
// 	sd.fbName, _ = doc.GetDataSingle("h1[itemprop='name']")
// 	if sd.fbName == "" {
// 		sd.fbName, _ = doc.GetDataSingle("meta[property='og:title']:attr(content)")
// 	}
// 	switch {
// 	case sd.fbName == "Entre ou cadastre-se para visualizar":
// 		sd.fbName = ""
// 	case sd.fbName == "Log In or Sign Up to View":
// 		sd.fbName = ""
// 	case sd.fbName == "Log in or sign up to view":
// 		sd.fbName = ""
// 	case sd.fbName == "Masuk atau Daftar untuk Melihat":
// 		sd.fbName = ""
// 	case sd.fbName == "Đăng nhập hoặc đăng ký để xem":
// 		sd.fbName = ""
// 	case sd.fbName == "يرجى تسجيل الدخول أو التسجيل لعرض المحتوى":
// 		sd.fbName = ""
// 	case sd.fbName == "登入或註冊即可查看":
// 		sd.fbName = ""
// 	case sd.fbName == "देखने के लिए लॉग इन या साइन अप करें":
// 		sd.fbName = ""
// 	case sd.fbName == "เข้าสู่ระบบหรือสมัครใช้งานเพื่อดู":
// 		sd.fbName = ""
// 	case sd.fbName == "Accedi o iscriviti per visualizzare":
// 		sd.fbName = ""
// 	case sd.fbName == "Inicia sesión o regístrate para verlo":
// 		sd.fbName = ""
// 	case sd.fbName == "Connectez-vous ou inscrivez-vous pour voir le contenu":
// 		sd.fbName = ""
// 	default:
// 		fmt.Println(sd.fbName)
// 	}

// 	preg := regexp.MustCompile(`\+?\d?\d? ?\d\d \d?\d\d\d\d-?\d\d\d\d`)
// 	body, _ := doc.GetDataSingle("body")
// 	sd.fbPhone = preg.FindString(body)
// 	links, _ := doc.GetData(map[string]string{"links": "div[class='_50f4']"})
// 	for _, val := range links["links"] {
// 		if strings.Contains(val, "http") {
// 			sd.fbWebsite = val
// 		}
// 	}
// 	ereg := regexp.MustCompile(` ?'?"?([a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+\.?[a-zA-Z0-9-.]+?) ?"?'?`)
// 	emails := ereg.FindString(body)
// 	emails = strings.ReplaceAll(emails, "ADICIONAIS", "")
// 	emails = strings.ReplaceAll(emails, "INFO", "")
// 	emails = strings.ReplaceAll(emails, "DETAILS", "")
// 	emails = strings.ReplaceAll(emails, "MORE", "")
// 	emails = strings.ReplaceAll(emails, "TAMBAHAN", "")
// 	emails = strings.ReplaceAll(emails, "categoriesPrivate", "")
// 	emails = strings.ReplaceAll(emails, "Favorite", "")
// 	emails = strings.ReplaceAll(emails, "categoriesShopping", "")
// 	if strings.Contains(emails, "scraperapi") {
// 		emails = ""
// 	}
// 	sd.fbEmails = strings.ReplaceAll(emails, "http", "")
// 	links, _ = doc.GetData(map[string]string{"links": "div[class='_4bl9'] a:attr(href)"})
// 	ireg := regexp.MustCompile(`.+instagram.com%2F`)
// 	exreg := regexp.MustCompile(`&.+$`)
// 	for _, val := range links["links"] {
// 		if strings.Contains(val, "instagram.com") {
// 			val = string(exreg.ReplaceAll([]byte(val), []byte("")))
// 			sd.insta = string(ireg.ReplaceAll([]byte(val), []byte("")))
// 		}
// 	}
// 	return sd
// }

// func formatFacebook(link string) string {
// 	fbreg := regexp.MustCompile(`/\d+/$`)

//		switch {
//		case strings.Contains(link, "posts"):
//			link = ""
//			fmt.Println("Invalid fblink")
//		case strings.Contains(link, "public"):
//			link = ""
//			fmt.Println("Invalid fblink")
//		case strings.Contains(link, "groups"):
//			link = ""
//			fmt.Println("Invalid fblink")
//		case strings.Contains(link, "photos"):
//			link = ""
//			fmt.Println("Invalid fblink")
//		case strings.Contains(link, "people"):
//			link = ""
//			fmt.Println("Invalid fblink")
//		case strings.Contains(link, "videos"):
//			link = ""
//			fmt.Println("Invalid fblink")
//		case strings.Contains(link, "directory"):
//			link = ""
//			fmt.Println("Invalid fblink")
//		case strings.Contains(link, "/search?gl=BR"):
//			link = ""
//			fmt.Println("Invalid fblink")
//		case strings.Contains(link, "events"):
//			link = ""
//			fmt.Println("Invalid fblink")
//		case strings.Contains(link, "about"):
//			link = link
//			fmt.Println("It worked")
//		case fbreg.MatchString(link):
//			link = link
//			fmt.Println("Matched string with regex")
//		case strings.HasSuffix(link, "/"):
//			link = link + "about/"
//			fmt.Println("Ends with /")
//		default:
//			link = link + "/about/"
//			fmt.Println("Didn't work")
//		}
//		//filter out the translate link with regex
//		treg := regexp.MustCompile(`^http.+?http`)
//		link = string(treg.ReplaceAll([]byte(link), []byte("http")))
//		exreg := regexp.MustCompile(`&.+$`)
//		return string(exreg.ReplaceAll([]byte(link), []byte("")))
//	}
