// Copyright 2012-2015 Vladimir Gorbunov. All rights reserved.  Use of
// this source code is governed by a MIT license that can be found in
// the LICENSE file.

package gmtrn

import (
	"fmt"
	"log"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"net/url"
	"strconv"
)

// Available languages
var Languages = map[string]int{
	"english":   1,
	"russian":   2,
	"german":    3,
	"french":    4,
	"spanish":   5,
	"italian":   23,
	"dutch":     24,
	"latvian":   27,
	"estonian":  26,
	"japanese":  28,
	"afrikaans": 31,
	"esperanto": 34,
	"kalmyk":    35,
}

// Get parsed data and links array from url
func getData(url string) (result WordList, links []link, err error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	// Parser doesn't work with mobile version
	req.Header.Set("User-Agent", "Mozilla/5.0 Firefox/75.0")
        resp, err := client.Do(req)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Url %s returned HTTP code: %d",
			url, resp.StatusCode)
		return
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}

	links, err = parseLinks(doc)
	if err != nil {
		return
	}
	result, err = parsePage(doc)
	return
}

// Create search query
func getQuery(query string, langFrom int, langTo int) (result string) {
	values := url.Values{}
	values.Add("l1", strconv.Itoa(langFrom))
	values.Add("l2", strconv.Itoa(langTo))
	values.Add("s", query)
	result = domain + "m.exe?" + values.Encode()
	return
}

// Run HTTP query to http://www.multitran.ru and return parsed results
// or error. Function could return error if translation isn't found or
// something wrong happens.
//
// query - query string
// langFrom - integer from Languages map, source language
// langTo - integer from Languages map, target language
func Query(query string, langFrom int, langTo int) (result []WordList, err error) {
	queryUrl := getQuery(query, langFrom, langTo)
	wordList, links, err := getData(queryUrl)
	if err != nil {
		return
	}
	wordList.Link = queryUrl
	// Assign word to wordlist in case of single page
	if len(links) == 0 {
		wordList.Query = query
	} else {
		wordList.Query = links[0].word
	}
	result = append(make([]WordList, 0), wordList)
	if len(links) == 0 {
		return
	}
	// Process other pages, starting from second link
	for i := 1; i < len(links); i++ {
		wordList, _, oerr := getData(links[i].link)
		if oerr != nil {
			log.Print("Error when getting %s: %s\n",
				links[i], oerr)
		} else {
			wordList.Query = links[i].word
			wordList.Link = links[i].link
			result = append(result, wordList)
		}
	}
	return
}
