package main

// Convert pages to strings otherwise the render() function won't work for both generated and templated html

import "embed"

//go:embed templates/index.html
var hindex string

//go:embed templates/sidebar.html
var hsidebarpage string

//go:embed templates/chat.html
var hchatpage string

//go:embed templates/newchat.html
var hnewmessage string

//go:embed templates/edit.html
var hedit string

//go:embed templates/edited.html
var hedited string

//go:embed templates/save.html
var hsave string

//go:embed templates/chatload.html
var hchatloadpage string

//go:embed templates/functions.html
var hfunctionpage string

//go:embed templates/editfunction.html
var heditfunction string

//go:embed templates/prompts.html
var hpromptspage string

//go:embed templates/settings.html
var hsettingspage string

//go:embed static
var hcss embed.FS
