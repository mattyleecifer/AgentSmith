package main

// Convert pages to strings otherwise the render() function won't work for both generated and templated html

import "embed"

//go:embed templates/index.html
var hindexpage string

//go:embed templates/sidebar.html
var hsidebarpage string

//go:embed templates/chat.html
var hchatpage string

//go:embed templates/chatnew.html
var hchatnewpage string

//go:embed templates/chatsave.html
var hchatsavepage string

//go:embed templates/chatload.html
var hchatloadpage string

//go:embed templates/edit.html
var hedit string

//go:embed templates/edited.html
var hedited string

//go:embed templates/function.html
var hfunctionpage string

//go:embed templates/functionedit.html
var hfunctioneditpage string

//go:embed templates/prompt.html
var hpromptspage string

//go:embed templates/settings.html
var hsettingspage string

//go:embed static
var hcss embed.FS
