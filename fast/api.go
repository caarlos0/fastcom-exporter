package fast

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sync/atomic"
	"time"

	"github.com/prometheus/common/log"
	"golang.org/x/sync/errgroup"
)

const (
	baseURL    = "https://fast.com"
	defaultURL = "https://api.fast.com/netflix/speedtest"
	userAgent  = "caarlos0/fastcom-exporter/v1"
)

var (
	urlRE   = regexp.MustCompile(`(?U)"url":"(.*)"`)
	jsRE    = regexp.MustCompile(`app-.*\.js`)
	tokenRE = regexp.MustCompile(`token:"[[:alpha:]]*"`)
)

func Measure() (float64, error) {
	var wg errgroup.Group
	var sumBytes int64
	urls := findURLs()

	start := time.Now()
	for _, url := range urls {
		url := url
		wg.Go(func() error {
			bytes, err := doMeasure(url)
			atomic.AddInt64(&sumBytes, bytes)
			return err
		})
	}

	err := wg.Wait()
	return float64(sumBytes) / time.Since(start).Seconds(), err
}

func doMeasure(url string) (int64, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return io.Copy(io.Discard, resp.Body)
}

func findURLs() []string {
	token := getToken()

	url := fmt.Sprintf("https://api.fast.com/netflix/speedtest/v2?https=true&token=%s&urlCount=5", token)
	// fmt.Printf("url=%s\n", url)
	log.Debugf("getting url list from %s", url)

	jsonData, err := getPage(url)
	if err != nil {
		log.Errorf("error getting fast page: %s: %s", url, err)
	}

	var urls []string
	for _, url := range urlRE.FindAllStringSubmatch(string(jsonData), -1) {
		urls = append(urls, url[1])
		log.Debugf("url: %s", url[1])
	}

	return urls
}

func getToken() string {
	fastBody, err := getPage(baseURL)
	if err != nil {
		log.Errorf("error getting fast page: %s: %s", baseURL, err)
	}

	scriptNames := jsRE.FindAllString(string(fastBody), 1)
	scriptURL := fmt.Sprintf("%s/%s", baseURL, scriptNames[0])
	log.Debugf("trying to get fast api token from %s", scriptURL)

	scriptBody, err := getPage(scriptURL)
	if err != nil {
		log.Errorf("error getting fast page: %s: %s", scriptURL, err)
	}
	tokens := tokenRE.FindAllString(string(scriptBody), 1)

	if len(tokens) > 0 {
		token := tokens[0][7 : len(tokens[0])-1]
		log.Debugf("token found: %s", token)
		return token
	}
	log.Warn("no token found")
	return ""
}

func getPage(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// type measurementTransport struct {
// 	read        int64
// 	start, stop time.Time
// 	decorated   http.RoundTripper
// }

// func (t *measurementTransport) RoundTrip(req *http.Request) (*http.Response, error) {
// 	t.start = time.Now().UTC()
// 	resp, err := t.decorated.RoundTrip(req)
// 	t.stop = time.Now().UTC()
// 	t.read += resp.ContentLength
// 	return resp, err
// }

// func (t *measurementTransport) Bytes() int64 {
// 	return t.read
// }

// func (t *measurementTransport) Seconds() float64 {
// 	return t.stop.Sub(t.start).Seconds()
// }
