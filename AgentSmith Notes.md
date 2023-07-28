AgentSmith Notes

Bugs:
- Requestfunction sometimes spams like a million requests at once - figure out why
- keeping api key in folder won't work if you are depending on it being in anything other than the default folder - you must specify a key for all production agents - not sure if bug


Change notes:
- completely revamped code and UI - worth re-reading README for new features


http://127.0.0.1:5000/launchbrowser/
http://127.0.0.1:5000/browserget/www.google.com/r/AskReddit/comments/157bdw4/what_is_denied_by_many_people_but_it_is_actually/
http://127.0.0.1:5000/browserclickcss/a%5Bclass=%22gb_pa%20gb_md%20gb_Od%20gb_me%22%5D/
http://127.0.0.1:5000/gettagfromtext/How%20Search%20works/
http://127.0.0.1:5000/gettextcoordinates/

./browser '{"command":"launchbrowser"}'
./browser '{"command":"browserget", "args": "www.reddit.com/r/AskReddit/comments/157bdw4/what_is_denied_by_many_people_but_it_is_actually"}'
