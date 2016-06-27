package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

const (
	rootURL string = "https://store.591.com.tw/"
)

var (
	regionId   int
	outputType string
)

func init() {
	flag.IntVar(&regionId, "region", 8, "To see the region id, please visit https://store.591.com.tw/index.php")
	flag.StringVar(&outputType, "output", "csv", "You can specify as json or csv")
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
	Addr  string `json:"addr"`
	Price string `json:"price"`
}

type Crawler struct {
	regionId int
	//key: pageNumber int
	infos map[int][]HouseInfo
}

type Task struct {
	PageNumber             int
	URLsToDetailedInfoPage []string
	Prices                 []string
}

type intermediateTask struct {
	index int
	url   string
	price string
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
	urlCount := len(task.URLsToDetailedInfoPage)
	finalResult := Result{
		PageNumber: task.PageNumber,
		Infos:      make([]HouseInfo, urlCount),
	}

	results := make(chan intermediateResult)
	tasks := make(chan intermediateTask)
	//start 5 workers
	for i := 0; i < 5; i++ {
		go func(tasks chan intermediateTask, results chan intermediateResult) {
			for task := range tasks {
				doc, err := goquery.NewDocument(task.url)
				if err != nil {
					log.Fatal(err)
				}

				doc.Find(".addr").Each(func(i int, s *goquery.Selection) {
					info := HouseInfo{Addr: s.Text(), Price: task.price}

					results <- intermediateResult{
						index: task.index,
						info:  info,
					}
				})
			}
		}(tasks, results)
	}

	receivedResultCount := 0
	for i := 0; (i < urlCount) || (receivedResultCount != urlCount); i++ {
		select {
		case result := <-results:
			finalResult.Infos[result.index] = result.info
			receivedResultCount++
			i--
		default:
			if i < urlCount {
				url := task.URLsToDetailedInfoPage[i]
				price := task.Prices[i]

				select {
				case tasks <- intermediateTask{index: i, url: url, price: price}:

				default:
					i--
					continue
				}
			}

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
		prices := make([]string, 0)

		doc.Find("#photolist > li").Each(func(i int, s *goquery.Selection) {
			s.Find(".address > a").Each(func(i int, s *goquery.Selection) {
				urlToHouse, _ := s.Attr("href")
				urlsToDetailedInfoPage = append(urlsToDetailedInfoPage, rootURL+urlToHouse)
			})

			s.Find(".prices > .price:nth-child(2) > span").Each(func(i int, s *goquery.Selection) {
				price := s.Text()
				prices = append(prices, price)
			})
		})

		//no more pages to be crawled
		if len(urlsToDetailedInfoPage) == 0 {
			pageNumber--
			break
		} else {
			task := Task{
				PageNumber:             pageNumber,
				URLsToDetailedInfoPage: urlsToDetailedInfoPage,
				Prices:                 prices,
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

func (c *Crawler) OutputAsCSVFile() {
	file, err := os.Create("result.csv")
	if err != nil {
		log.Fatal(err)
	}

	writer := csv.NewWriter(file)
	row := make([]string, 2)
	for page := 1; page <= len(c.infos); page++ {
		for i := 0; i < len(c.infos[page]); i++ {
			row[0] = c.infos[page][i].Addr
			row[1] = c.infos[page][i].Price

			if err := writer.Write(row); err != nil {
				log.Fatal(err)
			}
		}
	}
	writer.Flush()
}

func (c *Crawler) OutputAsJSONFile() {
	result := make([]HouseInfo, 0)
	for i := 0; i < len(c.infos); i++ {
		page := i + 1
		result = append(result, c.infos[page]...)
	}
	content, err := json.Marshal(result)
	if err != nil {
		log.Fatal(err)
	}
	ioutil.WriteFile("result.json", content, 0644)
}

func main() {
	flag.Parse()

	crawler := &Crawler{
		regionId: regionId,
		infos:    make(map[int][]HouseInfo),
	}

	crawler.Run()
	if outputType == "json" {
		crawler.OutputAsJSONFile()
	} else {
		crawler.OutputAsCSVFile()
	}
}
