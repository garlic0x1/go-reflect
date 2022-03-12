// credit to @hakluke for most of this code https://github.com/hakluke/hakrawler

package main

import (
	"bufio"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
)

type injection struct {
	Hash         string
	FormLocation string
}

type input struct {
	Type  string
	Name  string
	Value string
}

type form struct {
	URL    string
	Method string
	Inputs []input
}

var (
	// Thread safe map
	sm      sync.Map
	headers map[string]string
	// record all the form inputs performed se we know where each found hash comes from
	injectionMap []injection
	// seed rand for randomString()
	seededRand *rand.Rand = rand.New(
		rand.NewSource(time.Now().UnixNano()))
)

func main() {
	threads := flag.Int("t", 8, "Number of threads to utilise.")
	depth := flag.Int("d", 2, "Depth to crawl.")
	insecure := flag.Bool("insecure", false, "Disable TLS verification.")
	subsInScope := flag.Bool("subs", false, "Include subdomains for crawling.")
	showSource := flag.Bool("s", false, "Show the source of URL based on where it was found (href, form, script, etc.)")
	rawHeaders := flag.String(("h"), "", "Custom headers separated by two semi-colons. E.g. -h \"Cookie: foo=bar;;Referer: http://example.com/\" ")
	proxy := flag.String(("proxy"), "", "Proxy URL, example: -proxy http://127.0.0.1:8080")
	unique := flag.Bool(("u"), false, "Show only unique urls")

	flag.Parse()

	if *proxy != "" {
		os.Setenv("PROXY", *proxy)
		*insecure = true
	}
	proxyURL, _ := url.Parse(os.Getenv("PROXY"))

	// Convert the headers input to a usable map (or die trying)
	err := parseHeaders(*rawHeaders)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error parsing headers:", err)
		os.Exit(1)
	}

	// Check for stdin input
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		fmt.Fprintln(os.Stderr, "No urls detected. Hint: cat urls.txt | hakrawler")
		os.Exit(1)
	}

	results := make(chan string, *threads)
	go func() {
		// get each line of stdin, push it to the work channel
		s := bufio.NewScanner(os.Stdin)
		for s.Scan() {
			url := s.Text()
			hostname, err := extractHostname(url)
			if err != nil {
				log.Println("Error parsing URL:", err)
				return
			}

			allowed_domains := []string{hostname}
			// if "Host" header is set, append it to allowed domains
			if headers != nil {
				if val, ok := headers["Host"]; ok {
					allowed_domains = append(allowed_domains, val)
				}
			}

			// Instantiate default collector
			c := colly.NewCollector(
				// default user agent header
				colly.UserAgent("Mozilla/5.0 (X11; Linux x86_64; rv:78.0) Gecko/20100101 Firefox/78.0"),
				// set custom headers
				colly.Headers(headers),
				// limit crawling to the domain of the specified URL
				colly.AllowedDomains(allowed_domains...),
				// allow revisiting to find stored hashes
				colly.AllowURLRevisit(),
				// set MaxDepth to the specified depth
				colly.MaxDepth(*depth),
				// specify Async for threading
				colly.Async(true),
			)

			// if -subs is present, use regex to filter out subdomains in scope.
			if *subsInScope {
				c.AllowedDomains = nil
				c.URLFilters = []*regexp.Regexp{regexp.MustCompile(".*(\\.|\\/\\/)" + strings.ReplaceAll(hostname, ".", "\\.") + "((#|\\/|\\?).*)?")}
			}

			// Set parallelism
			c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: *threads})

			c.OnResponse(func(r *colly.Response) {
				for i := 0; i < len(injectionMap); i++ {
					if strings.Contains(string(r.Body), injectionMap[i].Hash) {
						// build response
						response := fmt.Sprintf("Injection from %s found at %s", injectionMap[i].FormLocation, r.Request.URL)
						printReflection(response, "reflector", *showSource, results)
					}
				}
			})

			// Print every href found, and visit it
			c.OnHTML("a[href]", func(e *colly.HTMLElement) {
				link := e.Attr("href")
				printResult(link, "href", *showSource, results, e)
				e.Request.Visit(link)
			})

			// find and print all the JavaScript files
			c.OnHTML("script[src]", func(e *colly.HTMLElement) {
				printResult(e.Attr("src"), "script", *showSource, results, e)
			})

			// find and print all the form action URLs
			c.OnHTML("form[action]", func(e *colly.HTMLElement) {
				printResult(e.Attr("action"), "form", *showSource, results, e)
			})

			c.OnHTML("form", func(e *colly.HTMLElement) {
				hash := randomString(8)
				action := e.Request.AbsoluteURL(e.Attr("action"))
				method := e.Attr("method")

				var inputs []input
				e.ForEach("input", func(_ int, e *colly.HTMLElement) {
					inputs = append(inputs, input{
						Type:  e.Attr("type"),
						Name:  e.Attr("name"),
						Value: e.Attr("value"),
					})
				})
				e.ForEach("textarea", func(_ int, e *colly.HTMLElement) {
					inputs = append(inputs, input{
						Type:  "text",
						Name:  e.Attr("name"),
						Value: e.Attr("value"),
					})
				})

				f := form{
					URL:    action,
					Method: method,
					Inputs: inputs,
				}

				// append to injectionMap
				injectionMap = append(injectionMap, injection{
					Hash:         hash,
					FormLocation: action,
				})

				// set up proxy
				if *proxy != "" {
					// Skip TLS verification if -insecure flag is present
					c.WithTransport(&http.Transport{
						Proxy:           http.ProxyURL(proxyURL),
						TLSClientConfig: &tls.Config{InsecureSkipVerify: *insecure},
					})
				} else {
					c.WithTransport(&http.Transport{
						TLSClientConfig: &tls.Config{InsecureSkipVerify: *insecure},
					})
				}
				// add the custom headers
				if headers != nil {
					c.OnRequest(func(r *colly.Request) {
						for header, value := range headers {
							r.Headers.Set(header, value)
						}
					})
				}

				// send the form request
				if method == "POST" || method == "post" {
					e.Request.PostRaw(action, generateFormData(f, hash))
				} else if method == "GET" || method == "get" {
					e.Request.Visit(string(generateFormData(f, hash)))
				}

			})

			// add the custom headers
			if headers != nil {
				c.OnRequest(func(r *colly.Request) {
					for header, value := range headers {
						r.Headers.Set(header, value)
					}
				})
			}

			if *proxy != "" {
				// Skip TLS verification if -insecure flag is present
				c.WithTransport(&http.Transport{
					Proxy:           http.ProxyURL(proxyURL),
					TLSClientConfig: &tls.Config{InsecureSkipVerify: *insecure},
				})
			} else {
				c.WithTransport(&http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: *insecure},
				})
			}

			// Start scraping
			c.Visit(url)
			// Wait until threads are finished
			c.Wait()

		}
		if err := s.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading standard input:", err)
		}
		close(results)
	}()

	// listen to results channel and write to stdout
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	if *unique {
		for res := range results {
			if isUnique(res) {
				fmt.Fprintln(w, res)
			}
		}
	}
	for res := range results {
		fmt.Fprintln(w, res)
	}

}

