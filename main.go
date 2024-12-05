package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

// SecretSanta stores the givers and their giftees
type SecretSanta struct {
	Pairings map[string]string
}

// readNamesFromFile reads the student names from a file
func readNamesFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var names []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		name := strings.TrimSpace(scanner.Text())
		if name != "" {
			names = append(names, name)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return names, nil
}

// makePairings generates Secret Santa pairings
func makePairings(names []string) (map[string]string, error) {
	if len(names) < 2 {
		return nil, errors.New("not enough participants for Secret Santa")
	}

	rand.Seed(time.Now().UnixNano())
	shuffled := append([]string(nil), names...)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	pairings := make(map[string]string)
	for i, giver := range names {
		recipient := shuffled[i]
		if giver == recipient { // Retry shuffle if a giver matches their recipient
			return makePairings(names)
		}
		pairings[giver] = recipient
	}

	return pairings, nil
}

// queryHandler handles requests to find the giftee for a given name
func queryHandler(santa *SecretSanta) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if name == "" {
			http.Error(w, "Name parameter is required", http.StatusBadRequest)
			return
		}

		giftee, ok := santa.Pairings[name]
		if !ok {
			http.Error(w, "Name not found in Secret Santa list", http.StatusNotFound)
			return
		}

		response := map[string]string{"giver": name, "giftee": giftee}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// serveStaticHTML serves the HTML form to users
func serveStaticHTML(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/index.html")
}

func main() {
	names, err := readNamesFromFile("students.txt")
	if err != nil {
		fmt.Println("Error reading students file:", err)
		return
	}

	pairings, err := makePairings(names)
	if err != nil {
		fmt.Println("Error generating pairings:", err)
		return
	}

	santa := &SecretSanta{Pairings: pairings}

	http.HandleFunc("/", serveStaticHTML)          // Serve the HTML form
	http.HandleFunc("/query", queryHandler(santa)) // Handle Secret Santa lookups

	fmt.Println("Server running at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
