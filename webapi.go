/*
Package gmtrn implements http client library for http://www.multitran.ru/

DISCLAIMER: Yes, I know that usage of regexes for html parsing is a
bad practice, but site's markup is very poor-formed and other parsing
methods are too complex in this case.

Usage:
	result, err := webapi.Query("Query string",
				    webapi.Languages["english"])

Known issues:

- There are some problems with translation to Kalmyk language but
reverse translation works fine. This problem happens because site
uses wrong guessing algorithm for determining the source language.

- Only default language for site interface is implemented.

- There is no tests.

How multitran works:

Site splits incoming query to multiple parts and displays results for
first part (or page without results at all).  Displayed page contains
corresponding part of the query, one or multiple words as result and
links to other pages with different parts of query (if exist).

How this library works:

Library parses response and extracts links to other pages if they
exist. Then page content is splitted to words and parsed.  Words and
their definitions form the WordList for current part of query.

Description of types and their meaning in site terms:

Meaning - one line with multiple definitions in specific topic.
  eng.    | chain; complex; structure; type; integer (essence);
  ^ topic   ^ MeaningWord

MeaningWord - word from Meaning line.
  integer (essence)
  ^ word   ^ add (additional info)

Word - list of Meanings for word.
  число сущ. // Word.Word, Word.Part (part of speech)
     genet. number; date; figure; numeric; // Meaning
     autom. digit                          // Meaning

WordList - part of initial query with corresponding words.
  числа // WordList.Query
    число, ...  // Words
*/
package gmtrn

import (
	"bytes"
	"errors"
	"fmt"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"html"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// Available languages
var Languages = map[string]int{
	"english":   1,
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
var wordRe = regexp.MustCompile(`^<a href="(.*)">([^<]+)</a>.*( <em>|&nbsp;)`)

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

// Get data from url
func getData(url string) (result []Word, links []link, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Url %s returned HTTP code: %d",
			url, resp.StatusCode)
		return
	}
	defer resp.Body.Close()
	e := charmap.Windows1251
	decoder := transform.NewReader(io.Reader(resp.Body), e.NewDecoder())
	contentRaw, err := ioutil.ReadAll(decoder)
	if err != nil {
		return
	}
	content := string(contentRaw)
	// Find links to another parts of query
	otherLinks := otherRe.FindStringSubmatch(content)
	// Get links to another part of query
	if len(otherLinks) > 0 {
		linksRaw := linkRe.FindAllStringSubmatch(otherLinks[1], -1)
		links = make([]link, len(linksRaw))
		for i := 0; i < len(linksRaw); i++ {
			links[i] = link{html.UnescapeString(linksRaw[i][2]),
				domain + linksRaw[i][1]}
		}
	}
	result, err = parsePage(content)
	return
}

// Encode query to cp1251 and add necessart GET parameters
func getQuery(query string, lang int) (result string, err error) {
	buf := new(bytes.Buffer)
	encoded := ""
	// Dumb and dirty escaping of non-cyrillic characters
	for _, s := range query {
		code := uint64(s)
		if code < 127 || code > 1039 {
			encoded += string(s)
		} else {
			encoded += "&#" + strconv.FormatUint(code, 10) + ";"
		}
	}

	// Encode to 1251
	e := charmap.Windows1251
	encoder := transform.NewWriter(buf, e.NewEncoder())
	fmt.Fprintf(encoder, encoded)
	encoder.Close()

	values := url.Values{}
	values.Add("l1", strconv.Itoa(lang))
	values.Add("CL", "1") // Don't know what is it
	values.Add("q", buf.String())
	result = domain + "m.exe?" + values.Encode()
	return
}

// Run HTTP query to http://www.multitran.ru and return parsed results
// or error. Function can return error if translation isn't found or
// something wrong happens.
//
// lang - integer from Languages map
func Query(query string, lang int) (result []WordList, err error) {
	queryUrl, err := getQuery(query, lang)
	if err != nil {
		return
	}
	content, links, err := getData(queryUrl)
	if len(links) > 1 || (err != nil && len(links) > 0) {
		// Get data for another words in query
		result = make([]WordList, 0, len(links))
		var i int
		if err == nil {
			// First part of query found
			wl := WordList{content, links[0].word, links[0].link}
			result = append(result, wl)
			// Start from next link
			i = 1
		}
		// Get other links
		for ; i < len(links); i++ {
			other, _, oerr := getData(links[i].link)
			if oerr != nil {
				fmt.Printf("Error when getting %s: %s\n",
					links[i], oerr)
			} else {
				wl := WordList{other,
					links[i].word,
					links[i].link}
				result = append(result, wl)
			}
		}
		if len(result) > 0 && err != nil {
			// Case when some words not found is normal
			err = nil
		}
	} else {
		// Only one word returned by site
		result = []WordList{WordList{content, query, ""}}
	}
	return
}
