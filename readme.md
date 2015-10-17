# What is gorelated ?

*gorelated* is a prototype demonstrating how to build a list of related posts in a static website generator written in go.
It relies on SIMHash algorithm, so the objective is to quickly build an approximate list and not to aim an exact lexical match.
In order to improve the accuracy, we apply the following process for each file:
1. Remove HTML tags.
2. Decode HTML entities if any.
3. Remove stop words (the most frequent words are useless for the meaning of the text).
4. Compute a fingerprint representative of the content (SimHash).

And then, we can sort the list of related content by computing the distance between the fingerprints.

*gorelated* comes with various HTML files coming from Wikipedia or the public domain.
You'll notice that you can improve the accuracy if you run the algorithm with the content only and not the entire HTML page.

# Usage

* **input**  Input folder (default: "input")
* **extensions** Regexp matching files to be added to the list(default: ".*\\.html")
* **length** Number of related files to be displayed (default: 5)
* **langCode** ISO 639-1 language code of the content (default: en)

# Dependancies

* https://github.com/bbalet/stopwords Libray cleaning the stop words, HTML tags and duplicated spaces in a content
* https://github.com/mfonda/simhash Calculate a hash reprensenting a text
* https://github.com/kardianos/osext get the folder where the executable is installed (not running)
* https://golang.org/x/text/unicode/norm Unicode normalization
