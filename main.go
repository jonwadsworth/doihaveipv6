package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

//go:embed public/style.css
var cssContent string

func clientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func hasIPv6(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	// IPv4-mapped addresses (::ffff:x.x.x.x) are not real IPv6
	return parsed.To4() == nil && parsed.To16() != nil
}

func handleAPI(w http.ResponseWriter, r *http.Request) {
	ip := clientIP(r)
	v6 := hasIPv6(ip)
	ipType := "IPv4"
	if v6 {
		ipType = "IPv6"
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"hasIpv6": v6,
		"ip":      ip,
		"type":    ipType,
	})
}

func handleCSS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	fmt.Fprint(w, cssContent)
}

func handlePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	ip := clientIP(r)
	v6 := hasIPv6(ip)
	safeIP := html.EscapeString(ip)

	verdict, verdictClass, ipLabel, ipColor, ipTypeText := "NO", "verdict-no", "Your IPv4 address", "#ffab40", "IPv4"
	extraNote := `<p class="no-explainer">Your device connected over IPv4 — either your ISP doesn't offer IPv6 yet, or it's disabled on your network.</p>`

	if v6 {
		verdict, verdictClass, ipLabel, ipColor, ipTypeText = "YES", "verdict-yes", "Your IPv6 address", "#40c4ff", "IPv6"
		extraNote = ""
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Do I Have IPv6?</title>
  <link rel="stylesheet" href="/style.css">
</head>
<body>
  <main>
    <h1 class="site-title">Do I Have IPv6?</h1>
    <div class="verdict %s">%s</div>
    <div class="ip-info">
      <p class="ip-label">%s</p>
      <p class="ip-value" style="color:%s">%s</p>
      %s
    </div>
    <div class="explainer">
      <p>IPv6 is the modern internet protocol that provides virtually unlimited addresses.
         Most major ISPs and networks now support it.</p>
      <p>You connected from <strong>%s</strong>, which is a <strong>%s</strong> address.</p>
    </div>
    <footer>
      <a href="/api">JSON API</a> &mdash; <a href="https://github.com/jonwadsworth/doihaveipv6">source</a>
    </footer>
  </main>
</body>
</html>`, verdictClass, verdict, ipLabel, ipColor, safeIP, extraNote, safeIP, ipTypeText)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	http.HandleFunc("/api", handleAPI)
	http.HandleFunc("/style.css", handleCSS)
	http.HandleFunc("/", handlePage)

	log.Printf("doihaveipv6 listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
