package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"log"
	"strconv"
)

const (
	rootURL string = "https://store.591.com.tw/"
)

var (
	regionId int
)

func init() {
	flag.IntVar(&regionId, "region", 8, "To see the region id, please visit https://store.591.com.tw/index.php")
}

func (c *Crawler) getPageURL(pageNumber int) string {
	if pageNumber == 1 {
		return rootURL + "house-rentSale.html?storeType=1&regionid=" + strconv.Itoa(c.regionId) + "&search=1"
	} else {
		firstRow := 20 * (pageNumber - 1)
		return rootURL + "index.php?firstRow=" + strconv.Itoa(firstRow) + "&storeType=1&regionid=" + strconv.Itoa(c.regionId) + "&search=1&module=house&action=rentSale"
	}
}

type HouseInfo struct {
	Addr string `json:"addr"`
}

type Crawler struct {
	regionId int
	infos    map[int][]HouseInfo
}

type Task struct {
	PageNumber             int
	URLsToDetailedInfoPage []string
}

type intermediateTask struct {
	index int
	url   string
}

type Result struct {
	PageNumber int
	Infos      []HouseInfo
}

type intermediateResult struct {
	index int
	info  HouseInfo
}

func (c *Crawler) handle(task Task, output chan Result) {
	finalResult := Result{
		PageNumber: task.PageNumber,
		Infos:      make([]HouseInfo, len(task.URLsToDetailedInfoPage)),
	}

	results := make(chan intermediateResult)
	tasks := make(chan intermediateTask)
	//start 10 workers
	for i := 0; i < 10; i++ {
		go func(tasks chan intermediateTask, results chan intermediateResult) {
			for task := range tasks {
				doc, err := goquery.NewDocument(task.url)
				if err != nil {
					log.Fatal(err)
				}

				doc.Find(".addr").Each(func(i int, s *goquery.Selection) {
					info := HouseInfo{Addr: s.Text()}

					results <- intermediateResult{
						index: task.index,
						info:  info,
					}
				})
			}
		}(tasks, results)
	}

	for i := 0; i < len(task.URLsToDetailedInfoPage); i++ {
		url := task.URLsToDetailedInfoPage[i]

		select {
		case result := <-results:
			finalResult.Infos[result.index] = result.info
			i--
		case tasks <- intermediateTask{index: i, url: url}:
			continue
		}

	}
	close(tasks)

	output <- finalResult
}

func (c *Crawler) Run() {
	output := make(chan Result, 30)
	pageNumber := 0

	for {
		pageNumber++
		doc, err := goquery.NewDocument(c.getPageURL(pageNumber))

		if err != nil {
			log.Fatal(err)
		}

		urlsToDetailedInfoPage := make([]string, 0)
		doc.Find(".address > a").Each(func(i int, s *goquery.Selection) {
			urlToHouse, _ := s.Attr("href")
			urlsToDetailedInfoPage = append(urlsToDetailedInfoPage, rootURL+urlToHouse)
		})

		//no more pages to be crawled
		if len(urlsToDetailedInfoPage) == 0 {
			pageNumber--
			break
		} else {
			task := Task{
				PageNumber:             pageNumber,
				URLsToDetailedInfoPage: urlsToDetailedInfoPage,
			}
			fmt.Printf("Start crawling page %d\n", pageNumber)
			go c.handle(task, output)
		}
	}

	for len(c.infos) != pageNumber {
		result := <-output
		c.infos[result.PageNumber] = result.Infos
		fmt.Printf("Page %d is finished.\n", result.PageNumber)
	}
}

func (c *Crawler) OutputJSONAsFile() {
	for pageNumber, infos := range c.infos {
		content, err := json.Marshal(infos)
		if err != nil {
			log.Fatal(err)
		}
		ioutil.WriteFile("page"+strconv.Itoa(pageNumber)+".json", content, 0644)
	}
}

func main() {
	flag.Parse()

	crawler := &Crawler{
		regionId: regionId,
		infos:    make(map[int][]HouseInfo),
	}

	crawler.Run()
	crawler.OutputJSONAsFile()

}
