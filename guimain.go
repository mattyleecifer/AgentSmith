package main

import (
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
)

func (agent *Agent) gui() {
	http.HandleFunc("/", RequireAuth(index))
	http.HandleFunc("/chat/", RequireAuth(agent.hchat))
	http.HandleFunc("/chat/save/", RequireAuth(agent.hchatsave))
	http.HandleFunc("/chat/load/", RequireAuth(agent.hchatload))
	http.HandleFunc("/chat/delete/", RequireAuth(agent.hchatdelete))
	http.HandleFunc("/chat/clear/", RequireAuth(agent.hchatclear))
	http.HandleFunc("/chat/reset/", RequireAuth(agent.hreset))
	http.HandleFunc("/edit", RequireAuth(agent.hedit))
	http.HandleFunc("/settings/", RequireAuth(agent.hsettings))
	http.HandleFunc("/sidebar/", RequireAuth(hsidebar))
	http.HandleFunc("/sidebar/tokens/", RequireAuth(agent.htokenupdate))
	http.HandleFunc("/autofunction/", RequireAuth(hautofunction))
	http.HandleFunc("/autorequestfunction/", RequireAuth(agent.hautorequestfunction))
	http.HandleFunc("/function/", RequireAuth(agent.hfunction))
	http.HandleFunc("/function/files/", RequireAuth(agent.hfunctionfiles))
	http.HandleFunc("/function/run/", RequireAuth(agent.hfunctionrun))
	http.HandleFunc("/prompt/", RequireAuth(agent.hprompt))
	http.HandleFunc("/prompt/files/", RequireAuth(agent.hpromptfiles))

	http.Handle("/static/", http.FileServer(http.FS(hcss)))
	fmt.Println("Running GUI on http://127.0.0.1"+port, "(ctrl-click link to open)")
	log.Fatal(http.ListenAndServe(port, nil))
}

func index(w http.ResponseWriter, r *http.Request) {
	render(w, hindexpage, nil)
}

func render(w http.ResponseWriter, html string, data any) {
	// Render the HTML template
	// fmt.Println("Rendering...")
	w.WriteHeader(http.StatusOK)
	tmpl, err := template.New(html).Parse(html)
	if err != nil {
		fmt.Println(err)
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func RequireAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		allowedIps = append(allowedIps, GetLocalIP())
		allowedIps = append(allowedIps, "127.0.0.1")
		// fmt.Println("\nAllowed ips: ", allowedIps)
		// Get the IP address of the client
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// fmt.Println("\nConnecting IP: ", ip)
		// Check if the client's IP is in the list of allowed IP
		if allowAllIps {
			handler.ServeHTTP(w, r)
			return
		} else {
			for _, allowedIp := range allowedIps {
				if ip == allowedIp {
					// If the client's IP is in the list of allowed IPs, allow access to the proxy server
					handler.ServeHTTP(w, r)
					return
				}
			}
		}

		// If the client's IP is not in the list of allowed IPs, return a 403 Forbidden error
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden.\n"))
	}
}

// GetLocalIP returns the non loopback local IP of the host
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
