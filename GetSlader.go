package main
import (
	"os"
	"io"
	"fmt"
	"bufio"
	"io/ioutil"
	"strings"
	"errors"
	"math/rand"
	"net/http"
	"github.com/PuerkitoBio/goquery"
	"time"
)

func main() {

	sladerURL := os.Args[1]

	usrag := getRandomUserAgent()
	cli := createHTTPClient(usrag)
	
	solutionsURL := getSolutionsLocation(cli, sladerURL)
	html := getSolutions(cli, solutionsURL)
	
	fmt.Println(html)
}


func getRandomUserAgent() string {

	f, err := os.Open("user-agents.txt")
	check(err)

	defer f.Close()

	fi, err := f.Stat()
	check(err)

	rand.Seed(time.Now().UnixNano())

	rpos := rand.Int63n(fi.Size())
	_, err = f.Seek(rpos, 0)
	check(err)

	re := bufio.NewReader(f)
	_, _ = re.ReadString('\n')
	usrag, err := re.ReadString('\n')

	if err == io.EOF {
		_, err = f.Seek(0,0)
		check(err)
		re = bufio.NewReader(f)
		usrag, err = re.ReadString('\n')
	}

	check(err)

	return strings.TrimSuffix(usrag, "\n")
}

func createHTTPClient(usrag string) *http.Client {

	cli := http.DefaultClient

	rt := WithHeader(cli.Transport)

	rt.Set("User-Agent", usrag)
	rt.Set("Host", "www.slader.com")
	rt.Set("Accept", "*/*")
	rt.Set("Accept-Language", "en-US,en;q=0.5")
	rt.Set("X-Requested-With", "XMLHttpRequest")
	rt.Set("DNT", "1")
	rt.Set("Connection", "keep-alive")
	rt.Set("Pragma", "no-cache")
	rt.Set("Cache-Control", "no-cache")

	
	cli.Transport = rt

	return cli

}


func getSolutionsLocation(cli *http.Client, sladerURL string) string {

	resp, err := cli.Get(sladerURL)
	check(err)


	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		check(errors.New("Network Error: reponse code of" + string(resp.StatusCode)))
	}

	doc, err := goquery.NewDocumentFromResponse(resp)
	check(err)


	el := doc.Find(".solutions-list.unloaded.reloadable").First()
	val, has := el.Attr("data-url")

	if has == false {
		check(errors.New("Parsing Error: no solution url found"))
	}

	solutionsURL := "https://www.slader.com" + val

	return solutionsURL
}

func getSolutions(cli *http.Client, solutionsURL string) string {

	resp, err := cli.Get(solutionsURL)
	check(err)


	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		check(errors.New("Network Error: reponse code of " + http.StatusText(resp.StatusCode)))
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	check(err)
	
	html := string(bytes)


	return html
}


func check(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// functions for client header
type withHeader struct {
        http.Header
        rt http.RoundTripper
}

func WithHeader(rt http.RoundTripper) withHeader {
        if rt == nil {
                rt = http.DefaultTransport
        }

        return withHeader{Header: make(http.Header), rt: rt}
}

func (h withHeader) RoundTrip(req *http.Request) (*http.Response, error) {
        for k, v := range h.Header {
                req.Header[k] = v
        }

        return h.rt.RoundTrip(req)
}
