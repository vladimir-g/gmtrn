// Copyright 2012-2023 Vladimir Gorbunov. All rights reserved. Use of
// this source code is governed by a MIT license that can be found in
// the LICENSE file.

package gmtrn

import (
	"fmt"
	"golang.org/x/net/html"
	"io"
	"strings"
)

// Multitran base domain with search path
var domain = "https://www.multitran.com/"

// Single part of definition that contains a single word, link to the
// word page and additional information. Link often couldn't be opened
// without "http://www.multitran.ru" referer for access.
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
	Link string
}

func (d Meaning) String() string {
	return fmt.Sprintf(`<Meaning, Topic: "%s", "%s", "%s">`,
		d.Topic, d.Link, d.Words)
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

// List of words at page
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

// Check if node is specific tag
func isTag(node *html.Node, tagName string) bool {
	return node.Type == html.ElementNode && node.Data == tagName
}

// Get node attribute value, returns empty string if not found
func attrValue(node *html.Node, attrName string) string {
	for _, attr := range node.Attr {
		if attr.Key == attrName {
			return attr.Val
		}
	}
	return ""
}

// Get text contents of node and all siblings
func textContents(node *html.Node) string {
	var sb strings.Builder
	var recurse func(n *html.Node)
	recurse = func(n *html.Node) {
		if n.Type == html.TextNode {
			if sb.Len() > 0 {
				sb.WriteString(" ")
			}
			sb.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			recurse(c)
		}
	}
	recurse(node)
	return sb.String()
}

// Parse single Word
func parseWord(node *html.Node) (word Word) {
	// <tr><td class="gray"><a name="PART"></a><a href="LINK">TEXT</a></td></tr>
	var sb strings.Builder
	for n := node.FirstChild; n != nil; n = n.NextSibling {
		if isTag(n, "a") {
			name := attrValue(n, "name")
			if name != "" {
				word.Part = name
				continue
			}
			href := attrValue(n, "href")
			if href != "" {
				word.Link = domain + href
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.TextNode {
					if sb.Len() > 0 {
						sb.WriteString(" ")
					}
					sb.WriteString(c.Data)
				}
			}
		}
	}
	word.Word = sb.String()
	return
}

// Parse list of MeaningWords
func parseMeaningWords(node *html.Node) (mwords []MeaningWord) {
	// <tr><td class="subj">SUBJ</td><td class="trans">LIST OF WORDS</td></td>
	var mword MeaningWord
	for n := node.FirstChild; n != nil; n = n.NextSibling {
		if n.Type == html.TextNode && n.Data == ";" {
			// Next word
			if mword.Word != "" {
				mwords = append(mwords, mword)
				mword = MeaningWord{}
			}
			continue
		}
		if isTag(n, "a") {
			// Word contents
			mword.Word = textContents(n)
			mword.Link = domain + attrValue(n, "href")
			continue
		}
		if isTag(n, "span") {
			mword.Add = textContents(n)
			continue
		}
	}
	if mword.Word != "" {
		mwords = append(mwords, mword)
	}
	return
}

// Parse single Meaning
func parseMeaning(node *html.Node) (meaning Meaning) {
	// <tr><td class="subj">SUBJ</td><td class="trans">LIST OF WORDS</td></td>
	for td := node.FirstChild; td != nil; td = td.NextSibling {
		if !isTag(td, "td") {
			continue
		}
		if attrValue(td, "class") == "subj" {
			meaning.Topic = textContents(td)
		} else if attrValue(td, "class") == "trans" {
			meaning.Words = parseMeaningWords(td)
		}
	}
	return
}

// Parse table with words
func parseTable(table *html.Node, list *WordList) {
	var tbody *html.Node
	for n := table.FirstChild; n != nil; n = n.NextSibling {
		if isTag(n, "tbody") {
			tbody = n
			break
		}
	}
	if tbody == nil {
		return
	}
	word := Word{}
TR:
	for tr := tbody.FirstChild; tr != nil; tr = tr.NextSibling {
		if !isTag(tr, "tr") {
			continue
		}
		// Parse trs
		for td := tr.FirstChild; td != nil; td = td.NextSibling {
			if !isTag(td, "td") {
				continue
			}
			if attrValue(td, "class") == "gray" {
				// New word
				if word.Word != "" {
					list.Words = append(list.Words, word)
				}
				word = parseWord(td)
				continue TR
			} else if attrValue(td, "class") == "subj" {
				meaning := parseMeaning(tr)
				word.Meanings = append(word.Meanings, meaning)
				continue TR
			}
		}
	}
	if word.Word != "" {
		list.Words = append(list.Words, word)
	}
}

// Get links to other pages of query if exist
func parseLink(n *html.Node, links *[]link) {
	a := n.FirstChild
	if isTag(a, "a") {
		*links = append(*links, link{textContents(a), domain + attrValue(a, "href")})
	}
}

// Recursively walk over html nodes tree and parse contents
func walkTree(n *html.Node, list *WordList, links *[]link) {
	if isTag(n, "table") && attrValue(n, "width") == "100%" {
		// Check width fixme
		parseTable(n, list)
	} else if isTag(n, "span") && attrValue(n, "class") == "tooltip" {
		parseLink(n, links)
	} else {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walkTree(c, list, links)
		}
	}
	return
}

// Parse site page
func parsePage(r io.Reader, list *WordList, links *[]link) (err error) {
	doc, err := html.Parse(r)
	if err != nil {
		return
	}
	walkTree(doc, list, links)
	return
}
