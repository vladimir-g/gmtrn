// Copyright 2012-2023 Vladimir Gorbunov. All rights reserved. Use of
// this source code is governed by a MIT license that can be found in
// the LICENSE file.

/*
Package gmtrn implements http client library for http://www.multitran.ru/


Usage:
	result, err := gmtrn.Query("Query string",
			    gmtrn.Languages["english"], // source language (from)
                            gmtrn.Languages["russian"]) // target language (to)


How multitran works

Site splits incoming query into multiple parts and displays results
for first part (or page without results at all).  Displayed page
contains corresponding part of the query, results as list of words and
links to other pages with other parts of query (if exist).

How this library works

Library makes request to the site and extracts reponse. Depending on
response, library may do additional requests to get all found word
definitions.

Every requested page is splitted into Words that have multiple
Meanings, and combined into WordList. For example, for query
"translation library" there would be two WordList objects, one for
"translation", other for "library". First one would contain multiple
words ("translation" (verb, noun), "translations" etc), and every Word
would have list of Meanings. Every object also contains a link to
corresponding page that may be used by library user.

Description of types in site terms:

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

How it looks on site:

 Word
   topic   meaning, meaning, meaning
   topic   meaning, meaning, meaning
   ...
 Word
   topic   meaning, meaning, meaning
   topic   meaning, meaning, meaning
 ...

Known issues

- Site has autodetection algorithm, so sometimes even uses different
  source/target languages, depending on query. It mostly ok though.

- Only default site interface language is implemented.

- There is no tests.

*/
package gmtrn
