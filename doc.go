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
