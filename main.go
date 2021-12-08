package main

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"
)

func init() {
	// Definir a quantidade de CPU que pode usar

	// Definir apenas 1 CPU
	// runtime.GOMAXPROCS(1)

	// Irá forçar a execução na quantidade máxima de CPU's, mas a partir da versão 1.5 do go não é mais necessário essa definição
	// runtime.GOMAXPROCS(runtime.NumCPU())
}

var result int

var mtx sync.Mutex

func main() {
	// Processo simultâneos
	simultaneousProcess()

	// -----------------------------------------

	// Processos concorrentes - wait-groups
	// concurrentProcess()

	// -----------------------------------------

	// Prevenção de race condition com Mutex
	// useMutexToPreventRaceCondition()

	// -----------------------------------------

	// Channel
	// useChannels()

	// -----------------------------------------

	// Channel with waitgroup
	// useChannelsWithWaitGroup()

	// -----------------------------------------

	// Semáforo
	// semaphore()

	// -----------------------------------------

	// Pipeline pattern
	// usingPipelinePattern()

	// -----------------------------------------

	// Fun IN
	// useFunIN()

	// -----------------------------------------

	// Fun OUT
	// useFunOUT()

}

func useFunOUT() {

	c := generateINTs(4, 10)

	d1 := divide(c)
	d2 := divide(c)

	fmt.Println(<-d1)

	fmt.Println(<-d2)

}

func generateINTs(numbers ...int) chan int {
	channel := make(chan int)

	go func() {
		for _, n := range numbers {

			channel <- n
		}
		close(channel)
	}()

	return channel
}

func useFunIN() {

	X := funnel(generateMSG("hello go"), generateMSG("hello world"))

	for i := 0; i < 10; i++ {
		fmt.Println(<-X)
	}

	fmt.Println("Finished...")
}

func funnel(channel1, channel2 <-chan string) <-chan string {

	channel := make(chan string)

	go func() {

		for {
			channel <- <-channel1
		}
	}()

	go func() {
		for {
			channel <- <-channel2
		}
	}()

	return channel
}

func generateMSG(s string) <-chan string {
	channel := make(chan string)

	go func() {
		for i := 0; ; i++ {

			channel <- fmt.Sprintf("String: %s - Value: %d", s, i)

			time.Sleep(time.Duration(rand.Intn(255)) * time.Millisecond)
		}
	}()

	return channel
}

func usingPipelinePattern() {

	numbers := generate(2, 4, 6)

	result := divide(numbers)

	fmt.Println(<-result)

	fmt.Println(<-result)

	fmt.Println(<-result)
}

func divide(input chan int) chan int {

	channel := make(chan int)

	go func() {
		for number := range input {
			channel <- number / 2
		}

		close(channel)
	}()

	return channel
}

func generate(numbers ...int) chan int {

	channel := make(chan int)

	go func() {

		for _, number := range numbers {

			channel <- number

		}

	}()

	return channel

}

func semaphore() {

	msg := make(chan string)
	ok := make(chan bool)

	go runProcess("P1", 20, nil, &msg, &ok)

	go runProcess("P2", 40, nil, &msg, &ok)

	go func() {
		<-ok
		<-ok
		close(msg)
	}()

	for text := range msg {
		fmt.Println(text)
	}
}

func useChannelsWithWaitGroup() {

	msgs := make(chan string)

	wtg := sync.WaitGroup{}

	wtg.Add(9)

	go runProcess("P1", 20, &wtg, &msgs, nil)

	go runProcess("P2", 30, &wtg, &msgs, nil)

	go runProcess("P3", 40, &wtg, &msgs, nil)

	go runProcess("P4", 50, &wtg, &msgs, nil)

	go runProcess("P5", 60, &wtg, &msgs, nil)

	go runProcess("P6", 70, &wtg, &msgs, nil)

	go runProcess("P7", 80, &wtg, &msgs, nil)

	go runProcess("P8", 90, &wtg, &msgs, nil)

	go runProcess("P9", 100, &wtg, &msgs, nil)

	go func(wait *sync.WaitGroup, chn *chan string) {

		if wait != nil {
			wait.Wait()
		}

		if chn != nil {
			close(*chn)
		}

	}(&wtg, &msgs)

	for text := range msgs {

		fmt.Println(text)

	}

	fmt.Println("Final result: ", result)

}

func useChannels() {

	// make consegue, trabalhar apenas com 3 tipos: maps, slices e chan(channel)

	msg := make(chan string)
	finish := make(chan bool)

	go runProcess("P1", 20, nil, &msg, &finish)

	go func() {
		<-finish
		close(msg)
	}()

	for {

		text := <-msg

		fmt.Println(text)

		if strings.EqualFold(text, "") {
			break
		}

	}

}

func useMutexToPreventRaceCondition() {

	// Verificar race condition: go run -race main.go

	go runProcess("P1", 20, nil, nil, nil)
	go runProcess("P2", 20, nil, nil, nil)

	var s string

	// Obter os retorno da Goroutines
	fmt.Scanln(&s)

	fmt.Println("Final result: ", result)
}

func concurrentProcess() {
	var waitGroup sync.WaitGroup

	waitGroup.Add(2)

	go runProcess("P1", 20, &waitGroup, nil, nil)
	go runProcess("P2", 20, &waitGroup, nil, nil)

	waitGroup.Wait()
}

func simultaneousProcess() {
	go runProcess("P1", 20, nil, nil, nil)
	go runProcess("P2", 20, nil, nil, nil)

	var s string

	// Segurar o processo para que as rotinas acima executem, se pressionar qualquer tecla irá encerrar o programa
	fmt.Scanln(&s)
}

func runProcess(name string, total int, waitGroup *sync.WaitGroup, chn *chan string, hasSemaphore *chan bool) {

	if waitGroup != nil {
		defer waitGroup.Done()
	}

	for i := 1; i <= total; i++ {

		t := time.Duration(rand.Intn(255))

		time.Sleep(time.Millisecond * t)

		mtx.Lock() // Prevenir race condition - bloquear alteração fora de fluxo

		result++

		text := fmt.Sprintf("Processo: %10s -> %04d -- %30s -- Partial result: %04d",
			name,
			i,
			time.Now().Format("2006-01-02 15:04:05.999999999"),
			result)

		mtx.Unlock() // Prevenir race condition - desbloquear alteração para que o fluxo flua normalmente

		if chn == nil {
			fmt.Println(text)
		} else {
			*chn <- text
		}

	}

	if hasSemaphore != nil {
		*hasSemaphore <- true
	}

}
