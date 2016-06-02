package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
)

type HouseInfo struct {
	Addr string
}

type Crawler struct {
	rootURL    string
	parameters string
	infos      []HouseInfo
}

func (c *Crawler) SetRootURL(url string) {
	c.rootURL = url
}

func (c *Crawler) SetParameters(parameters string) {
	c.parameters = parameters
}

func (c *Crawler) Run() {
	doc, err := goquery.NewDocument(c.rootURL + c.parameters)
	if err != nil {
		log.Fatal(err)
	}

	urlsToDetailedInfo := make([]string, 0)
	doc.Find(".address > a").Each(func(i int, s *goquery.Selection) {
		urlToHouse, _ := s.Attr("href")
		urlsToDetailedInfo = append(urlsToDetailedInfo, c.rootURL+urlToHouse)
	})

	for i := 0; i < len(urlsToDetailedInfo); i++ {
		doc, err := goquery.NewDocument(urlsToDetailedInfo[i])
		if err != nil {
			//TODO
			continue
		}

		doc.Find(".addr").Each(func(i int, s *goquery.Selection) {
			info := HouseInfo{Addr: s.Text()}
			c.infos = append(c.infos, info)
		})
	}
}

func (c *Crawler) GetResult() []HouseInfo {
	return c.infos
}

func main() {
	crawler := &Crawler{}
	//Taichung
	crawler.SetRootURL("https://store.591.com.tw/")
	crawler.SetParameters("house-rentSale.html?storeType=1&regionid=8&search=1")
	crawler.Run()

	results := crawler.GetResult()
	for i := 0; i < len(results); i++ {
		fmt.Printf("%d: %s\n", i+1, results[i].Addr)
	}
}
