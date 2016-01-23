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
* **jekyll** Path to the list of posts. If not empty, generate a list of related articles for Jekyll. If empty, display the result in the console

## How to use it with Jekyll?

### 1. Compile gorelated

 Compile *gorelated* and copy it into Jekyll folder or add it to the PATH.

### 2. Create a Jekyll plugin

This example of Jekyll plugin will perform the following tasks :
1. Extract the list of your posts and their path.
2. Execute the *gorelated* program with the list of your posts and wait until it exits
3. *gorelated* will create a JSON file into your `_data` folder for a later use.

    module Reading
      class Generator < Jekyll::Generator
        def generate(site)
          #Prepare a list of posts that will be processed by the Go program
          url  = Jekyll.configuration({})['url']
          listPath = File.join(site.source, '_data', 'posts.txt')
          file = File.new(listPath, "w")
          site.posts.each do |post|
            if post.data['published']
              file.write(url + post.url + "," + post.path  + ",\"" + post.data['title'] + "\",\"" + post.data['description'] + " \"," + post.data['thumbnail'] + "\n")
            end
          end
          file.close
          #Call the Go program that will process the list and generate a JSON file
          exePath = File.join(site.source, '_data', 'gorelated')
          system exePath + " -jekyll \"" + listPath + "\""
        end
      end
    end

### 3. Use the list of related articles with Liquid tags

From this point, the JSON file can be used by Liquid tags, for example:

    <div class="row-fluid">
      {% for postscol in site.data.posts %}
      {% if postscol.Title == page.title %}
      {% for related in postscol.Related limit:5 %}
      <div class="span2">
        <a href="{{ related.URL }}" target="_top" title="Link to the article {{ related.Title }}">
          <img height=120 width=120 src="{{ related.Thumbnail }}" title="Read the article {{ related.Title }}" class="img-polaroid">
        <h5>{{ related.Title }}</h5></a>
        <p>{{ related.Description }}</p>
      </div>
      {% endfor %}
      {% endif %}
      {% endfor %}
    </div>

# Dependancies

* https://github.com/bbalet/stopwords Libray cleaning the stop words, HTML tags and duplicated spaces in a content
* https://github.com/kardianos/osext get the folder where the executable is installed (not running)
* https://golang.org/x/text/unicode/norm Unicode normalization
