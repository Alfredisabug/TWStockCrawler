package main

import (
	"log"
	"strings"

	"bytes"
	"encoding/csv"
	"io/ioutil"
	"net/http"
	"os"

	"golang.org/x/text/transform"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/encoding/traditionalchinese"
)

func init() {
	log.SetFlags(log.Ldate | log.Lshortfile)
}
func main() {
	//	FindStockNumber()
	FindStockNumberBySTDLib()
}

// 'dtype', '國際證券辨識號碼', '上市日', '市場別', '產業別', 'CFI', '備註'
type Stock struct {
	Dtype            string
	IdetifyNumber    string
	Date             string
	MarketCategory   string
	IndustryCategory string
	CFI              string
	Note             string
}

func FindStockNumberBySTDLib() {
	log.Print("Visiting http://isin.twse.com.tw/isin/C_public.jsp?strMode=2")
	resp, err := http.Get("http://isin.twse.com.tw/isin/C_public.jsp?strMode=2")
	if err != nil {
		log.Println("Get url error.")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("get content failed status code is %d. \n", resp.StatusCode)
		return
	}

	BodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("HTML body read error.")
		return
	}
	log.Println("Get HTML.")
	// 將抓到的html網頁資訊交給goquery解析
	htmlDoc, err := goquery.NewDocumentFromReader(bytes.NewReader([]byte(BodyBytes)))
	if err != nil {
		log.Println("Goquery parse fail.")
		return
	}

	var StockList []Stock
	htmlDoc.Find(".h4").Each(func(j int, contentSelection *goquery.Selection) {
		var stock Stock
		contentSelection.Find("tr[align!='center']").Each(func(_ int, tr *goquery.Selection) {
			tr.Find("td[colspan!='7']").Each(func(i int, s *goquery.Selection) {
				// 轉換編碼
				decodeStr, _ := DecodeBig5([]byte(s.Text()))

				data := string(decodeStr)
				number := i % 8
				switch number {
				case 0:
					stock.Dtype = data
				case 1:
					stock.IdetifyNumber = data
				case 2:
					stock.Date = data
				case 3:
					stock.MarketCategory = data
				case 4:
					stock.IndustryCategory = data
				case 5:
					stock.CFI = data
				case 6:
					stock.Note = data
				}
			})
			StockList = append(StockList, stock)
		})

	})
	f, err := os.Create("./StockList.csv")
	if err != nil {
		return
	}

	f.WriteString("\xEF\xBB\xBF")
	defer f.Close()
	fw := csv.NewWriter(f)
	// 'dtype', '國際證券辨識號碼', '上市日', '市場別', '產業別', 'CFI'
	fw.Write([]string{"有價證券代號", "名稱", "國際證券辨識號碼(ISIN Code)", "上市日", "市場別", "產業別", "CFI", "備註"})
	sep := string([]byte{227, 128, 128})
	for _, i := range StockList {
		SplitStr := strings.Split(i.Dtype, sep)
		if len(SplitStr) > 1 {
			fw.Write([]string{
				SplitStr[0],
				SplitStr[1],
				i.IdetifyNumber,
				i.Date,
				i.MarketCategory,
				i.IndustryCategory,
				i.CFI,
				i.Note,
			})
		} else {
			fw.Write([]string{
				i.Dtype,
				" ",
				i.IdetifyNumber,
				i.Date,
				i.MarketCategory,
				i.IndustryCategory,
				i.CFI,
				i.Note,
			})
		}

	}
	fw.Flush()
	log.Println("Write to .csv success.")
}

//convert BIG5 to UTF-8
func DecodeBig5(s []byte) ([]byte, error) {
	I := bytes.NewReader(s)
	O := transform.NewReader(I, traditionalchinese.Big5.NewDecoder())
	d, e := ioutil.ReadAll(O)
	if e != nil {
		return nil, e
	}
	return d, nil
}

//convert UTF-8 to BIG5
func EncodeBig5(s []byte) ([]byte, error) {
	I := bytes.NewReader(s)
	O := transform.NewReader(I, traditionalchinese.Big5.NewEncoder())
	d, e := ioutil.ReadAll(O)
	if e != nil {
		return nil, e
	}
	return d, nil
}
