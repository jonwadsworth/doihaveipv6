package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

//go:embed public/style.css
var cssContent string

//go:embed public/index.html
var indexHTMLTemplate string

var geoClient = &http.Client{Timeout: 3 * time.Second}

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
	return parsed.To4() == nil && parsed.To16() != nil
}

func handleAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "https://doihaveipv6.com")
	w.Header().Set("Access-Control-Allow-Methods", "GET")

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

func handleGeo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "https://doihaveipv6.com")

	ip := r.URL.Query().Get("ip")
	if net.ParseIP(ip) == nil {
		http.Error(w, "invalid ip", http.StatusBadRequest)
		return
	}

	resp, err := geoClient.Get("https://ipwho.is/" + url.PathEscape(ip))
	if err != nil {
		http.Error(w, "lookup failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	var data struct {
		City    string `json:"city"`
		Country string `json:"country"`
		Success bool   `json:"success"`
		Conn    struct {
			ISP string `json:"isp"`
		} `json:"connection"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil || !data.Success {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, "{}")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	json.NewEncoder(w).Encode(map[string]string{
		"isp":     data.Conn.ISP,
		"city":    data.City,
		"country": data.Country,
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
	v6JS := "false"
	if v6 {
		v6JS = "true"
	}

	fallback := fmt.Sprintf(`<script>window.__sip=%q;window.__sv6=%s;</script>`, ip, v6JS)
	page := strings.Replace(indexHTMLTemplate, "__SERVER_FALLBACK__", fallback, 1)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, page)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	http.HandleFunc("/api", handleAPI)
	http.HandleFunc("/geo", handleGeo)
	http.HandleFunc("/style.css", handleCSS)
	http.HandleFunc("/", handlePage)

	log.Printf("doihaveipv6 listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
