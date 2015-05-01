// Copyright 2012-2015 Vladimir Gorbunov. All rights reserved.  Use of
// this source code is governed by a MIT license that can be found in
// the LICENSE file.

/* 
Command-line interface for http://www.multitran.ru/
*/
package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/vladimir-g/gmtrn"
	"os"
	"sort"
	"strings"
	"unicode/utf8"
)

var lang string
var availableLangs []string

// Set command-line flags
func init() {
	// Get languages in alphabetical order
	availableLangs = make([]string, len(gmtrn.Languages))
	var i int
	for k, _ := range gmtrn.Languages {
		availableLangs[i] = k
		i++
	}
	sort.Strings(availableLangs)
	// Get list of available languages for usage
	usage := "Translation language. Available values:\n\t" +
		strings.Join(availableLangs, "\n\t")
	defaultLang := "english"
	flag.StringVar(&lang, "language", defaultLang, usage)
	flag.StringVar(&lang, "l", defaultLang, "Same as -language")
}

// Usage text
func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [-l|-language lang] query\n",
		os.Args[0])
	flag.PrintDefaults()
}

// Validate command-line arguments
func parseArgs() (query string, err error) {
	var found bool
	for _, v := range availableLangs {
		if v == lang {
			found = true
		}
	}
	if !found {
		err = errors.New("Language not available")
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
				if len(m.Abbrev) > maxLen {
					maxLen = len(m.Abbrev)
					max = m.Abbrev
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
		partLen := 0
		if len(word.Part) > 0 {
			partLen += utf8.RuneCountInString(word.Part) + 1
		}
		length := utf8.RuneCountInString(word.Word) + partLen
		// fmt.Printf("%s\n", strings.Repeat("-", length))
		fmt.Printf(" %s %s\n", word.Word, word.Part)
		fmt.Printf(" %s\n", strings.Repeat("-", length))
		printWord(&word, maxTopicLen)
	}
}

// Print one word
func printWord(word *gmtrn.Word, maxTopicLen int) {
	for _, m := range word.Meanings {
		fmt.Printf(" %*s  ", maxTopicLen, m.Abbrev)
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
	result, err := gmtrn.Query(query, gmtrn.Languages[lang])
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	printResult(&result)
}
