package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/sahilm/fuzzy"
)

//go:embed templates/*
var templatesFS embed.FS

type Repository struct {
	NameWithOwner  string    `json:"name_with_owner"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Language       string    `json:"language"`
	URL            string    `json:"url"`
	StargazerCount int       `json:"stargazer_count"`
	ForkCount      int       `json:"fork_count"`
	PushedAt       time.Time `json:"pushed_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	StarredAt      time.Time `json:"starred_at"`
	IsPrivate      bool      `json:"is_private"`
	Topics         []string  `json:"topics"`
}

type Server struct {
	repos     []Repository
	languages []string
	tmpl      *template.Template
}

type SearchResult struct {
	Repos      []Repository
	Query      string
	Language   string
	Sort       string
	TotalCount int
}

func (s *Server) loadData(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}
	if err := json.Unmarshal(data, &s.repos); err != nil {
		return fmt.Errorf("parse json: %w", err)
	}

	// Extract unique languages
	langSet := make(map[string]struct{})
	for _, r := range s.repos {
		if r.Language != "" {
			langSet[r.Language] = struct{}{}
		}
	}
	for lang := range langSet {
		s.languages = append(s.languages, lang)
	}
	sort.Strings(s.languages)

	return nil
}

// Implement fuzzy.Source interface
type repoSource []Repository

func (r repoSource) String(i int) string {
	repo := r[i]
	return repo.NameWithOwner + " " + repo.Description + " " + strings.Join(repo.Topics, " ")
}

func (r repoSource) Len() int { return len(r) }

func (s *Server) search(query, language, sortBy string) SearchResult {
	filtered := s.repos

	// Filter by language
	if language != "" {
		var langFiltered []Repository
		for _, r := range filtered {
			if r.Language == language {
				langFiltered = append(langFiltered, r)
			}
		}
		filtered = langFiltered
	}

	// Fuzzy search if query provided
	if query != "" {
		matches := fuzzy.FindFrom(query, repoSource(filtered))
		matched := make([]Repository, len(matches))
		for i, m := range matches {
			matched[i] = filtered[m.Index]
		}
		filtered = matched
	}

	// Sort results (skip if fuzzy search, already ranked by relevance)
	if query == "" {
		switch sortBy {
		case "stars":
			sort.Slice(filtered, func(i, j int) bool {
				return filtered[i].StargazerCount > filtered[j].StargazerCount
			})
		case "pushed":
			sort.Slice(filtered, func(i, j int) bool {
				return filtered[i].PushedAt.After(filtered[j].PushedAt)
			})
		case "starred":
			sort.Slice(filtered, func(i, j int) bool {
				return filtered[i].StarredAt.After(filtered[j].StarredAt)
			})
		case "name":
			sort.Slice(filtered, func(i, j int) bool {
				return filtered[i].NameWithOwner < filtered[j].NameWithOwner
			})
		default: // default: recently starred
			sort.Slice(filtered, func(i, j int) bool {
				return filtered[i].StarredAt.After(filtered[j].StarredAt)
			})
		}
	}

	return SearchResult{
		Repos:      filtered,
		Query:      query,
		Language:   language,
		Sort:       sortBy,
		TotalCount: len(filtered),
	}
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Languages []string
		Results   SearchResult
	}{
		Languages: s.languages,
		Results:   s.search("", "", "starred"),
	}
	s.tmpl.ExecuteTemplate(w, "index.html", data)
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	language := r.URL.Query().Get("lang")
	sortBy := r.URL.Query().Get("sort")

	results := s.search(query, language, sortBy)
	s.tmpl.ExecuteTemplate(w, "results.html", results)
}

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	dataFile := flag.String("data", "data/stars.json", "path to stars.json")
	flag.Parse()

	funcMap := template.FuncMap{
		"timeAgo": func(t time.Time) string {
			d := time.Since(t)
			switch {
			case d < time.Hour:
				return fmt.Sprintf("%dm ago", int(d.Minutes()))
			case d < 24*time.Hour:
				return fmt.Sprintf("%dh ago", int(d.Hours()))
			case d < 30*24*time.Hour:
				return fmt.Sprintf("%dd ago", int(d.Hours()/24))
			case d < 365*24*time.Hour:
				return fmt.Sprintf("%dmo ago", int(d.Hours()/(24*30)))
			default:
				return fmt.Sprintf("%dy ago", int(d.Hours()/(24*365)))
			}
		},
		"formatNum": func(n int) string {
			if n >= 1000 {
				return fmt.Sprintf("%.1fk", float64(n)/1000)
			}
			return fmt.Sprintf("%d", n)
		},
		"langColor": func(lang string) string {
			colors := map[string]string{
				"Go":         "#00ADD8",
				"Python":     "#3572A5",
				"JavaScript": "#f1e05a",
				"TypeScript": "#3178c6",
				"Rust":       "#dea584",
				"Ruby":       "#701516",
				"Java":       "#b07219",
				"C":          "#555555",
				"C++":        "#f34b7d",
				"C#":         "#178600",
				"Swift":      "#F05138",
				"Kotlin":     "#A97BFF",
				"Shell":      "#89e051",
				"HTML":       "#e34c26",
				"CSS":        "#563d7c",
				"Vim Script": "#199f4b",
				"Lua":        "#000080",
				"Zig":        "#ec915c",
			}
			if c, ok := colors[lang]; ok {
				return c
			}
			return "#8b949e"
		},
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		log.Fatalf("parse templates: %v", err)
	}

	srv := &Server{tmpl: tmpl}
	if err := srv.loadData(*dataFile); err != nil {
		log.Fatalf("load data: %v", err)
	}

	log.Printf("Loaded %d repos, %d languages", len(srv.repos), len(srv.languages))

	http.HandleFunc("/", srv.handleIndex)
	http.HandleFunc("/search", srv.handleSearch)
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		data, _ := templatesFS.ReadFile("templates/favicon.ico")
		w.Header().Set("Content-Type", "image/x-icon")
		w.Write(data)
	})

	log.Printf("Listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
