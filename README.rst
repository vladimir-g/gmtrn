==========================================================
 gmtrn - Parser and CLI tool for http://www.multitran.com
==========================================================

Gmtrn is a parser for http://www.multitran.com written in Go.

This project contains parser library and simple CLI client.

Install
-------

Install Go and set up Go language environment (`official docs`_).

Simple installation of CLI tool::

 git clone https://github.com/vladimir-g/gmtrn/
 cd gmtrn/cmd/gmtrn-cli
 go build .

These commands would generate ``gmtrn-cli`` binary.

Library also can be installed without CLI::

 go get github.com/vladimir-g/gmtrn

For new versions of golang import statement is enough.

CLI usage
---------

Simple usage::

 $GOPATH/bin/gmtrn-cli translation string

More usage options available in help::

 $GOPATH/bin/gmtrn-cli -h

If name of the binary looks too long just add alias for it.

Example script with xsel and freedesktop notifications::

 #/bin/sh
 notify-send -t 0 \
   "<span font='monospace'>$(xsel | xargs -0 gmtrn-cli | fold -sw 100)</span>"

Another example with zenity::

 #/bin/sh

 xsel | xargs -0 gmtrn-cli | fold -sw 120 | zenity --no-wrap --text-info --width=800 --height=500

CLI app can also output results in JSON format that may be parsed with
some utility.

Library usage
-------------

Use this code::

 result, err := gmtrn.Query("Query string",
                            gmtrn.Languages["english"],
                            gmtrn.Languages["russian"])

More documentation in `doc.go`_


Known issues
------------

* Site has autodetection algorithm, so sometimes even uses different
  source/target languages, depending on query. It mostly ok though.

* Only default site interface language is implemented.

* There is no tests.

License
=======

This library released under MIT license, see LICENSE file.

.. _official docs: https://golang.org/doc/install
.. _doc.go: doc.go
