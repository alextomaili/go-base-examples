package chanel_send

import (
	"sync"
	"testing"
)

var result, load int64

func SendChanel(b *testing.B, c chan int64) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	//читаем пока канал не закрыт
	go func() {
		for {
			select {
			case i, ok := <-c:
				if !ok {
					wg.Done()
					return
				}
				result = result + i
			}
		}
	}()
	//пишем в тестируемый канал
	for i := 0; i < b.N; i++ {
		//добавим работы чтобы читатель успел прочитать
		for j := 0; j < 10; j++ {
			load = load + int64(i*j)
		}
		c <- load //пишем в канал
	}
	close(c)
	wg.Wait()
}