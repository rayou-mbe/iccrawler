package links

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"techtoolkit.ickey.cn/crawlerSZLC/mylog"
	"techtoolkit.ickey.cn/crawlerSZLC/ocsv"
	"techtoolkit.ickey.cn/crawlerSZLC/request"
	"techtoolkit.ickey.cn/crawlerSZLC/seed"
)

//Links export Links to main Cralwer
type Links struct {
	Cat    string
	s      *seed.Seed
	c      *ocsv.Ocsv
	l      *mylog.Log
	out    chan string
	finish chan int
	Wg     *sync.WaitGroup
}

//NewLinks newlinks
func NewLinks() *Links {
	ret := &Links{
		out:    make(chan string),
		finish: make(chan int),
	}
	ret.init()
	return ret
}

//Close release
func (l *Links) Close() {
	defer l.c.Close()
}

func (l *Links) init() {
	// Initialize the internal hosts map
	// c.hosts = make(map[string]struct{}, len(ctxs))
	l.s = new(seed.Seed)
	l.Wg = new(sync.WaitGroup)
	l.c = ocsv.NewOcsv()
	err := l.c.Init()
	if err != nil {
		log.Fatal(err)
	}

	l.l = mylog.NewLog()
	l.l.Init("szlcsc.log")

}

//GetSeedURLS get init seeds.
func (l *Links) GetSeedURLS(u string) ([]string, error) {
	//init data

	seeds, err := l.s.RootLinksGet(u)
	l.s.RootLinksGet(u)
	if err != nil {
		log.Fatal(err)
	}

	for i, v := range seeds {
		fmt.Printf("a[%d] = %s\n", i, v)
	}
	return seeds, err

}

func (l *Links) CrawlerDetailPageFromNode(href string, out chan<- interface{}) {

	//get list page from seed and request module
	fmt.Printf("###href From = %s\n", href)
	var data []request.PartNumber
	var err error

	var i int
	for i = 0; i < 3; i++ {

		data, err = l.s.GetPageListDetail(href)
		if err != nil && data != nil {
			fmt.Printf("http reques detail page:%s from nodejs err: %v, data:%v \n", href, err, data)
			continue
		} else {
			for _, v := range data {
				fmt.Printf("###>>>>>>>---->>>>ok-ok-ok will save cs data  %v\n", data)
				out <- v
			}
			//i < 3 break ok
			break
		}

	}

	fmt.Printf(">>>XXXXXAfter retry get detail page i=%d url=%s\n", i, href)
	if i >= 15 {
		fmt.Printf(">>>XXXXXAfter retry get detail page err url=%s\n", href)
		l.l.Error("err get page max" + href)
	}

}

func (l *Links) convertAndSave(d interface{}) error {
	o := new(ocsv.CSVPartNumber)

	in, ok := d.(request.PartNumber)
	if !ok {
		return errors.New("Err data error")
	}
	o.Part = in.Part
	o.Comments = in.Keyword
	o.Promaf = in.Promaf
	o.Stock = in.Stock

	//非数字
	// pattern := `[\\d+$]`
	for i, v := range in.Steps {
		if i < 10 {
			if 0 == i {
				o.PurchaseNum1 = v
			} else if 1 == i {
				o.PurchaseNum2 = v
			} else if 2 == i {
				o.PurchaseNum3 = v
			} else if 3 == i {
				o.PurchaseNum4 = v
			} else if 4 == i {
				o.PurchaseNum5 = v
			} else if 5 == i {
				o.PurchaseNum6 = v
			} else if 6 == i {
				o.PurchaseNum7 = v
			} else if 7 == i {
				o.PurchaseNum8 = v
			} else if 8 == i {
				o.PurchaseNum9 = v
			} else if 9 == i {
				o.PurchaseNum10 = v
			}
		}
	}

	for i, v := range in.Prices {
		if i < 10 {
			if 0 == i {
				o.PurchaseUnitPrice1 = v
			} else if 1 == i {
				o.PurchaseUnitPrice2 = v
			} else if 2 == i {
				o.PurchaseUnitPrice3 = v
			} else if 3 == i {
				o.PurchaseUnitPrice4 = v
			} else if 4 == i {
				o.PurchaseUnitPrice5 = v
			} else if 5 == i {
				o.PurchaseUnitPrice6 = v
			} else if 6 == i {
				o.PurchaseUnitPrice7 = v
			} else if 7 == i {
				o.PurchaseUnitPrice8 = v
			} else if 8 == i {
				o.PurchaseUnitPrice9 = v
			} else if 9 == i {
				o.PurchaseUnitPrice10 = v
			}
		}
	}
	l.c.Append(o)
	return nil
}

