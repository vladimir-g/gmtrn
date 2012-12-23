=======================================================
 Gomultitran - HTTP client for http://www.multitran.ru
=======================================================

Gomultitran is a HTTP client for http://www.multitran.ru

This project contains two packages:

* webapi/webapi.go - library for HTTP access
* gomultitran-cli/gomultitran-cli.go - CLI client that uses webapi.go

Install
-------

CLI client::

 go get bitbucket.org/vladimir_g/gomultitran/gomultitran-cli

Previous command should install webapi.go library too, but you also
can install library without CLI client::

 go get bitbucket.org/vladimir_g/gomultitran/webapi

CLI usage
---------

Simple usage::

 $GOPATH/bin/gomultitran-cli translation string

Run this to get more usage options::

 $GOPATH/bin/gomultitran-cli -h

If name of the binary looks too long just add alias to your ~/.bashrc.

Or you can create simple wrapper script with freedesktop
notifications like this::

 #/bin/sh
 RESULT="$($GOPATH/bin/gomultitran-cli $@ | fold -sw 150)"
 notify-send -t 0 "<span font_family=\"monospace\">$RESULT</span>"

Change 150 to your preferred width (or remove fold completely), set
required popup timeout (0 in this example), and run this wrapper like
this (you can bind this command to some key)::

 /path/to/script/gmtw translation string

Library usage
-------------

Use this code::

	result, err := webapi.Query("Query string", 
        			    webapi.Languages["english"])

Look for more documentation at *webapi/webapi.go*


Known issues
------------

* There are some problems with translation to Kalmyk language, but
  reverse translation works fine. This problem happens because site uses
  wrong guessing algorithm for determining the source language.

* Only default language for site interface is implemented.

* There is no tests.

