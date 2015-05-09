// Copyright 2012-2015 Vladimir Gorbunov. All rights reserved.  Use of
// this source code is governed by a MIT license that can be found in
// the LICENSE file.

package gmtrn

import (
	"errors"
	"fmt"
	"html"
	"regexp"
	"strings"
)

// Single part of definition, contains word, link to word page and
// additional information.  Link may not be opened without
// "http://www.multitran.ru" referer for access.
type MeaningWord struct {
	Word,
	Link,
	Add string
}

func (w MeaningWord) String() string {
	return fmt.Sprintf(`<MeaningWord "%s" link: "%s" add: "%s">`,
		w.Word, w.Link, w.Add)
}

// Meaning of some word by topic.
type Meaning struct {
	Words []MeaningWord
	Topic,
	Link,
	Abbrev string
}

func (d Meaning) String() string {
	return fmt.Sprintf(`<Meaning, Topic: "%s", "%s", "%s" "%s">`,
		d.Topic, d.Link, d.Abbrev, d.Words)
}

// Word with list of meanings
type Word struct {
	Meanings []Meaning
	Word,
	Link,
	Part string
}

func (w Word) String() string {
	return fmt.Sprintf(`<Word "%s", "%s", "%s", "%s">`,
		w.Word, w.Link, w.Part, w.Meanings)
}

// List of words for requested page
type WordList struct {
	Words []Word
	Query,
	Link string
}

func (w WordList) String() string {
	return fmt.Sprintf(`<WordList "%s", "%s": "%s">`,
		w.Query, w.Link, w.Words)
}

// Link to page
type link struct {
	word,
	link string
}

// Multitran base domain with search path
var domain = "http://www.multitran.ru/c/"

/* Regexes */
// Links to other words
var otherRe = regexp.MustCompile(
	`(?s)height="20">&nbsp;(.*)&nbsp;&nbsp;\r\n<br>&nbsp;найдены`)
var linkRe = regexp.MustCompile(`(?s)href="([^"]+)">([^<]+)`)

// Main content
var blockRe = regexp.MustCompile("(?s)<table(.*)</table><a name=\"phrases\">")

// Word
var wordRe = regexp.MustCompile(`^<a href="([^"]+)">(.+?)</a>.*( <em>|&nbsp;)?`)

// Part of speech (if exists)
var partRe = regexp.MustCompile(` <em>([^<]+)</em>`)

// Meaning
var itemsRe = regexp.MustCompile(
	`(?s)<a title="(.*?)" href="([^"]+)"><i>(.*?)</i>&nbsp;` +
		`</a>\r\n</td><td>(.*?)(<span STYLE="color:black"></td>` +
		`|<td >&nbsp;&nbsp;|</a></td></tr>|<tr>)`)

// MeaningWord
var mwordRe = regexp.MustCompile(`<a href="([^"]+)">([^<]+)`)

// Additional word info
var addRe = regexp.MustCompile(`<span STYLE="color:gray">(.*)`)

// Strip tags
var stripRe = regexp.MustCompile("<(.*?)>")

// Remove tags from string
func stripTags(str string) string {
	return stripRe.ReplaceAllString(str, "")
}

// Remove &nbsp; from string
func removeNbsp(str string) string {
	return strings.Replace(str, "&nbsp;", "", -1)
}

// Get list of word meanings by topic
func parseWord(chunk string) []Meaning {
	items := itemsRe.FindAllStringSubmatch(chunk, -1)
	result := make([]Meaning, 0, len(items))
	// Process multiple words in meaning
	for _, item := range items {
		if len(item) != 6 {
			continue
		}
		cleanStr := html.UnescapeString(removeNbsp(item[4]))
		words := strings.Split(cleanStr, ";")
		mwords := make([]MeaningWord, 0, len(words))
		for _, v := range words {
			word := mwordRe.FindStringSubmatch(v)
			if len(word) != 3 {
				continue
			}
			mw := MeaningWord{word[2], domain + word[1], ""}
			// Get additional information if exist
			if add := addRe.FindStringSubmatch(v); len(add) == 2 {
				mw.Add = stripRe.ReplaceAllString(add[1], "")
			}
			mwords = append(mwords, mw)
		}
		m := Meaning{mwords,
			html.UnescapeString(item[1]),
			domain + item[2],
			removeNbsp(item[3])}
		result = append(result, m)
	}
	return result
}

// Parse page content and return word meanings
func parsePage(content string) (result []Word, err error) {
	// Get main block with content
	block := blockRe.FindStringSubmatch(content)
	if len(block) != 2 {
		err = errors.New("Translation can not be found")
		return
	}
	// Split content to chunks with word
	chunks := strings.Split(block[1],
		`<td bgcolor="#DBDBDB" colspan="2" width="700">&nbsp;`)
	result = make([]Word, 0, len(chunks))
	for _, chunk := range chunks {
		matches := wordRe.FindStringSubmatch(chunk)
		if len(matches) == 4 {
			meanings := parseWord(chunk)
			part := ""
			partMatches := partRe.FindStringSubmatch(chunk)
			if len(partMatches) == 2 {
				part = partMatches[1]
			}
			word := Word{meanings,
				stripTags(html.UnescapeString(matches[2])),
				domain + matches[1],
				part}
			result = append(result, word)
		}
	}
	return
}
