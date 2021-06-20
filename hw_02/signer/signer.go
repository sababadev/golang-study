package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

const thxCount = 6

func ExecutePipeline(jj ...job) {
	in := make(chan interface{}, MaxInputDataLen)
	wg := &sync.WaitGroup{}

	for _, j := range jj {
		out := make(chan interface{})

		wg.Add(1)
		go func(j job, in, out chan interface{}) {
			defer close(out)
			j(in, out)
			wg.Done()
		}(j, in, out)

		in = out
	}
	wg.Wait()
}

func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	mu := sync.Mutex{}

	for i := range in {
		i := strconv.Itoa(i.(int))
		neck := make(chan string, 1)

		wg.Add(1)
		go func() {
			defer wg.Done()

			go func(neck chan string) {
				mu.Lock()
				md5 := DataSignerMd5(i)
				mu.Unlock()

				neck <- DataSignerCrc32(md5)
			}(neck)

			out <- DataSignerCrc32(i) + "~" + <-neck
		}()
	}
	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}

	for i := range in {
		i := i.(string)

		wg.Add(1)
		go func() {
			defer wg.Done()

			thxChs := make([]chan string, thxCount)
			for j := 0; j < thxCount; j++ {
				thxChs[j] = make(chan string, 1)
				ordered := strconv.Itoa(j) + i

				go func(thCh chan string, ordered string) {
					thCh <- DataSignerCrc32(ordered)
				}(thxChs[j], ordered)
			}

			var concated string
			for _, ch := range thxChs {
				concated += <-ch
			}
			out <- concated
		}()
	}
	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	combined := []string{}

	for i := range in {
		i := i.(string)
		combined = append(combined, i)
	}

	sort.Strings(combined)
	out <- strings.Join(combined, "_")
}