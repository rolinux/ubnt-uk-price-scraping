package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
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
	targets []struct {
		Name   string `json:"name"`
		Method string `json:"method"`
		VAT    bool   `json:"vat"`
		Urls   []struct {
			URL  string `json:"url"`
			Name string `json:"name"`
		} `json:"urls"`
	}
	results []*result

	// use goroutines to make it fast
	wg sync.WaitGroup

	// regexp matching numbers and dots (maybe extend to commas)
	priceReg       = regexp.MustCompile("[^0-9.]+")
	priceRegEndDot = regexp.MustCompile("\\.$")
)

/*
// own structure of targets.json
// it can be different by having more URL and Method under the same Name
type urls struct {
	URL    string `json:"url"`
	Name   string `json:"name"`
	Method string `json:"method"`
}
*/

// results return same Name and Method by the extracted price
type result struct {
	PriceWithoutVAT string
	PriceWithVAT    string
	Name            string
	Method          string
}

// scraping function
func eachScrape(url, nameUrl, nameMethod, method string, vat bool) {
	// wait for func to finish before closing the wait group for this call
	defer wg.Done()

	// fetch URL
	doc, err := goquery.NewDocument(url)
	// generic error trap
	if err != nil {
		log.Fatal(err)
	}

	// this will return wrong things if the method definition is wrong
	doc.Find(method).Each(func(index int, item *goquery.Selection) {
		price, err := strconv.ParseFloat(priceRegEndDot.ReplaceAllString(priceReg.ReplaceAllString(item.Text(), ""), ""), 64)
		if err != nil {
			log.Fatal(err)
		}
		priceWithoutVAT := ""
		priceWithVAT := ""
		if vat {
			priceWithoutVAT = fmt.Sprintf("%.2f", (price / 1.2))
			priceWithVAT = fmt.Sprintf("%.2f", price)
		} else {
			priceWithoutVAT = fmt.Sprintf("%.2f", price)
			priceWithVAT = fmt.Sprintf("%.2f", (price * 1.2))
		}
		// append to results the extracted price
		results = append(results, &result{
			Method:          nameMethod,
			Name:            nameUrl,
			PriceWithoutVAT: priceWithoutVAT,
			PriceWithVAT:    priceWithVAT,
		})
	})
}

// wrapper to goroutice each target
func postScrape() {
	totalURls := 0
	for _, t := range targets {
		totalURls += len(t.Urls)
	}

	wg.Add(totalURls)
	for _, t := range targets {
		for _, u := range t.Urls {
			go eachScrape(u.URL, u.Name, t.Name, t.Method, t.VAT)
		}
	}
	wg.Wait()
}

func main() {
	// definition of each website method to extract the price
	// methods["netxl"] = "body div.wrapper div.product-top div div.left div.product-cost p span:nth-child(1)"
	// methods["wifi-stock"] = "#det_price table tbody tr:nth-child(2) td:nth-child(2)"
	// methods["senetic"] = "#lewa div div.product-main-info div.product-prices div:nth-child(8) span:nth-child(1)"
	// methods["senetic2"] = "#lewa div div.product-main-info div.product-prices div:nth-child(5) span:nth-child(1)"
	// methods["4gon"] = "#price_gross strong"
	// methods["comms-express"] = "#vat"
	// methods["alternate"] = "#pageContent table:nth-child(3) tbody tr td.productShort div:nth-child(3) p.price"
	// methods["voipon"] = "#price_gross strong"

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
		fmt.Printf("%s,%s,%s,%s\n", r.Name, r.PriceWithoutVAT, r.PriceWithVAT, r.Method)
	}
}
