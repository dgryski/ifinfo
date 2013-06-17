// ifinfo: a clone of the fabulous ifconfig.me
package main

// TODO: html if useragent =~ /[Mm]ozilla/ (== browser)

import (
	"encoding/json"
	"encoding/xml"
	"log"
	"net"
	"net/http"
	"os"
)

type info struct {
	Connection string `xml:"connection" json:"connection"`
	Encoding   string `xml:"encoding" json:"encoding"`
	Forwarded  string `xml:"forwarded" json:"forwarded"`
	IpAddr     string `xml:"ip_addr" json:"ip_addr"`
	Lang       string `xml:"lang" json:"lang"`
	Mime       string `xml:"mime" json:"mime"`
	RemoteHost string `xml:"remote_host" json:"remote_host"`
	UserAgent  string `xml:"user_agent" json:"user_agent"`
	Via        string `xml:"via" json:"via"`
}

func maybeGet(h map[string][]string, key string) string {

	if v, ok := h[key]; ok {
		return v[0]
	}

	return ""
}

func makeInfo(r *http.Request) *info {

	inf := &info{}

	inf.Connection = maybeGet(r.Header, "Connection")
	inf.Encoding = maybeGet(r.Header, "Accept-Encoding")
	inf.Forwarded = maybeGet(r.Header, "X-Forwarded-For")
	inf.Lang = maybeGet(r.Header, "Accept-Language")
	inf.Mime = maybeGet(r.Header, "Accept")
	inf.UserAgent = maybeGet(r.Header, "User-Agent")
	inf.Via = maybeGet(r.Header, "Via")

	ipAddr := maybeGet(r.Header, "X-Real-Ip")

	if ipAddr == "" {
		ipAddr = inf.Forwarded
	}

	if ipAddr == "" {
		ipAddr, _, _ = net.SplitHostPort(r.RemoteAddr)
	}

	inf.IpAddr = ipAddr

	hosts, err := net.LookupAddr(ipAddr)
	if err == nil {
		inf.RemoteHost = hosts[0]
	} else {
		inf.RemoteHost = ipAddr
	}

	return inf
}

func main() {

	http.HandleFunc("/all.json", func(w http.ResponseWriter, r *http.Request) {
		m := makeInfo(r)
		s, _ := json.Marshal(m)
		w.Header().Set("Content-Type", "application/json")
		w.Write(s)
		w.Write([]byte("\n"))
	})

	http.HandleFunc("/all.xml", func(w http.ResponseWriter, r *http.Request) {
		m := makeInfo(r)
		s, _ := xml.MarshalIndent(m, "", "  ")
		w.Header().Set("Content-Type", "text/xml")
		w.Write(s)
		w.Write([]byte("\n"))
	})

	http.HandleFunc("/ip", func(w http.ResponseWriter, r *http.Request) {
		m := makeInfo(r)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(m.IpAddr))
		w.Write([]byte("\n"))
	})

	http.HandleFunc("/host", func(w http.ResponseWriter, r *http.Request) {
		m := makeInfo(r)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(m.RemoteHost))
		w.Write([]byte("\n"))
	})

	port := ":8080"

	for _, e := range []string{
		"VCAP_APP_PORT", // cloudfoundry / appfog
		"PORT",          // heroku
	} {
		if p := os.Getenv(e); p != "" {
			port = ":" + p
			break
		}
	}

	log.Fatal(http.ListenAndServe(port, nil))
}
