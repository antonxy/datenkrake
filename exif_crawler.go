package main

import (
	"fmt"
	"golang.org/x/net/html"
	"net/http"
    "net/url"
    "math/rand"
    "time"
	"os"
	"strings"
    "sync"
)

type UrlProvider interface {
    getNextUrl() string
    putUrl(url string)
}

type UrlProviderImpl struct {
    hosts map[string]*Host
}

type Host struct {
    openUrls []string
    visitedUrls map[string]bool
    lastVisit time.Time
}

func NewUrlProvider() UrlProvider {
    return &UrlProviderImpl{hosts: make(map[string]*Host)}
}
func (up *UrlProviderImpl) getNextUrl() string {
    for _, v := range up.hosts {
        if len(v.openUrls) > 0 && v.lastVisit.Before(time.Now().Add(-2 * time.Second)) {
            idx := rand.Intn(len(v.openUrls))
            foundUrl := v.openUrls[idx]
            v.openUrls = append(v.openUrls[:idx], v.openUrls[idx+1:]...)
            v.lastVisit = time.Now()
            return foundUrl
        } else {
            continue
        }
    }
    return ""
}

func (up *UrlProviderImpl) putUrl(foundUrl string) {
    u, err := url.Parse(foundUrl)
    if err != nil {
        fmt.Printf("Cant parse url `%s`\n", foundUrl)
        return
    }
    host, exists := up.hosts[u.Host]
    if !exists {
        host = NewHost()
        up.hosts[u.Host] = host
    }

    host.putUrl(foundUrl)
}

func NewHost() *Host {
    return &Host{visitedUrls: make(map[string]bool)}
}

func (h *Host) putUrl(url string) {
    if len(h.openUrls) > 100 {
        return
    }
    if _, exists := h.visitedUrls[url]; !exists {
        h.visitedUrls[url] = true
        h.openUrls = append(h.openUrls, url)
    }
}

// Helper function to pull the href attribute from a Token
func getHref(t html.Token) (ok bool, href string) {
	// Iterate over all of the Token's attributes until we find an "href"
	for _, a := range t.Attr {
		if a.Key == "href" {
			href = a.Val
			ok = true
		}
	}

	// "bare" return will return the variables (ok, href) as defined in
	// the function definition
	return
}

// Extract all http** links from a given webpage
func crawl(crawl_url string) {
	resp, err := http.Get(crawl_url)

	if err != nil {
		fmt.Println("ERROR: Failed to crawl \"" + crawl_url + "\"")
		return
	}

	b := resp.Body
	defer b.Close() // close Body when the function returns

	z := html.NewTokenizer(b)

	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			return
		case tt == html.StartTagToken:
			t := z.Token()

			// Check if the token is an <a> tag
			isAnchor := t.Data == "a"
			if !isAnchor {
				continue
			}

			// Extract the href value, if there is one
			ok, a_url := getHref(t)
			if !ok {
				continue
			}

            u, err := url.Parse(a_url)
            if err != nil {
                continue
            }
            base, err := url.Parse(crawl_url)
            if err != nil {
                continue
            }
            a_url_res := base.ResolveReference(u).String()

			// Make sure the url begines in http**
			hasProto := strings.Index(a_url_res, "http") == 0
			if hasProto {
                urlProviderLock.Lock()
                urlProvider.putUrl(a_url_res)
                urlProviderLock.Unlock()
				//ch <- a_url_res
			}
		}
	}
}

var urlProvider UrlProvider
var urlProviderLock sync.Mutex

func crawler(/*openUrls chan string, foundUrls chan string,*/ number int) {
    for {
//        url := <-openUrls
        urlProviderLock.Lock()
        url := urlProvider.getNextUrl()
        urlProviderLock.Unlock()
        if url == "" {
            fmt.Println("Did not find url - waiting")
            time.Sleep(2 * time.Second)
        } else {
            fmt.Printf("%d is crawling %s\n", number, url)
            crawl(url)
            fmt.Printf("%d is done\n", number)
        }
    }
}

func main() {
    urlProvider = NewUrlProvider()
    //visitedUrls := make(map[string]bool)
    //hostVisitCount := make(map[string]int)
	seedUrls := os.Args[1:]

	// Channels
//	chOpenUrls := make(chan string, 20)
    for _, url := range seedUrls {
//        chOpenUrls <- url
        urlProvider.putUrl(url)
    }
//    foundUrls := make (chan string, 100)

	// Kick off the crawl process (concurrently)
	for i := 0; i < 100; i++ {
		go crawler(i)
	}

//    for {
//        foundUrl := <-foundUrls
//        if _, exists := visitedUrls[foundUrl]; !exists {
//            u, err := url.Parse(foundUrl)
//            if err != nil {
//                continue
//            }
//            if val, exists := hostVisitCount[u.Host]; !exists || val < 20 || len(chOpenUrls) < 10 {
//                if !exists {
//                    hostVisitCount[u.Host] = 0
//                }
//                hostVisitCount[u.Host]++
//            } else {
//                fmt.Printf("Host %s visited too often\n", u.Host)
//            }
//            visitedUrls[foundUrl] = true
//            select {
//                case chOpenUrls <- foundUrl:
//                default:
//                    fmt.Println("Open list full")
//            }
//        }
//    }

//	close(chOpenUrls)
    select {}
}
