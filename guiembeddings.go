package main

// Convert pages to strings otherwise the render() function won't work for both generated and templated html

import "embed"

//go:embed templates/index.html
var hindex string

//go:embed templates/tokencount.html
var htokencount string

//go:embed templates/newchat.html
var hnewchat string

//go:embed templates/usermessage.html
var husermessage string

//go:embed templates/inputbox.html
var hinputbox string

//go:embed templates/edit.html
var hedit string

//go:embed templates/edited.html
var hedited string

//go:embed templates/save.html
var hsave string

//go:embed templates/functions.html
var hfunctionspage string

//go:embed templates/editfunction.html
var heditfunction string

//go:embed templates/functionloading.html
var hfunctionloadingtemplate string

//go:embed templates/editprompt.html
var heditprompt string

//go:embed templates/settings.html
var hsettingspage string

//go:embed templates/loadchat.html
var hloadchat string

//go:embed templates/sidebar.html
var hsidebarpage string

//go:embed templates/chatscreen.html
var hchatscreen string

//go:embed static
var hcss embed.FS
