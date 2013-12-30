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
	"bitbucket.org/kardianos/osext"
	"code.google.com/p/go.text/unicode/norm"
	"github.com/mfonda/simhash"
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
	Key			string		//MD5 hash of its path
	Path		string		//Article full path 
	SimHash		uint64		//Hash representing the content
	Score		uint8		//Distance to another article
	Related 	articleList	//List of related articles sorted out by their scores
}

var (
	// Command line flags
    input = flag.String("input", "input", "Input folder")
	extensions = flag.String("extensions", ".*\\.html", "Regexp matching files to be added to the list")
	length = flag.Int("length", 5, "Number of related files to be displayed")
	// Global vars
	extRule *regexp.Regexp
	// Regular expressions
	remTags = regexp.MustCompile(`<[^>]*>`)
	oneSpace = regexp.MustCompile(`\s{2,}`)
	// Working directory
	inPath string
	//Cache of articles
	cache map[string]article
)

// Proof of concept demonstrating how to compute a list of related posts
func main() {

	start := time.Now()

	flag.Parse()
	exePath, _ := osext.ExecutableFolder()
	
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
	cache = make(map[string]article)
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
		results := getRelatedPosts(v.Path, *length)		//A bit silly to call this function here...
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
//	- The comparaison is the Levenshtein Distance with the other files' content
//	- Sorts the files by these scores (the lower the score is, the more related the content is)
//  - Stores the list of compared articles with their LD scores into a JSON file
func compareFiles() {
	start := time.Now()
	for baseKey, _ := range cache {
		var listOfArticles articleList
		for k, v := range cache {
			if k != baseKey {
				r := article{Key: v.Key,
							Path: v.Path,
							Score: simhash.Compare(cache[baseKey].SimHash,
													cache[k].SimHash)}
				listOfArticles = append(listOfArticles, r)
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

// cleanHTML cleans the input text so as to make it suitable for a comparison :
// - Replace all HTML tags <*> by space characters
// - Simplify the content by removing duplicated spaces
// NB: tolower is done by the Unicode/normalization
func cleanHTML(content string) string {
	result := content
	result = remTags.ReplaceAllString(result, " ")
	result = oneSpace.ReplaceAllString(result, " ")
	return result
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
			content = remTags.ReplaceAll(content, []byte(" "))
			content = oneSpace.ReplaceAll(content, []byte(" "))
			a := article{Key:name2key(path),
						Path: path,
						SimHash: simhash.Simhash(NewUnicodeWordFeatureSet(content, norm.NFKC))}
			cache[a.Key] = a
		}
	}
	return nil
}

/////////////////////////////////////////////////////////////
// simhash patch for Unicode / normalization support
// see : http://blog.golang.org/normalization
// see : https://groups.google.com/forum/#!topic/golang-nuts/YyH1f_qCZVc

// UnicodeWordFeatureSet is a feature set in which each word is a feature,
// all equal weight.
type UnicodeWordFeatureSet struct {
	b []byte
	f norm.Form
}

func NewUnicodeWordFeatureSet(b []byte, f norm.Form) *UnicodeWordFeatureSet {
	fs := &UnicodeWordFeatureSet{b, f}
	fs.normalize()
	return fs
}

func (w *UnicodeWordFeatureSet) normalize() {
	b := bytes.ToLower(w.f.Append(nil, w.b...))
	w.b = b
}

var boundaries = regexp.MustCompile(`[\pL-_']+`)

// Returns a []Feature representing each word in the byte slice
func (w *UnicodeWordFeatureSet) GetFeatures() []simhash.Feature {
	words := boundaries.FindAll(w.b, -1)
	features := make([]simhash.Feature, len(words))
	for i, w := range words {
		features[i] = simhash.NewFeature(w)
	}
	return features
}

/////////////////////////////////////////////////////////////