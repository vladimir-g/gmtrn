=========================================================
 gmtrn - Parser and CLI tool for http://www.multitran.ru
=========================================================

Gmtrn is a parser for http://www.multitran.ru written in Go.

This project contains parser library and simple CLI client.

Install
-------

Install Go and set up Go language environment (`official docs`_).

Install CLI client::

 go get github.com/vladimir-g/gmtrn/cmd/gmtrn-cli

Library also can be installed without CLI::

 go get github.com/vladimir-g/gmtrn

CLI usage
---------

Simple usage::

 $GOPATH/bin/gmtrn-cli translation string

More usage options available in help::

 $GOPATH/bin/gmtrn-cli -h

If name of the binary looks too long just add alias for it.

Or use simple wrapper script with freedesktop notifications like
this::

 #/bin/sh
 RESULT="$($GOPATH/bin/gmtrn-cli $@ | fold -sw 150)"
 notify-send -t 0 "<span font_family=\"monospace\">$RESULT</span>"

Change 150 to your preferred width (or remove fold completely), set
required popup timeout (0 in this example), and run this wrapper like
this (you can bind this command to some key)::

 /path/to/script/gmtw translation string

CLI app can also output results in JSON format.

Library usage
-------------

Use this code::

 result, err := gmtrn.Query("Query string", gmtrn.Languages["english"])

More documentation in `doc.go`_


Known issues
------------

* There are some problems with translation to Kalmyk language, but
  reverse translation works fine. This problem happens because site uses
  wrong guessing algorithm for determining the source language.

* Sometimes parser fails (very rare).

* Only default language for site interface is implemented.

* There is no tests.

License
=======

This library released under MIT license, see LICENSE file.

.. _official docs: https://golang.org/doc/code.html
.. _doc.go: doc.go
