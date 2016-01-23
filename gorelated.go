package main

/*  gorelated finds the list of related text files from a folder
	  Copyright (C) 2013 Benjamin BALET

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with this program.  If not, see <http://www.gnu.org/licenses/>.*/

import (
	"bytes"
	"crypto/md5"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"encoding/csv"
	"encoding/json"

	"github.com/bbalet/stopwords"
	"github.com/kardianos/osext"
)

// Definitions used by sort package
type articleList []article

func (a articleList) Len() int {
	return len(a)
}
func (a articleList) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a articleList) Less(i, j int) bool {
	return a[i].Score < a[j].Score
}

// An article
type article struct {
	Key     		string      //MD5 hash of its path
	Path    		string      //Article full path
	URL					string			//If we parse a Jekyll list of posts, export the URL
	Title 			string			//If we parse a Jekyll list of posts, export the Title
	Description	string			//If we parse a Jekyll list of posts, export the Description
	Thumbnail	  string			//If we parse a Jekyll list of posts, export the Thumbnail
	SimHash 		uint64      //Hash representing the content
	Score   		uint8       //Distance to another article
	Related 		articleList //List of related articles sorted out by their scores
}

// A Jekyll Post
type post struct {
	URL					string			//If we parse a Jekyll list of posts, export the URL
	Title 			string			//If we parse a Jekyll list of posts, export the Title
	Description	string			//If we parse a Jekyll list of posts, export the Description
	Thumbnail	  string			//If we parse a Jekyll list of posts, export the Thumbnail
	Related 		[]post		  //List of related articles sorted out by their scores
}

var (
	// Command line flags
	input      = flag.String("input", "input/en", "Input folder")
	extensions = flag.String("extensions", ".*\\.html", "Regexp matching files to be added to the list")
	length     = flag.Int("length", 5, "Number of related files to be displayed")
	langCode   = flag.String("lang", "en", "ISO 639-1 language code of the content")
	jekyll     = flag.String("jekyll", "", "Path to the list of posts")
	//Location of executable
	exePath string
	// Global vars
	extRule *regexp.Regexp
	// Working directory
	inPath string
	//Cache of articles
	cache map[string]article
)

// gorealated is a prototype showing how to build a list of related articles
// for a static website generator such as higo or Jekyll
func main() {
	flag.Parse()
	exePath, _ = osext.ExecutableFolder()
  cache = make(map[string]article)
	if *jekyll=="" {
		protoSimHash()
	} else {
		parseJekyllList()
	}
}

// parseJekyllList builds a list of related posts
func parseJekyllList() {
	start := time.Now()
	log.Println("Analyzing a list of Jekyll posts...")

	// open input file
	fi, err := os.Open(*jekyll)
	if err != nil {
			panic(err)
	}
	defer func() {
			if err := fi.Close(); err != nil {
					panic(err)
			}
	}()
	r := csv.NewReader(fi)

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		content, err := ioutil.ReadFile(record[1])
		if err != nil {
			log.Fatal("I Can't read this file : ", record[1])
		}
		a := article{Key:name2key(record[1]),
			 Path:record[1],	SimHash:stopwords.Simhash(content, *langCode, true),
			 URL:record[0],
			 Title: record[2],
			 Description: record[3],
		   Thumbnail: record[4]}
		cache[a.Key] = a
	}
	//Sort the list of articles
	compareFiles()

	//Create a collection of posts easily usable with Jekyll Liquid tags
	var jekyllPosts []post
	for _, articlePost := range cache {
			aPost := post {URL:articlePost.URL,Title:articlePost.Title,
				Description:articlePost.Description,
				Thumbnail:articlePost.Thumbnail}
			for _, articleRelated := range articlePost.Related {
					aRelatedPost := post {URL:articleRelated.URL,Title:articleRelated.Title,
						Description:articleRelated.Description,
						Thumbnail:articleRelated.Thumbnail}
					aPost.Related = append(aPost.Related, aRelatedPost)
			}
			jekyllPosts = append(jekyllPosts, aPost)
	}

	b, err := json.Marshal(jekyllPosts)
	if err != nil {
		log.Fatal(err)
	}
	var out bytes.Buffer
	json.Indent(&out, b, "", "\t")

	// open output file
	fo, err := os.Create(filepath.Join(exePath, "posts.json"))
	if err != nil {
			panic(err)
	}
	defer func() {
			if err := fo.Close(); err != nil {
					panic(err)
			}
	}()
	out.WriteTo(fo)
	log.Println("Total time : ", time.Since(start))
}

