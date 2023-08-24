package main

// Convert pages to strings otherwise the render() function won't work for both generated and templated html

import "embed"

//go:embed templates/index.html
var hindexpage string

//go:embed templates/chat.html
var hchatpage string

//go:embed templates/chatnew.html
var hchatnewpage string

//go:embed templates/chatedit.html
var hchatedit string

//go:embed templates/chatsave.html
var hchatsavepage string

//go:embed templates/chatfiles.html
var hchatfilespage string

//go:embed templates/settings.html
var hsettingspage string

//go:embed templates/sidebar.html
var hsidebarpage string

//go:embed templates/function.html
var hfunctionpage string

//go:embed templates/functionedit.html
var hfunctioneditpage string

//go:embed templates/prompt.html
var hpromptspage string

//go:embed templates/auth.html
var hauthpage string

//go:embed static
var hcss embed.FS