//Storages channel one the last channel close done channal for singal all channal done
func (l *Links) Storages(in <-chan interface{}, done chan<- struct{}) {
	//consurmer
	defer close(done)
	for {
		select {
		case v, ok := <-in:
			//do resualt.\
			if ok {
				fmt.Printf("receive  one chan storage %v.....", v)
				l.convertAndSave(v)
			} else {
				fmt.Println("recieve all chan storage....")
				return
			}
		}
	}
}

//ListPage out channel for list page. first output channel
func (l *Links) detailPage(out chan<- interface{}, in <-chan string) {
	//consurmer
	defer close(out)
	for {
		select {
		case page, ok := <-in:
			//do resualt.\
			if ok {
				fmt.Println("received one page: ", page)
				l.CrawlerDetailPageFromNode(page, out)
			} else {
				fmt.Println("received all chan pages")
				return
			}
		}
	}
}

//CrawlerCatListFromNode page
func (l *Links) CrawlerCatListFromNode(url string, out chan<- string) error {
	// l.Wg.Add(1)

	pURL, max, sURL, err := l.s.GetPagesFromNodeJS(url)
	if err != nil {
		fmt.Printf("CrawlerCatListFromNode url:%s from nodejs err: %v\n", url, err)
		return err
	}

	fmt.Printf("######## >>>> GetPagesFromNodeJS pre: %s, max: %d, sufix: %s len\n", pURL, max, sURL)
	if max > 1 {
		// fmt.Printf("######## >>>>>> out<-data[%d] = %s\n", max, url)
		for i := 1; i <= max; i++ {
			s := fmt.Sprintf("%s%d%s", pURL, i, sURL)
			fmt.Printf("########### >>>>After send to l.out<- list: %s,  %v\n", s, max)
			out <- s
		}

	} else if max == 1 {
		s := fmt.Sprintf("%s%d%s", pURL, max, sURL)
		// fmt.Printf("###After send to l.out<- list: %s  %v\n", s, max)
		out <- s
	} else {
		//error
		return errors.New("max=0")
	}

	return nil
}

//ListPage out channel for list page. first output channel
func (l *Links) detailURLS(out chan<- string, in <-chan string) {
	//consurmer
	defer close(out)
	for {
		select {
		case href, ok := <-in:
			//do resualt.\
			if ok {
				fmt.Printf("received chan: list url= %s\n", href)
				// go l.CrawlerCatListFromNode(href, out)
				//if crawler max page num err retry 3 times.
				var i int
				for i = 0; i < 15; i++ {
					err := l.CrawlerCatListFromNode(href, out)
					if err == nil {
						break
					}
				}
				fmt.Printf(">>>XXXXXAfter retry get max page err i=%d,url=%s\n", i, href)
				if i >= 15 {
					fmt.Printf(">>>XXXXXAfter retry get max page err url=%s\n", href)
					l.l.Error("err get page max" + href)
				}

			} else {
				//channel closed
				fmt.Println("received all chan list")
				return
			}
		}
	}
}

//ListURLS out channel for list page. first output channel
func (l *Links) ListURLS(urls []string, out chan<- string) error {

	for i, url := range urls {
		fmt.Printf("XXXX first i=%d url=%s", i, url)
		if len(url) > 0 {
			out <- url
		}
	}
	return nil
}

/*CrawlerSZLC  crawler list links from root seed []string  array.
  urls root seeds.
  out chan  put real per ic url into out channal
*/
func (l *Links) CrawlerSZLC(urls []string) error {
	start := time.Now().Unix()

	list := make(chan string)
	pages := make(chan string)
	storages := make(chan interface{})

	done := make(chan struct{})

	//input, out-chan, out-chan

	go l.detailURLS(pages, list)
	go l.detailPage(storages, pages)
	go l.Storages(storages, done)

	defer close(storages)
	defer l.l.Close()
	defer close(list)

	for i := 0; i < 100; i++ {
		fmt.Println("Loop once, Your turn.")
		begin := time.Now()
		l.ListURLS(urls, list)

		dur := time.Since(begin).Seconds()
		//each day
		if dur < 86400 {
			select {
			case <-time.After(time.Second * time.Duration(dur)):
			}
		}

	}

	//wait all finished.

	<-done

	end := time.Now().Unix()
	fmt.Printf("All Done:  spend - time %d\n", end-start)
	return nil
}
