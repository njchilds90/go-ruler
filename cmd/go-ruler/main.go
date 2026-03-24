package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	ruler "github.com/njchilds90/go-ruler"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "eval":
		must(evalCmd(os.Args[2:]))
	case "load":
		must(loadCmd(os.Args[2:]))
	case "serve":
		must(serveCmd(os.Args[2:]))
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Println("go-ruler <eval|load|serve> <rules.json> [facts.json|:8080]")
}

func evalCmd(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: go-ruler eval <rules.json> <facts.json>")
	}
	engine, err := loadEngine(args[0])
	if err != nil {
		return err
	}
	facts, err := loadFacts(args[1])
	if err != nil {
		return err
	}
	decision, err := engine.EvaluateDecision(context.Background(), facts)
	if err != nil {
		return err
	}
	return writeJSON(os.Stdout, decision)
}

func loadCmd(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: go-ruler load <rules.json>")
	}
	rules, err := ruler.LoadRulesFile(args[0])
	if err != nil {
		return err
	}
	return writeJSON(os.Stdout, map[string]any{"loaded": len(rules), "rule_names": names(rules)})
}

func serveCmd(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: go-ruler serve <rules.json> <addr>")
	}
	engine, err := loadEngine(args[0])
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("ok")) })
	mux.HandleFunc("/eval", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var facts ruler.FactMap
		if err := json.NewDecoder(r.Body).Decode(&facts); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = writeJSON(w, map[string]any{"error": err.Error()})
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		decision, err := engine.EvaluateDecision(ctx, facts)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = writeJSON(w, map[string]any{"error": err.Error()})
			return
		}
		_ = writeJSON(w, decision)
	})
	log.Printf("go-ruler serve listening on %s", args[1])
	return http.ListenAndServe(args[1], mux)
}

func loadEngine(path string) (*ruler.Engine, error) {
	rules, err := ruler.LoadRulesFile(path)
	if err != nil {
		return nil, err
	}
	engine, err := ruler.NewFromRules(rules, ruler.WithCache())
	if err != nil {
		return nil, err
	}
	engine.Freeze()
	return engine, nil
}

func loadFacts(path string) (ruler.FactMap, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var facts ruler.FactMap
	if err := json.Unmarshal(b, &facts); err != nil {
		return nil, err
	}
	return facts, nil
}

func writeJSON(w any, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	switch out := w.(type) {
	case *os.File:
		_, err = out.Write(append(b, '\n'))
		return err
	case http.ResponseWriter:
		out.Header().Set("Content-Type", "application/json")
		_, err = out.Write(append(b, '\n'))
		return err
	default:
		return fmt.Errorf("unsupported writer %T", w)
	}
}

func names(rules []ruler.Rule) []string {
	out := make([]string, 0, len(rules))
	for _, r := range rules {
		out = append(out, r.Name)
	}
	return out
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
