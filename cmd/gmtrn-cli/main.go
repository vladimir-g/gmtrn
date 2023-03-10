// Copyright 2012-2023 Vladimir Gorbunov. All rights reserved.  Use of
// this source code is governed by a MIT license that can be found in
// the LICENSE file.

/*
   Command-line interface for http://www.multitran.ru/
*/
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/vladimir-g/gmtrn"
	"os"
	"sort"
	"strings"
	"unicode/utf8"
)

var langFrom string
var langTo string
var format string
var availableLangs []string

// Set command-line flags
func init() {
	// Get languages in alphabetical order
	availableLangs = make([]string, len(gmtrn.Languages))
	var i int
	for k := range gmtrn.Languages {
		availableLangs[i] = k
		i++
	}
	sort.Strings(availableLangs)
	// Get list of available languages for usage
	usage := "Translation language. Available values:\n\t" +
		strings.Join(availableLangs, "\n\t")
	defaultFromLang := "english"
	defaultToLang := "russian"
	flag.StringVar(&langFrom, "language", defaultFromLang, usage)
	flag.StringVar(&langFrom, "l", defaultFromLang, "Same as -language")
	flag.StringVar(&langTo, "target", defaultToLang, usage)
	flag.StringVar(&langTo, "t", defaultToLang, "Same as -target")
	flag.StringVar(&format, "f", "text",
		"Output format (\"text\" or \"json\"")
}

// Usage text
func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [-l|-language source language] [-t|-target target language] query\n",
		os.Args[0])
	flag.PrintDefaults()
}

// Validate command-line arguments
func parseArgs() (query string, err error) {
	var foundFrom bool
	var foundTo bool
	for _, v := range availableLangs {
		if v == langFrom {
			foundFrom = true
		}
		if v == langTo {
			foundTo = true
		}
	}
	if !foundFrom {
		err = errors.New("Source language not available")
	}
	if !foundTo {
		err = errors.New("Target language not available")
	}
	query = strings.Join(flag.Args(), " ")
	if len(query) == 0 {
		err = errors.New("Query required")
	}
	return
}

// Get maximum length of word topic (left column)
func getMaxTopicLen(wl *[]gmtrn.WordList) int {
	var maxLen int
	var max string
	for _, wlist := range *wl {
		for _, w := range wlist.Words {
			for _, m := range w.Meanings {
				if len(m.Topic) > maxLen {
					maxLen = len(m.Topic)
					max = m.Topic
				}
			}
		}
	}
	return utf8.RuneCountInString(max)
}

// Print WordList heading
func printWordList(wlist *gmtrn.WordList, maxTopicLen int) {
	length := utf8.RuneCountInString(wlist.Query)
	fmt.Printf(" %s\n", strings.Repeat("=", length))
	fmt.Printf(" %s\n", wlist.Query)
	fmt.Printf(" %s\n", strings.Repeat("=", length))
	printWordListContents(wlist, maxTopicLen)
}

// Print WordList contents
func printWordListContents(wlist *gmtrn.WordList, maxTopicLen int) {
	for _, word := range wlist.Words {
		length := 0
		parts := [...]string{word.Pre, word.Post, word.Word,
			word.Spelling, word.Part}
		line := " "
		for _, part := range parts {
			if len(part) > 0 {
				length += utf8.RuneCountInString(part) + 1
				line += part + " "
			}
		}
		fmt.Printf(" %s\n", line)
		fmt.Printf(" %s\n", strings.Repeat("-", length + 1))
		printWord(&word, maxTopicLen)
	}
}

// Print one word
func printWord(word *gmtrn.Word, maxTopicLen int) {
	for _, m := range word.Meanings {
		fmt.Printf(" %*s ", maxTopicLen, m.Topic)
		words := make([]string, 0, len(m.Words))
		for _, w := range m.Words {
			// Add space to additional info if needed
			if len(w.Add) > 0 && !strings.HasPrefix(w.Add, " ") {
				w.Add = " " + w.Add
			}
			words = append(words,
				fmt.Sprintf("%s%s", w.Word, w.Add))
		}
		fmt.Printf("%s\n", strings.Join(words, ", "))
	}
}

// Print meaning list
func printResult(result *[]gmtrn.WordList) {
	// Get longest column for pretty printing
	maxTopicLen := getMaxTopicLen(result)
	for _, wlist := range *result {
		printWordList(&wlist, maxTopicLen)
		fmt.Printf("\n")
	}
}

func main() {
	flag.Usage = usage
	flag.Parse()
	query, err := parseArgs()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		usage()
		return
	}

	result, err := gmtrn.Query(query,
		gmtrn.Languages[langFrom],
		gmtrn.Languages[langTo])
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	switch format {
	case "text":
		printResult(&result)
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.Encode(&result)
	}
}
