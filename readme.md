# What is gorelated ?
*gorelated* is a prototype demonstrating how to build a list of related posts in a static website generator written in go.
It comes with some HTML files coming from Wikipedia or the public domain.

# Usage

* **input**  Input folder (default: "input")
* **extensions* Regexp matching files to be added to the list(default: ".*\\.html")
* **length** Number of related files to be displayed (default: 5)

# Dependancies

* bitbucket.org/kardianos/osext get the folder where the executable is installed (not running)
* code.google.com/p/go.text/unicode/norm Unicode normalization
* github.com/mfonda/simhash Calculate a hash reprensenting a text
