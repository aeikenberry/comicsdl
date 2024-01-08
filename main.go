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

func GetResults(doc *html.Node) ([]*html.Node, error) {
	var titles []*html.Node
	var crawler func(*html.Node)
	crawler = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "h1" {
			title := node
			titles = append(titles, title)
			return
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			crawler(child)
		}
	}
	crawler(doc)
	if titles != nil {
		return titles, nil
	}
	return nil, errors.New("Missing <body> in the node tree")
}

func collectText(n *html.Node, r *SearchResult) {
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
		collectText(c, r)
	}
}

func GetLinkNodes(doc *html.Node) ([]*html.Node, error) {
	var links []*html.Node
	var crawler func(*html.Node)

	crawler = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "a" {
			for _, a := range node.Attr {
				if a.Val == "Download Now" {
					links = append(links, node)
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

type DownloadLink struct {
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

func parseSearchResults(nodes []*html.Node) []SearchResult {
	var searchResults []SearchResult

	for _, v := range nodes {
		var searchResult SearchResult
		collectText(v, &searchResult)
		if searchResult.URL != "" {
			searchResults = append(searchResults, searchResult)
		}
	}

	return searchResults
}

func getUserSeletion(searchResults []SearchResult) SearchResult {
	for i, v := range searchResults {
		fmt.Printf("%d: %s - %s\n", i, v.Title, v.URL)
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

func downloadNode(node *html.Node) error {
	link := getLink(node)
	fmt.Printf("Downloading: %s\n", link)
	// Download it.
	out, err := os.Create("output.txt")
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
	fmt.Println(urlStr)
	if !strings.Contains(last, ".cbz") && !strings.Contains(last, ".cbr") && !strings.Contains(last, ".zip") {
		return errors.New("Invalid download")
	}
	_, err = io.Copy(out, resp.Body)
	err = os.Rename("output.txt", last)

	fmt.Printf("Done! Saved: %s\n", last)
	return nil
}

func main() {
	comic := flag.String("comic", "Swamp Thing", "What comic are we shooting for?")
	flag.Parse()
	fmt.Printf("Searching : %s\n", *comic)

	// Find the page links
	doc, err := getSearchResults(*comic)
	if err != nil {
		log.Fatalln(err)
	}
	// Parse the search results
	results, err := GetResults(doc)
	if err != nil {
		log.Fatalln(err)
	}
	searchResults := parseSearchResults(results)

	// Show the options
	if len(searchResults) == 0 {
		log.Fatalln("No results found.")
	}

	// Get selection
	selection := getUserSeletion(searchResults)

	// Go to the url of their selection
	doc, err = getParsedHTML(selection.URL)

	// Find the download links
	nodes, err := GetLinkNodes(doc)
	if err != nil {
		log.Fatalln(err)
	}

	// Download the file
	for _, v := range nodes {
		err = downloadNode(v)
		if err == nil {
			return
		}
	}

}
