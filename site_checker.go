package main

import (
	"bufio"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strings"
	"time"
)

var statusMap = map[int]string{
	200: "✅ OK",
	301: "➡️  Redirection permanente",
	302: "↪️  Redirection temporaire",
	400: "🧨 Mauvaise requête",
	401: "🔒 Non autorisé",
	403: "⛔ Accès interdit 🔒 (site ok mais bloqué par cloudflare)",
	404: "🔍 Introuvable",
	408: "⏳ Timeout",
	429: "🚫 Trop de requêtes",
	451: "🛑 Problème légal RIP",
	500: "💀 Erreur interne serveur",
	502: "🚧 Bad Gateway",
	503: "🔧 Service indisponible",
	504: "⏱️ Gateway timeout",
}

func readSitesFromFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var sites []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			sites = append(sites, line)
		}
	}

	return sites, scanner.Err()
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./checksites <fichier.txt>")
		return
	}

	filePath := os.Args[1]
	sites, err := readSitesFromFile(filePath)
	if err != nil {
		fmt.Printf("Erreur lecture fichier : %s\n", err)
		return
	}

	jar, _ := cookiejar.New(nil)

	client := http.Client{
		Timeout: 6 * time.Second,
		Jar:     jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > 0 {
				req.Header = via[0].Header
			}
			return nil
		},
	}

	for _, site := range sites {
		req, err := http.NewRequest("GET", site, nil)
		if err != nil {
			fmt.Printf("❌ %s → Erreur création requête : %s\n", site, err)
			continue
		}

		req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36")
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		req.Header.Set("Accept-Language", "fr-FR,fr;q=0.9,en-US;q=0.8,en;q=0.7")
		req.Header.Set("Connection", "keep-alive")

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("❌ %s → Erreur de connexion : %s\n", site, err)
			continue
		}

		defer resp.Body.Close()

		msg, found := statusMap[resp.StatusCode]
		if found {
			fmt.Printf("%s → %d %s\n", site, resp.StatusCode, msg)
		} else if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			fmt.Printf("%s → %d ✅ Réussi (non mappé)\n", site, resp.StatusCode)
		} else if resp.StatusCode >= 300 && resp.StatusCode < 400 {
			fmt.Printf("%s → %d 🔁 Redirection (non mappée)\n", site, resp.StatusCode)
		} else if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			fmt.Printf("%s → %d ⚠️ Erreur client (non mappée)\n", site, resp.StatusCode)
		} else if resp.StatusCode >= 500 {
			fmt.Printf("%s → %d 💥 Erreur serveur (non mappée)\n", site, resp.StatusCode)
		} else {
			fmt.Printf("%s → %d ❓ Code inconnu\n", site, resp.StatusCode)
		}
	}
}
