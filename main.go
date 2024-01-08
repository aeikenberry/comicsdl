package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

func parseSearch(doc *html.Node) ([]*SearchResult, error) {
	var results []*SearchResult
	var crawler func(*html.Node)
	crawler = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "h1" {
			var searchResult SearchResult
			setSearchResult(node, &searchResult)
			if searchResult.URL != "" {
				results = append(results, &searchResult)
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			crawler(child)
		}
	}
	crawler(doc)
	if results != nil {
		return results, nil
	}
	return nil, errors.New("No results found")
}

func setSearchResult(n *html.Node, r *SearchResult) {
	if n.Type == html.TextNode {
		r.Title = n.Data
	}
	for _, a := range n.Attr {
		if a.Key == "href" {
			r.URL = a.Val
			break
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		setSearchResult(c, r)
	}
}

func GetLinkURLS(doc *html.Node) ([]string, error) {
	var links []string
	var crawler func(*html.Node)

	crawler = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "a" {
			for _, a := range node.Attr {
				if a.Val == "Download Now" {
					url := getLink(node)
					links = append(links, url)
					return
				}
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			crawler(child)
		}
	}
	crawler(doc)
	if links != nil {
		return links, nil
	}

	return nil, errors.New("No direct download found")
}

type SearchResult struct {
	Title string
	URL   string
}

func getSearchResults(comics string) (*html.Node, error) {
	return getParsedHTML("https://getcomics.org/?s=" + url.QueryEscape(comics))
}

func getParsedHTML(url string) (*html.Node, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	// Parse
	return html.Parse(bytes.NewReader(body))
}

func getUserSeletion(searchResults []*SearchResult) *SearchResult {
	for i, v := range searchResults {
		fmt.Printf("%d: %s\n", i, v.Title)
	}

	// User chooses result
	fmt.Println("make selection: [0]")
	selection := "0"

	// Taking input from user
	_, err := fmt.Scanln(&selection)
	if err != nil {
		log.Fatalln(err)
	}

	// Convert to int
	n, err := strconv.Atoi(selection)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("You chose: %s\n", searchResults[n].Title)

	return searchResults[n]
}

func getLink(node *html.Node) string {
	for _, a := range node.Attr {
		if a.Key == "href" {
			return a.Val
		}
	}
	return ""
}

func download(link string, dest string) error {
	filePath := dest + "temp.txt"
	out, err := os.Create(filePath)
	if err != nil {
		log.Fatalln(err)
	}
	defer out.Close()
	resp, err := http.Get(link)
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()
	urlStr, err := url.QueryUnescape(resp.Request.URL.String())
	parts := strings.Split(urlStr, "/")
	last := parts[len(parts)-1]
	if !strings.Contains(last, ".cbz") && !strings.Contains(last, ".cbr") && !strings.Contains(last, ".zip") {
		return errors.New("Invalid download")
	}
	fmt.Printf("Downloading: %s\n", last)
	_, err = io.Copy(out, resp.Body)
	err = os.Rename(filePath, dest + last)

	fmt.Printf("Done! Saved: %s\n", dest + last)
	return nil
}

func main() {
	comic := flag.String("comic", "Swamp Thing", "What comic are we shooting for?")
	dest := flag.String("dest", "downloads/", "Where do you want to save the comic?")
	flag.Parse()
  if !strings.HasSuffix(*dest, "/") {
    suffix := "/"
    destDir := *dest + suffix
    os.Mkdir(destDir, 0755)
  } else {
    os.Mkdir(*dest, 0755)
  }
	fmt.Printf("Searching : %s\n", *comic)

	// Find the page links
	doc, err := getSearchResults(*comic)
	if err != nil {
		log.Fatalln(err)
	}
	// Parse the search results
	searchResults, err := parseSearch(doc)
	if err != nil {
		log.Fatalln(err)
	}

	// Show the options
	if len(searchResults) == 0 {
		log.Fatalln("No results found.")
	}

	// Get selection
	selection := getUserSeletion(searchResults)

	// Go to the url of their selection
	doc, err = getParsedHTML(selection.URL)

	// Find the download links
	urls, err := GetLinkURLS(doc)
	if err != nil {
		log.Fatalln(err)
	}

	// Download the file
	for _, v := range urls {
		err = download(v, *dest)
		if err == nil {
			return
		}
	}
}
