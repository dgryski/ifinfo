// ifinfo: a clone of the fabulous ifconfig.me
package main

import (
	"encoding/json"
	"encoding/xml"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

type info struct {
	Connection string `xml:"connection" json:"connection"`
	Encoding   string `xml:"encoding" json:"encoding"`
	Forwarded  string `xml:"forwarded" json:"forwarded"`
	IpAddr     string `xml:"ip_addr" json:"ip_addr"`
	Host       string `xml:"-" json:"-"`
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
	if fwds, ok := r.Header["X-Forwarded-For"]; ok {
		inf.Forwarded = strings.Join(fwds, ",")
	}
	inf.Host = r.Host
	inf.Lang = maybeGet(r.Header, "Accept-Language")
	inf.Mime = maybeGet(r.Header, "Accept")
	inf.UserAgent = maybeGet(r.Header, "User-Agent")
	inf.Via = maybeGet(r.Header, "Via")

	inf.IpAddr = maybeGet(r.Header, "X-Real-Ip")

	if inf.IpAddr == "" {
		inf.IpAddr = maybeGet(r.Header, "X-Forwarded-For")
	}

	if inf.IpAddr == "" {
		inf.IpAddr, _, _ = net.SplitHostPort(r.RemoteAddr)
	}

	hosts, err := net.LookupAddr(inf.IpAddr)
	if err == nil {
		inf.RemoteHost = hosts[0]
	} else {
		inf.RemoteHost = inf.IpAddr
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

	http.HandleFunc("/connection", func(w http.ResponseWriter, r *http.Request) {
		m := makeInfo(r)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(m.Connection))
		w.Write([]byte("\n"))
	})

	http.HandleFunc("/encoding", func(w http.ResponseWriter, r *http.Request) {
		m := makeInfo(r)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(m.Encoding))
		w.Write([]byte("\n"))
	})

	http.HandleFunc("/forwarded", func(w http.ResponseWriter, r *http.Request) {
		m := makeInfo(r)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(m.Forwarded))
		w.Write([]byte("\n"))
	})

	http.HandleFunc("/ip", func(w http.ResponseWriter, r *http.Request) {
		m := makeInfo(r)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(m.IpAddr))
		w.Write([]byte("\n"))
	})

	http.HandleFunc("/lang", func(w http.ResponseWriter, r *http.Request) {
		m := makeInfo(r)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(m.Lang))
		w.Write([]byte("\n"))
	})

	http.HandleFunc("/mime", func(w http.ResponseWriter, r *http.Request) {
		m := makeInfo(r)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(m.Mime))
		w.Write([]byte("\n"))
	})

	http.HandleFunc("/host", func(w http.ResponseWriter, r *http.Request) {
		m := makeInfo(r)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(m.RemoteHost))
		w.Write([]byte("\n"))
	})

	http.HandleFunc("/ua", func(w http.ResponseWriter, r *http.Request) {
		m := makeInfo(r)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(m.UserAgent))
		w.Write([]byte("\n"))
	})

	http.HandleFunc("/via", func(w http.ResponseWriter, r *http.Request) {
		m := makeInfo(r)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(m.Via))
		w.Write([]byte("\n"))
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		m := makeInfo(r)
		w.Header().Set("Cache-Control", "no-cache")
		if !strings.Contains(m.UserAgent, "Mozilla") {
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(m.IpAddr))
			w.Write([]byte("\n"))
			return
		}

		rootTemplate.Execute(w, m)

	})

	port := ":8080"

	if p := os.Getenv("PORT"); p != "" {
		port = ":" + p
	}

	log.Fatal(http.ListenAndServe(port, nil))
}

var rootTemplate = template.Must(template.New("root").Parse(rootTemplateHTML))

const rootTemplateHTML = `
<html>
  <head>
  <style type="text/css">

    @import url(//fonts.googleapis.com/css?family=Droid+Serif);
    @import url(//fonts.googleapis.com/css?family=Droid+Sans+Mono);

    body {
       background : lightgrey ;
       margin-top : 100px ;
       font-family : 'Droid Serif' ;
    }

    div#content
    {
       margin : auto ;
       width : 75%;
    }

    table {
       background : white ;
       border-style : solid ;
       border-collapse : collapse ;
       border-color : grey ;
    }

    td.value {
       font-family : 'Droid Sans Mono' ;
       font-size : 12px
    }

    td.key {
       font-size : 15px ;
    }

</style>

  <body>
    <div id="content">
    <p>Your connection info</p>
        <table border=1>
        <tr><td class="key">Ip Address</td><td style="font-size : 30px" class="value">{{.IpAddr}}</td></tr>
            <tr><td class="key">Connection</td><td class="value">{{.Connection}}</td></tr>
            <tr><td class="key">Encoding</td><td class="value">{{.Encoding}}</td></tr>
            <tr><td class="key">Forwarded</td><td class="value">{{.Forwarded}}</td></tr>
            <tr><td class="key">Language</td><td class="value">{{.Lang}}</td></tr>
            <tr><td class="key">MIME Type</td><td class="value">{{.Mime}}</td></tr>
            <tr><td class="key">Remote Host</td><td class="value">{{.RemoteHost}}</td></tr>
            <tr><td class="key">User Agent</td><td class="value">{{.UserAgent}}</td></tr>
            <tr><td class="key">Via</td><td class="value">{{.Via}}</td></tr>
        </table>

        <div style="font-size: 12px">

        <p>API

        <ul>
        <li><a href="/all.json">/all.json</a>
        <li><a href="/all.xml">/all.xml</a>
        <li>command line:
            <ul>
            <li>curl http://{{.Host}} ⇒ {{ .IpAddr }}
            <li>curl http://{{.Host}}/ip ⇒ {{ .IpAddr }}
            <li>curl http://{{.Host}}/host ⇒ {{ .RemoteHost }}
            <li>curl http://{{.Host}}/connection ⇒ {{ .Connection }}
            <li>curl http://{{.Host}}/encoding ⇒ {{ .Encoding }}
            <li>curl http://{{.Host}}/forwarded ⇒ {{ .Forwarded }}
            <li>curl http://{{.Host}}/lang ⇒ {{ .Lang }}
            <li>curl http://{{.Host}}/mime ⇒ {{ .Mime }}
            <li>curl http://{{.Host}}/ua ⇒ {{ .UserAgent }}
            <li>curl http://{{.Host}}/via ⇒ {{ .Via }}
            </ul>
        </ul>
        </div>
    </div>
  </body>
</html>
`
