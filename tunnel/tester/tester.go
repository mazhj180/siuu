package tester

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"siuu/logger"
	"siuu/tunnel/monitor"
	"siuu/tunnel/proxy"
	"sync"
	"sync/atomic"
	"time"
)

const maxTid = 0x400

var (
	counter int32
)

func genTid() string {
	for {
		cur := atomic.LoadInt32(&counter)
		newVal := (cur + 1) % (maxTid + 1)
		if atomic.CompareAndSwapInt32(&counter, cur, newVal) {
			return fmt.Sprintf("test-tid-%#X", newVal)
		}
	}
}

type Interface interface {
	Test()
	GetResult() (map[string]float64, error)
}

func NewTester(url, host string, proxies []proxy.Proxy) Interface {
	return &tester{
		proxies: proxies,
		url:     url,
		host:    host,
		res:     make(chan *result, len(proxies)+1),
	}
}

type tester struct {
	proxies []proxy.Proxy
	url     string
	host    string
	res     chan *result
	wg      sync.WaitGroup
}

func (t *tester) Test() {
	for _, v := range t.proxies {
		t.wg.Add(1)

		go func(prx proxy.Proxy) {
			defer t.wg.Done()

			// set timeout 5s
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			req, err := http.NewRequest("GET", t.url, nil)
			if err != nil {
				logger.SError("testing- [%s]-- err : %s", prx.GetName(), err)
				return
			}

			req.Header.Set("Host", t.host)
			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36")
			req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
			req.Header.Set("Accept-Language", "en-US,en;q=0.9")
			req.Header.Set("Accept-Encoding", "gzip, deflate, br")
			req.Header.Set("Connection", "keep-alive")
			req.Header.Set("Upgrade-Insecure-Requests", "1")

			r := proxy.NewHttpReader(req)
			w := &bytes.Buffer{}

			testConn := &TestConn{
				r,
				w,
			}
			timer := monitor.Timer{}
			timer.Start()
			m := monitor.Watch(testConn)
			cli := &proxy.Client{
				Sid:   genTid(),
				Conn:  m,
				Host:  t.host,
				Port:  uint16(443),
				IsTLS: true,
				Req:   nil,
			}

			done := make(chan struct{})
			go func() {
				if err = prx.Act(cli); err != nil {
					safeSend(t.res, &result{prx: prx.GetName(), cost: -1})
				} else {
					timer.Stop()
					cost := timer.Cost()
					up, down := m.SpendTime()
					safeSend(t.res, &result{prx: prx.GetName(), cost: cost - up - down})
				}
				close(done)
			}()

			select {
			case <-ctx.Done():
				t.res <- &result{prx: prx.GetName(), cost: -1}
			case <-done:
			}

		}(v)
	}

	t.wg.Wait()
	t.res <- &result{cost: math.NaN()}
	close(t.res)
}

func (t *tester) GetResult() (map[string]float64, error) {
	if len(t.res) == 0 {
		return nil, errors.New("testing result is empty")
	}
	resultMap := make(map[string]float64)
	for {
		res := <-t.res
		if math.IsNaN(res.cost) {
			break
		}
		resultMap[res.prx] = res.cost
	}
	return resultMap, nil
}

func safeSend(ch chan<- *result, res *result) {
	defer func() {
		if recover() != nil {
			logger.SWarn("attempted to send to a closed channel : [%s]", res.prx)
		}
	}()
	ch <- res
}

type result struct {
	prx  string
	cost float64
}