// takes a form struct and returns a byte array of form inputs
// if its a POST form it returns POST data
// if its a GET form it returns a URL
func generateFormData(f form, hash string) []byte {
	formData := url.Values{}
	for i := 0; i < len(f.Inputs); i++ {
		if f.Inputs[i].Type == "hidden" {
			//payload = payload + "&" + f.Inputs[i].Name + "=" + f.Inputs[i].Value
			formData.Add(f.Inputs[i].Name, f.Inputs[i].Value)
		} else if f.Inputs[i].Type == "email" {
			//payload = payload + "&" + f.Inputs[i].Name + "=" + HASH + "@gmail.com"
			formData.Add(f.Inputs[i].Name, fmt.Sprintf("%s@gmail.com", hash))
		} else if f.Inputs[i].Type == "text" {
			//payload = payload + "&" + f.Inputs[i].Name + "=http://" + HASH
			formData.Add(f.Inputs[i].Name, fmt.Sprintf("http://%s", hash))
		} else if f.Inputs[i].Type == "password" {
			//payload = payload + "&" + f.Inputs[i].Name + "=" + hash
			formData.Add(f.Inputs[i].Name, hash)
		}
	}
	byteData, err := ioutil.ReadAll(strings.NewReader(formData.Encode()))
	if err != nil {
		log.Println(err)
	}
	if f.Method == "POST" || f.Method == "post" {
		return byteData
	}
	return []byte(f.URL + "?" + string(byteData))
}

// parseHeaders does validation of headers input and saves it to a formatted map.
func parseHeaders(rawHeaders string) error {
	if rawHeaders != "" {
		if !strings.Contains(rawHeaders, ":") {
			return errors.New("headers flag not formatted properly (no colon to separate header and value)")
		}

		headers = make(map[string]string)
		rawHeaders := strings.Split(rawHeaders, ";;")
		for _, header := range rawHeaders {
			var parts []string
			if strings.Contains(header, ": ") {
				parts = strings.SplitN(header, ": ", 2)
			} else if strings.Contains(header, ":") {
				parts = strings.SplitN(header, ":", 2)
			} else {
				continue
			}
			headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return nil
}

// extractHostname() extracts the hostname from a URL and returns it
func extractHostname(urlString string) (string, error) {
	u, err := url.Parse(urlString)
	if err != nil {
		return "", err
	}
	return u.Hostname(), nil
}

// print result constructs output lines and sends them to the results chan
func printResult(link string, sourceName string, showSource bool, results chan string, e *colly.HTMLElement) {
	result := e.Request.AbsoluteURL(link)
	if result != "" {
		if showSource {
			result = "[" + sourceName + "] " + result
		}
		results <- result
	}
}

// print result constructs output lines and sends them to the results chan
func printReflection(link string, sourceName string, showSource bool, results chan string) {
	result := link
	if result != "" {
		if showSource {
			result = "[" + sourceName + "] " + result
		}
		results <- result
	}
}

// returns whether the supplied url is unique or not
func isUnique(url string) bool {
	_, present := sm.Load(url)
	if present {
		return false
	}
	sm.Store(url, true)
	return true
}

// returns a random alphabetical string of provided length
func randomString(length int) string {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
