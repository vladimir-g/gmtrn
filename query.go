// Copyright 2012-2023 Vladimir Gorbunov. All rights reserved. Use of
// this source code is governed by a MIT license that can be found in
// the LICENSE file.

package gmtrn

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// Available languages
var Languages = map[string]int{
	"english":    1,
	"russian":    2,
	"german":     3,
	"french":     4,
	"spanish":    5,
	"croatian":   8,
	"arabic":     10,
	"portuguese": 11,
	"lithuanian": 12,
	"romanian":   13,
	"polish":     14,
	"bulgarian":  15,
	"czech":      16,
	"chinese":    17,
	"danish":     22,
	"italian":    23,
	"dutch":      24,
	"latvian":    27,
	"estonian":   26,
	"japanese":   28,
	"swedish":    29,
	"norwegian":  30,
	"afrikaans":  31,
	"turkish":    32,
	"ukrainian":  33,
	"esperanto":  34,
	"kalmyk":     35,
	"finnish":    36,
	"latin":      37,
	"greek":      38,
	"korean":     39,
	"hungarian":  42,
	"irish":      49,
	"slovak":     60,
	"slovene":    67,
	"maltese":    78,
}

// Get parsed data and links array from url
func getData(url string, result *WordList, links *[]link) (err error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
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

	err = parsePage(resp.Body, result, links)

	if err != nil {
		return
	}

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

// Run HTTP query to http://www.multitran.com and return parsed
// results or error. Function may return error if translation isn't
// found or something wrong happens.
//
// query - query string
// langFrom - integer from Languages map, source language
// langTo - integer from Languages map, target language
func Query(query string, langFrom int, langTo int) (result []WordList, err error) {
	queryUrl := getQuery(query, langFrom, langTo)
	wordList := WordList{}
	links := make([]link, 0)
	err = getData(queryUrl, &wordList, &links)
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
		wordList = WordList{}
		oerr := getData(links[i].link, &wordList, &links)
		if oerr != nil {
			fmt.Printf("Error when getting %s: %s\n",
				links[i], oerr)
			return
		} else {
			wordList.Query = links[i].word
			wordList.Link = links[i].link
			result = append(result, wordList)
		}
	}
	return
}
