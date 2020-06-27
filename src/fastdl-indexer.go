package main

import (

	"net/http"
	"net/http/cookiejar"

	"fmt"
	"os"
	"io/ioutil"
	"strings"
	"time"
	"math/rand"
	"sync"

	// xpath / html
	"github.com/antchfx/htmlquery"
)

var user_agent string = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:74.0) Gecko/20100101 Firefox/74.0"; // contains the user agent


func SendGetReqBytes(client *http.Client, url string) []byte {
	req , err := http.NewRequest("GET",url,nil);
	if err != nil {
		fmt.Printf("Failed to get request url for %s",url);
		os.Exit(1);			
	}

	// mask the user agent and setup the correct content type
	req.Header.Add("User-Agent", user_agent);
	req.Header.Add("X-Requested-With","XMLHttpRequest"); // need this to pull the page	


	// repeat this until we can read it cause the page is not giving back data sometimes
	var body []byte;

	retry := 0;

	for  {

		retry += 1;

		if retry > 10 {
			fmt.Printf("retry limit reached for %s\n",url);
			os.Exit(1);
		}

		// send our get req
		resp, err := client.Do(req)
		if(err != nil) {
			//fmt.Printf("Failed to send get req for %s are you connected to the Internet?\n",url);
			time.Sleep(time.Duration(rand_time(1))*time.Second);
			continue;
		}
		
		// read out the response
		body, err = ioutil.ReadAll(resp.Body)
		if(err != nil) {
			//fmt.Printf("Failed to read get response for %s\n",url );
			time.Sleep(time.Duration(rand_time(1))*time.Second);
			continue;
		}
		resp.Body.Close();
		break;
	}

	return body;
}

func SendGetReq(client *http.Client, url string) string {

	body := SendGetReqBytes(client,url);

	// sleep for "load times"
	time.Sleep(time.Duration(rand_time(2)) * time.Second);
	return string(body);		
}







func fileExists(filename string) bool {
    info, err := os.Stat(filename);
    if os.IsNotExist(err) {
        return false;
    }
    return !info.IsDir();
}


func WriteToFileBytes(filename string, buf []byte) {
    fp, err := os.Create(filename);
    if err != nil {
		fmt.Printf("failed to create file: %v",err);
		os.Exit(1);
    }
    defer fp.Close();

    _, err = fp.Write(buf);
    if err != nil {
        fmt.Printf("failed to write string: %v",err);
		os.Exit(1);
    }	
}

func rand_time(t int) float64 {
	x := float64(3+rand.Intn(t)) + rand.Float64();
	return x;
}

var wg sync.WaitGroup;


func DownloadMap(client *http.Client,link string, fast_dl_link string) {
	defer wg.Done();

	fmt.Printf("fetching %s\n",link);

	arr := SendGetReqBytes(client,fast_dl_link + link);
	WriteToFileBytes(link,arr);
	fmt.Printf("wrote: %s\n",link);
	time.Sleep(time.Duration(rand_time(2)) * time.Second);
}
func main() {


	args := os.Args;


	if(len(args) == 1) {
		fmt.Printf("usage: %s <fastdl maps link>\n",args[0]);
		return;
	}

	// init our cookie jar + client
	jar, _ := cookiejar.New(nil);
	
	client := &http.Client{
		Jar: jar,
		//Transport: torTrasnport,
	}	


	fast_dl_link := args[1];
	fmt.Printf("Scraping %s\n",fast_dl_link);

	// pull the file list
	resp := SendGetReq(client,fast_dl_link);
	doc, err := htmlquery.Parse(strings.NewReader(string(resp)));
	if(err != nil) {
		fmt.Println("Failed to parse the html for xpath");
		os.Exit(1);
	}
	
	// get all the hrefs with xpath
	list := htmlquery.Find(doc,"//a[@href]");

	
	maps := make([]string,0)
	
	// filter by .bsp files 
	for i := 0; i < len(list); i++ {
		link := htmlquery.SelectAttr(list[i],"href");
		if(strings.Contains(link,".bsp") && !fileExists(link)) {
			maps = append(maps,link);
		}
	}
	
	wg.Add(len(maps));
	
	// use links and dl every map to a file
	for i := 0; i < len(maps); i++ {
		go DownloadMap(client,maps[i],fast_dl_link);
	}
	
	wg.Wait();
}