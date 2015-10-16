# What is gorelated ?
*gorelated* is a prototype demonstrating how to build a list of related posts in a static website generator written in go.
It comes with some HTML files coming from Wikipedia or the public domain.

# Usage

* **input**  Input folder (default: "input")
* **extensions** Regexp matching files to be added to the list(default: ".*\\.html")
* **length** Number of related files to be displayed (default: 5)
* **langCode** ISO 639-1 language code of the content (default: en)

# Dependancies

* github.com/kardianos/osext get the folder where the executable is installed (not running)
* golang.org/x/text/unicode/norm Unicode normalization
* github.com/mfonda/simhash Calculate a hash reprensenting a text
* github.com/bbalet/stopwords Libray cleaning the stop words, HTML tags and duplicated spaces in a content