package fast

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

const (
	baseURL               = "https://fast.com"
	userAgent             = "caarlos0/fastcom-exporter/v1"
	maxConcurrentRequests = 8                // from fast.com
	maxTime               = time.Second * 30 // from fast.com
)

var (
	urlRE   = regexp.MustCompile(`(?U)"url":"(.*)"`)
	jsRE    = regexp.MustCompile(`app-.*\.js`)
	tokenRE = regexp.MustCompile(`token:"[[:alpha:]]*"`)
)

func Measure() (float64, error) {
	var wg errgroup.Group
	var sumBytes int64
	var idx int32

	urls := findURLs()
	sem := semaphore.NewWeighted(maxConcurrentRequests)

	ctx, cancel := context.WithTimeout(context.Background(), maxTime)
	defer cancel()

	start := time.Now()

outer:
	for {
		select {
		case <-ctx.Done():
			break outer
		default:
			err := sem.Acquire(ctx, 1)
			if err != nil {
				if !errors.Is(err, context.DeadlineExceeded) {
					return 0, err
				}
				break outer
			}
			wg.Go(func() error {
				defer sem.Release(1)
				url := urls[int(idx)%len(urls)]
				atomic.AddInt32(&idx, 1)
				bytes, err := doMeasure(ctx, url)
				atomic.AddInt64(&sumBytes, bytes)
				return err
			})
		}
	}

	if err := wg.Wait(); err != nil && !errors.Is(err, context.DeadlineExceeded) {
		return 0, err
	}
	return float64(sumBytes) / time.Since(start).Seconds(), nil
}

func doMeasure(ctx context.Context, url string) (int64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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
	log.Debug().Msgf("getting url list from %s", url)

	jsonData, err := getPage(url)
	if err != nil {
		log.Error().Err(err).Msgf("error getting fast page %s", url)
	}

	var urls []string
	for _, url := range urlRE.FindAllStringSubmatch(string(jsonData), -1) {
		urls = append(urls, url[1])
		log.Debug().Str("url", urls[1]).Msg("got url")
	}

	return urls
}

func getToken() string {
	fastBody, err := getPage(baseURL)
	if err != nil {
		log.Error().Err(err).Msg("error getting fast page")
	}

	scriptNames := jsRE.FindAllString(string(fastBody), 1)
	scriptURL := fmt.Sprintf("%s/%s", baseURL, scriptNames[0])

	scriptBody, err := getPage(scriptURL)
	if err != nil {
		log.Error().Err(err).Msg("error getting fast page")
	}
	tokens := tokenRE.FindAllString(string(scriptBody), 1)

	if len(tokens) > 0 {
		token := tokens[0][7 : len(tokens[0])-1]
		log.Debug().Str("token", token).Msg("found token")
		return token
	}
	log.Warn().Msg("no token found")
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
