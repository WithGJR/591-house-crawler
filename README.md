# 591 House Crawler (591房屋交易網 店面出租 爬蟲)

URL:  [https://store.591.com.tw/index.php](https://store.591.com.tw/index.php)

## Usage

1. `go build main.go`
2. `./main -region [region-id] -output [output-type]`

請前往上述 URL 查看你要的地區的 id 為何。（使用瀏覽器的「檢查元素」功能）

output type 可選擇 `json` 或 `csv`，最後結果會根據你指定的 output type 被存成 `result.csv` 或是 `result.json`。 
