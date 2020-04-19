// Copyright 2012-2020 Vladimir Gorbunov. All rights reserved.  Use of
// this source code is governed by a MIT license that can be found in
// the LICENSE file.

package gmtrn

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"strings"
)

// Multitran base domain with search path
var domain = "https://www.multitran.ru/"

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

// Parse page content and get links to other pages if exist
func parseLinks(doc *goquery.Document) (links []link, err error) {
	form := doc.Find("#translation")
	if form.Length() == 0 {
		return
	}
	if goquery.NodeName(form.Next()) == "table" {
		return
	}
	el := form
	for {
		el = el.Next()
		if goquery.NodeName(el) == "br" {
			break
		}
		if goquery.NodeName(el) == "a" {
			href, _ := el.Attr("href")
			links = append(links, link{el.Text(), domain + href})
		}
	}
	return
}

// Parse page content and get WordList (without Word and Link attrs)
func parsePage(doc *goquery.Document) (list WordList, err error) {
	table := doc.Find("div.middle_col > table")
	if table.Length() < 1 {
		return
	}
	word := Word{}
	table.Find("tr").Each(func(i int, tr *goquery.Selection) {
		td := tr.Find("td.gray")
		// New word
		if td.Length() == 1 {
			if word.Word != "" {
				list.Words = append(list.Words, word)
				word = Word{}
			}
			link := td.Find("a:first-child")
			word.Word = link.Text()
			href, _ := link.Attr("href")
			word.Link = domain + href
			word.Part = td.Find("em").Text()
		}
		// Word meaning row
		subj := tr.Find("td.subj")
		if subj.Length() == 1 {
			// mword := Meaning{subj.Text()
			mwords := make([]MeaningWord, 0)
			tr.Find("td.trans a").Each(func(i int, w *goquery.Selection) {
				link, _ := w.Attr("href")
				link = domain + link
				wd := w.Text()
				add := ""
				if goquery.NodeName(w.Next()) == "span" {
					add = w.Next().Text()
				}
				if goquery.NodeName(w.Prev()) == "span" {
					add = w.Prev().Text() + " " + add
				}
				add = strings.Trim(add, "()")
				mwords = append(mwords, MeaningWord{wd, link, add})
			})
			meaning := Meaning{mwords, subj.Text(), word.Link}
			word.Meanings = append(word.Meanings, meaning)
		}
	})
	return
}
