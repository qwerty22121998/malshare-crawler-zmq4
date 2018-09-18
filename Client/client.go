package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin/json"
	"github.com/pebbe/zmq4"
)

const REG_EXP = `<a href=\"\d*-\d*-\d*`
const MALSHARE_URL = `http://www.malshare.com/daily/`

var HASH_TYPE = [3]string{"sha1", "sha256", ""}
var cnt = 0
var mp map[string]bool

type HashDatas struct {
	List []File
}

type HashData struct {
	Hash   string `json:"hash"`
	Type   string `json:"type"`
	Source string `json:"source"`
}

type File struct {
	File []HashData `json:"file"`
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

func getDataAll(date string, socket *zmq4.Socket) {

	url := fmt.Sprintf("http://www.malshare.com/daily/%s/malshare_fileList.%s.all.txt", date, date)

	fmt.Println("Crawl :", url)

	resp, e := http.Get(url)

	defer resp.Body.Close()

	if e != nil {
		return
	}

	red := bufio.NewScanner(resp.Body)
	listHashs := HashDatas{}
	for red.Scan() {
		currentData := red.Text()

		hashs := strings.Split(currentData, "	")
		hashdatas := make([]HashData, 0)
		types := []string{"md5", "sha1", "sha256"}
		for i := range types {
			if hashs[i] == "NULL" {
				continue
			}
			if !mp[hashs[i]] {
				mp[hashs[i]] = true
				cnt++
			}
			hashdatas = append(hashdatas, HashData{Hash: hashs[i], Type: types[i], Source: "Malshare"})
			// cnt++
		}

		listHashs.List = append(listHashs.List, File{hashdatas})
		// cnt++

		//newData, _ := json.Marshal(File{List: hashdatas})

		//socket.Send(string(newData), 0)

	}

	newData, _ := json.Marshal(listHashs.List)
	socket.Send(string(newData), 0)

}

func main() {
	fmt.Println(getLinkList())
	list := getLinkList()
	mp = make(map[string]bool)

	context, _ := zmq4.NewContext()
	pub, _ := context.NewSocket(zmq4.PUSH)
	pub.Connect("tcp://127.0.0.1:5555")

	time.Sleep(100 * time.Millisecond)
	for i, v := range list {
		if i > 3 {
			fmt.Println(cnt)
			// pub.Send("OK", 0)
			break
		}
		getDataAll(v, pub)
	}

	time.Sleep(1000 * time.Millisecond)
}
