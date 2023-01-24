package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
)

var words []string
var word string
var wordMask string
var wrongGuesses int

func init() {
	wordsFile, err := os.Open("words.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer wordsFile.Close()

	scanner := bufio.NewScanner(wordsFile)
	var words []string
	for scanner.Scan() {
		words = append(words, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading words file:", err)
		os.Exit(1)
	}

	word = words[0]
	wordMask = strings.Repeat("_", len(word))
	wrongGuesses = 0
}

func main() {
	http.HandleFunc("/", mainPage)
	http.HandleFunc("/hangman", hangman)
    http.Handle("/main.css", http.FileServer(http.Dir("."))) 
	http.ListenAndServe(":8080", nil)
}

func mainPage(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("main.html")
	t.Execute(w, map[string]interface{}{
		"word":         wordMask,
		"wrongGuesses": wrongGuesses,
	})
}

const maxWrongGuesses = 10

func hangman(w http.ResponseWriter, r *http.Request) {
	var index int
	var words []string
	var word, wordMask string
	var wrongGuesses int
	var maxWrongGuesses int = 10

	// Read words from file
	wordsFile, err := os.Open("words.txt")
	if err != nil {
		log.Println(err)
	}
	defer wordsFile.Close()

	scanner := bufio.NewScanner(wordsFile)
	for scanner.Scan() {
		words = append(words, scanner.Text())
	}
	word = words[index]
	wordMask = strings.Repeat("_", len(word))

	if r.Method == "POST" {
		if wrongGuesses >= maxWrongGuesses {
			http.Error(w, "Too many wrong guesses, the game is over.", http.StatusForbidden)
			return
		}

		var requestData struct {
			Letter string `json:"letter"`
		}
		err := json.NewDecoder(r.Body).Decode(&requestData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		letter := requestData.Letter

		if !strings.Contains(word, letter) {
			wrongGuesses++
		} else {
			for i := 0; i < len(word); i++ {
				if string(word[i]) == letter {
					wordMask = wordMask[:i] + letter + wordMask[i+1:]
				}
			}
		}
	}

	if word == wordMask {
		// read the next word and reset the wrong guesses count
		index++
		if index == len(words) {
			index = 0
		}
		word = words[index]
		wordMask = strings.Repeat("_", len(word))
		wrongGuesses = 0
	}

	if wrongGuesses >= maxWrongGuesses {
		responseData := map[string]interface{}{
			"word":         word,
			"wrongGuesses": wrongGuesses,
			"gameOver":     true,
		}
		json.NewEncoder(w).Encode(responseData)
		return
	}

	responseData := map[string]interface{}{
		"word":         wordMask,
		"wrongGuesses": wrongGuesses,
		"gameOver":     false,
	}
	json.NewEncoder(w).Encode(responseData)
}
