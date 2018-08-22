package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

var (
	// jsonFile for targers with urls, methods and names
        // change path to full path or run in same folder with targets.json
	jsonFile = "targets.json"

	// map of methods definitions
	methods = make(map[string]string)

	// lists of urls and results
	targets []*urls
	results []*result

	// use goroutines to make it fast
	wg sync.WaitGroup

	// regexp matching numbers and dots (maybe extend to commas)
	priceReg = regexp.MustCompile("[^0-9.]+")
)

// own structure of targets.json
// it can be different by having more URL and Method under the same Name
type urls struct {
	URL    string `json:"url"`
	Name   string `json:"name"`
	Method string `json:"method"`
}

// results return same Name and Method by the extracted price
type result struct {
	Price  string
	Name   string
	Method string
}

// scraping function
func eachScrape(url, name, method string) {
	// wait for func to finish before closing the wait group for this call
	defer wg.Done()

	// fetch URL
	doc, err := goquery.NewDocument(url)
	// generic error trap
	if err != nil {
		log.Fatal(err)
	}

	// this will return wrong things if the method definition is wrong
	doc.Find(methods[method]).Each(func(index int, item *goquery.Selection) {
		// append to results the extracted price
		results = append(results, &result{
			Method: method,
			Name:   name,
			Price:  priceReg.ReplaceAllString(item.Text(), ""),
		})
	})
}

// wrapper to goroutice each target
func postScrape() {
	wg.Add(len(targets))
	for _, t := range targets {
		go eachScrape(t.URL, t.Name, t.Method)
	}
	wg.Wait()
}

func main() {
	// definition of each website method to extract the price
	methods["netxl"] = "body div.wrapper div.product-top div div.left div.product-cost p span:nth-child(1)"
	methods["wifi-stock"] = "#det_price table tbody tr:nth-child(2) td:nth-child(2)"
	methods["senetic"] = "#lewa div div.product-main-info div.product-prices div:nth-child(8) span:nth-child(1)"
	methods["senetic2"] = "#lewa div div.product-main-info div.product-prices div:nth-child(5) span:nth-child(1)"
	methods["4gon"] = "#price_gross strong"
	methods["comms-express"] = "#vat"
	// methods["alternate"] = "#pageContent table:nth-child(3) tbody tr td.productShort div:nth-child(3) p.price"
	methods["voipon"] = "#price_gross strong"

	// load json if present
	if _, err := os.Stat(jsonFile); err == nil {
		file, err := ioutil.ReadFile(jsonFile)
		if err == nil {
			json.Unmarshal(file, &targets)
		}
	}

	// fetch results
	postScrape()

	// basic output of results
	for _, r := range results {
		fmt.Printf("%s,%s,%s\n", r.Name, r.Price, r.Method)
	}
}