// protoSimHash prints a list of related posts
// Proof of concept demonstrating how to compute a list of related posts
func protoSimHash() {
	start := time.Now()
	//Check if input folder exists
	if !filepath.IsAbs(*input) {
		inPath = filepath.Join(exePath, *input)
	} else {
		inPath = *input
	}
	if _, err := os.Stat(inPath); os.IsNotExist(err) {
		log.Fatal("no such directory: ", inPath)
	}

	//Recursively Browse a folder to find all files
	log.Println("Reading and analyzing all input files...")
	extRule = regexp.MustCompile(*extensions)
	err := filepath.Walk(inPath, visit)
	if err != nil {
		log.Fatal("I Can't browse recursively this folder : ", inPath)
	}
	log.Println("Reading time : ", time.Since(start))

	log.Println("Comparing all input files between one another...")
	compareFiles()
	log.Println("Total time : ", time.Since(start))

	log.Println("Below, the results ")
	for _, v := range cache {
		fmt.Println("---;File;", filepath.Base(v.Path))
		results := getRelatedPosts(v.Path, *length) //A bit silly to call this function here...
		for _, item := range results {
			fmt.Println("Distance;File;", item.Score, filepath.Base(item.Path))
		}
	}
}


// getRelatedPosts gets the list of the {numberPosts} related posts
func getRelatedPosts(filePath string, numberPosts int) articleList {
	if numberPosts >= len(cache) {
		numberPosts = len(cache) - 1
	}
	return cache[name2key(filePath)].Related[:numberPosts]
}

// compareFiles :
//  - Opens a file and compare it to all the other files
//	- The comparaison is the Hamming Distance with the other files' content (using SimHash algo)
//	- Sorts the files by these scores (the lower the score is, the more related the content is)
//  - Stores the list of compared articles with their LD scores into a JSON file
func compareFiles() {
	start := time.Now()
	for baseKey, _ := range cache {
		var listOfArticles articleList
		var max int
		for k, v := range cache {
			if k != baseKey && max<*length {
				r := article{Key: v.Key,
					Path:					v.Path,
					URL:					v.URL,
					Title: 				v.Title,
					Description:	v.Description,
					Thumbnail:		v.Thumbnail,
					SimHash: 			v.SimHash,
					Score: 				stopwords.CompareSimhash(cache[baseKey].SimHash,cache[k].SimHash)}
				listOfArticles = append(listOfArticles, r)
				max++
			}
		}

		//Sort the results by their relative scores
		sort.Sort(articleList(listOfArticles))
		//Query, and assign back... (see https://code.google.com/p/go/issues/detail?id=3117)
		temp := cache[baseKey]
		temp.Related = listOfArticles
		cache[baseKey] = temp
	}
	log.Println("analyzeFile duration : ", time.Since(start))
}

// name2key returns the MD5 hash of a given string
func name2key(name string) string {
	h := md5.New()
	io.WriteString(h, name)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// visit is called by Walker, it adds the files found recursively
// the files must match the pattern passed as a parameter of the application
// We hash the content of each file
func visit(path string, f os.FileInfo, err error) error {
	if !f.IsDir() {
		if extRule.MatchString(path) {
			content, err := ioutil.ReadFile(path)
			if err != nil {
				log.Fatal("I Can't read this file : ", path)
			}
			// Remove all HTML tags, simplify the content and calculate the hash
			a := article{Key:name2key(path), Path:path,	SimHash:stopwords.Simhash(content, *langCode, true)}
			cache[a.Key] = a
		}
	}
	return nil
}
