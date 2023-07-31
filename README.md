## AgentSmith

A forge for making agents - become an agentsmith!

AgentSmith is a command-line tool, chat assistant (command-line and GUI), Python module, and IDE for using and creating AI Agents. It features everything you need to craft AI Agents and the ability to call these agents to complete tasks. Add AI functionality to any program with less than four lines of code. 

Users can freely edit/save/load prompts, chats, and functions to experiment with responses. Agents have the ability to automatically request functions and can interact with literally any other program*. 

*You may have to build an interface but there's tools/examples to cover that

#### [Demo](http://3.26.158.237:49327/)

### Features

- **Chat** - You can chat with it like any OpenAI chatbot
- **CLI and GUI** - You can interact with AgentSmith via the CLI or GUI. The CLI is mostly just for chat, but there are a few handy functions in there - type 'help' to see an overview
- **Edit/delete/save/load chats** - Allows you to easily modify chats (even change the AI's response) and store/retrieve them for later use
- **Cost Estimator** - The GUI shows estimated call costs
- **Prompt editor** - An interface for easily editing/saving/loading/deleting prompts. Makes it really easy to prompt engineer
- **Function editor** - An interface for easily adding/removing/editing/saving/deleting functions. Use this to quickly and easily build and test functions
- **Function executor** - If AgentSmith detects a function call from the chat, it will ask the user if they want to execute the function. This can either look internally for a function to execute (in the case of an Agent) or look for an app to run. The app must be in the same directory or in `$PATH` and have the same name as the function call
- **AutoFunction** - This setting allows the main chat/agent to automatically run functions rather than having the prompt for approval - be careful I guess?
- **Auto Request Functions** - The main chat/agents can automatically detect which functions are available to them and request them if needed without loading them first - this creates significant savings on token counts as you don't need to load functions that might not be used, but they are still accessible.
- **Fully Customizable Agents** - You can control which models the main chat uses, max tokencount, autofunction, and more in 'Settings'

The default directory for the AgentSmith is `~/AgentSmith`, but you can easily set this using the `-home` flag.

### Efficient

Being able to remove/edit responses means you can remove redundant information to keep token counts low while retaining important information. Use the call cost estimator to keep costs down.

Agents are able to automatically detect available functions and load them as required, saving on token counts and boosting agent ability at the same time - agents still have access to all their abilities without sacrificing memory!

An `autoclearfunction` setting (`default: true`) will automatically clear function call requests and raw responses from the assistant's memory - this is because this can get quite large and take up tokencounts super easily. Turning `autoclearfunction` off will mean you can keep the raw data in memory - just refresh the screen to view and remove manually. 

### Plugins
- **Shell** - The shell plugin allows the assistant to write/call/run shell code. There is a Python mode for generating/running Python code.
- **Search** - The search plugin currently allows the assistant to seach Wikipedia and Google.
- **Browser** - Allows the assistant to control the browser eg. "Open a browser and take me to Taylor Swift's Wikipedia page" does just that.

Note: Plugins need to be compiled and then the executable needs to be put into the same folder as the main app. The JSON file in the folder also needs to be added to `~/AgentSmith/Functions`

### How to build agents

Look at `/examples` for an examples on how to build simple agents in Golang and Python. The plugins are also good examples for how to build interfaces for the agent to call to access other programs.

`core.go` contains everything you need to build an agent in Go. The `AgentSmith.py` contains everything you need to build an agent in Python. It takes less than four lines of code.

Agents can make/receive external requests for data. The main program receives data back as a string.

Here's an example of what a call to the `searchplugin` function looks like:

    {"command": "google", "args": "Why is the sky blue?"}

This makes the main app send out the following in the command-line:

    ./searchplugin '{"command": "google", "args": "Why is the sky blue?"}'

The `searchplugin` app then parses this data and does its thing and prints a string that the assistant will try to parse. 

The agents are able to make/receive/process data in whichever way they're programmed - you can even chain agents within agents and get them to talk to each other.

This allows anyone to easily create complex AI apps with multiple agents all with different prompts/functions that can work together to do anything.

### How to run

To run as just a command-line chat, run `agentsmith --console`

To start the GUI, you just have to run: `agentsmith --gui`

(Or `agentsmith.exe --gui` on Windows, etc.)

This will start a server at http://127.0.0.1:49327 - the server is secured so only localhost can connect to it. To allow external connections, launch the app with `-ip <ipaddress>` or `-allowallips`. Use `-port` to specify port.

The default folder is `~/AgentSmith` but this can be set with the `-home` flag

You can create an agent using flags.

Flags:
- `-key` api key" (this must be first)
- `-home` set the home directory for agent
- `-save` save the chat + response to `homedir/Saves/filename.json`
- `-load` load chat from `homedir/Saves` eg `-load example.json` will load `homedir/Saves/example.json`
- `-prompt` set model prompt - otherwise there is a default assistant prompt
- `-model` model name - default is gpt-3.5-turbo
- `-maxtokens` default max tokens is 2048
- `-function` add function to agent - specify function name in `homedir/Functions` eg `-function browser` will add `homedir/Functions/browser.json`
- `-message` add message from user to chat
- `-messageassistant` add message from assistant to chat
- `-messagefunction` add message from function to chat
- `-autofunction` automatically runs functions rather than returning a function response
- `-autoclearfunctionoff` autoclearfunction removes the second and third last messages from messagelist after a function call (eg the functioncall and response) as they take up a lot of memory/tokencount - turn off autoclearfunction to keep in memory
- `-autorequestfunction` automatically detects all functions in `homedir/Functions` and makes the agent aware of these functions. Agent can then request to add the function and it will be automatically added

This can be used to build a full agent. The Python module basically follows the same idea - you set the flags/messages and then make a call.

If you want to build from source, you might want to change the PGP keys in `core.go` (this protects your API key*) or set your API key manually in the code. Then it's just `go build` and run.

*The app stores an encrypted API key in `homedir` by default. It will not do this if you specify a key with the `-key` flag.

### How to contribute

I'm just a hobby developer trying to hone my skills. If you want to help, out feel free to open an issue, make a fork, or [email me](mailto:mattyleedev@gmail.com).
