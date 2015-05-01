package gmtrn

import (
	"bytes"
	"fmt"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"html"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
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
