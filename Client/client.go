package main

import (
	"regexp"
	"net/http"
	"fmt"
	"io/ioutil"
	"strings"
	"bufio"
	"zmq4"
	"time"
	"github.com/gin-gonic/gin/json"
)

const REG_EXP = `<a href=\"\d*-\d*-\d*`
const MALSHARE_URL = `http://www.malshare.com/daily/`

var HASH_TYPE = [3]string{".sha1", ".sha256", ""}

type HashData struct {
	Hash    string    `json:"hash"`
	Type    string    `json:"type"`
	Created time.Time `json:"created"`
	Desc    string    `json:"desc"`
}

func getBody() string {

	fmt.Println("Connecting to ", MALSHARE_URL)
	resp, e := http.Get(MALSHARE_URL)

	if e != nil {
		fmt.Println(e.Error())
		return ""
	}

	defer resp.Body.Close()
	defer fmt.Println("Closing Connect")

	bText, e := ioutil.ReadAll(resp.Body)

	if e != nil {
		fmt.Println(e.Error())
		return ""
	}
	return string(bText)

}

func getLinkList() []string {

	reg, _ := regexp.Compile(REG_EXP)
	body := getBody()
	//fmt.Println(body)

	list := reg.FindAllString(body, -1)

	for i, v := range list {
		list[i] = strings.Replace(v, `<a href="`, "", -1)
	}
	return list

}

func getData(date, dataType string, socket *zmq4.Socket) {

	url := "http://www.malshare.com/daily/{{date}}/malshare_fileList.{{date}}{{dataType}}.txt"
	url = strings.Replace(strings.Replace(url, "{{date}}", date, -1), "{{dataType}}", dataType, -1)

	fmt.Println("Crawl :", url)

	if dataType == "" {
		dataType = "md5"
	}

	resp, e := http.Get(url)

	defer resp.Body.Close()

	if e != nil {
		return
	}

	red := bufio.NewScanner(resp.Body)

	for red.Scan() {
		currentData := red.Text()

		newData, _ := json.Marshal(HashData{Hash: currentData, Type: dataType, Created: time.Now(), Desc: "Crawl from malshare"})
		socket.Send(string(newData),0)
		fmt.Println(currentData, dataType)
	}

}

func main() {
	//fmt.Println(getLinkList())
	list := getLinkList()

	context, _ := zmq4.NewContext()

	pub, _ := context.NewSocket(zmq4.PUB)

	pub.Bind("tcp://*:5555")

	for _, v := range list {
		//getData(v, `sha1`)
		for _, t := range HASH_TYPE {
			getData(v, t, pub)
		}
		return
	}

}
