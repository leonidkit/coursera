package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const TH = 6

func SingleHash(in chan interface{}, out chan interface{}) {
	var data string
	var wg sync.WaitGroup
	var mutex sync.Mutex

	for val := range in {
		switch val.(type) {
		case int:
			data = strconv.Itoa(val.(int))
		case string:
			data = val.(string)
		}

		wg.Add(1)
		go AsyncSingleHash(data, out, &wg, &mutex)
	}
	wg.Wait()
}

func AsyncSingleHash(data string, out chan interface{}, wg *sync.WaitGroup, mutex *sync.Mutex) {
	crc32Ch := make(chan string)
	md5Ch := make(chan string)

	go AsyncCrc32(data, crc32Ch)
	go AsyncMD5(data, md5Ch, mutex)

	md5Hash := <-md5Ch

	go AsyncCrc32(md5Hash, md5Ch)

	crc32Hash := <-crc32Ch
	crc32md5Hash := <-md5Ch

	out <- crc32Hash + "~" + crc32md5Hash

	wg.Done()
}

func AsyncCrc32(data string, res chan<- string) {
	res <- DataSignerCrc32(data)
}

func AsyncMD5(data string, res chan<- string, mutex *sync.Mutex) {
	mutex.Lock()
	res <- DataSignerMd5(data)
	mutex.Unlock()
}

func MultiHash(in chan interface{}, out chan interface{}) {
	var wg sync.WaitGroup

	for shash := range in {
		wg.Add(1)
		go AsyncMultiHash(shash.(string), out, &wg)
	}
	wg.Wait()
}

func AsyncMultiHash(data string, out chan interface{}, wg *sync.WaitGroup) {
	var wgl sync.WaitGroup
	multiHash := make([]string, TH)

	for i := 0; i < TH; i++ {
		wgl.Add(1)
		go func(data string, th int) {
			multiHash[th] = DataSignerCrc32(strconv.Itoa(th) + data)
			wgl.Done()
		}(data, i)
	}
	wgl.Wait()

	out <- strings.Join(multiHash, "")

	wg.Done()
}

func CombineResults(in chan interface{}, out chan interface{}) {
	results := []string{}

	for hash := range in {
		results = append(results, hash.(string))
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i] < results[j]
	})

	out <- strings.Join(results, "_")
}

func ExecutePipeline(jobs ...job) {
	var inCh = make(chan interface{})
	var outCh = make(chan interface{})
	var wg sync.WaitGroup

	for _, j := range jobs {
		wg.Add(1)

		go func(task job, in, out chan interface{}) {
			task(in, out)
			close(out)
			wg.Done()
		}(j, inCh, outCh)

		inCh = outCh
		outCh = make(chan interface{})
	}
	wg.Wait()
}

func main() {
	inputData := []int{0, 1, 2, 3, 4, 5, 6}

	hashSignJobs := []job{
		job(func(in, out chan interface{}) {
			for _, fibNum := range inputData {
				out <- fibNum
			}
		}),
		job(SingleHash),
		job(MultiHash),
		job(CombineResults),
		job(func(in, out chan interface{}) {
			dataRaw := <-in
			data, ok := dataRaw.(string)
			if !ok {
				fmt.Println("cant convert result data to string")
			}
			fmt.Println(data)
		}),
	}

	ExecutePipeline(hashSignJobs...)
}

// 29568666068035183841425683795340791879727309630931025356555_4958044192186797981418233587017209679042592862002427381542
// 29568666068035183841425683795340791879727309630931025356555_4958044192186797981418233587017209679042592862002427381542
//
// Вопросы:
// 1) Детектор гонок не ругается на res[th]
// 2) Mutex обязан определять на самом высоком уровне, почему
