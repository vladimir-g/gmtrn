// Copyright 2012-2023 Vladimir Gorbunov. All rights reserved. Use of
// this source code is governed by a MIT license that can be found in
// the LICENSE file.

/*
Package gmtrn implements http client library for [multitran.com]

Usage:

   result, err := gmtrn.Query("Query string",
		       gmtrn.Languages["english"], // source language (from)
                       gmtrn.Languages["russian"]) // target language (to)

# How multitran works
   
   Requested query is splitted into multiple parts depending on found
   translations. Results page contains corresponding part of the
   query, list of words and links to other pages with other parts of
   query (if exist).

   For example, query "first second third" may be splitted by
   multitran to pages with "first second" and "third".

# How this library works

   Library makes request to the site and extracts first page. If this
   page contains links to other pages with separate word translations,
   they are requested next.

   Every requested page is splitted into Words that have multiple
   Meanings, and combined into WordList. For example, for query
   "translation library" there would be two WordList objects, one for
   "translation", other for "library". First one would contain
   multiple words ("translation" (verb, noun), "translations" etc),
   and every Word would have list of Meanings. Word also may contain
   optional pre- and post-parts that provide some context, and also
   optional phonetic spelling. Every object also contains a link to
   corresponding page that may be used by library user.

   Description of types in site terms:

	Meaning - one line with multiple definitions in specific topic.
	  eng.    | chain; complex; structure; type; integer (essence);
	  ^ topic   ^ MeaningWord

	MeaningWord - word from Meaning line.
	  integer (essence)
	  ^ word   ^ add (additional info)

	Word - list of Meanings for word.
	  общее число [...] π сущ. // Word.Pre, Word.Word, Word.Post, Word.Spelling, Word.Part (part of speech)
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

# Known issues

  - Site has autodetection algorithm, so sometimes even uses different
    source/target languages, depending on query. It mostly ok though.

  - Only default site interface language is implemented.

  - Thesaurus is parsed as simple translation table.

  - There is no tests.

[multitran.com]: https://www.multitran.com/

*/
package gmtrn
