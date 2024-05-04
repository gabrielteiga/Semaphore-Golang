package main

import (
	"fmt"
	"sync"
	"time"
)

type Semaphore struct { // este semáforo implementa quaquer numero de creditos em "v"
	v    int           // valor do semaforo: negativo significa proc bloqueado
	fila chan struct{} // canal para bloquear os processos se v < 0
	sc   chan struct{} // canal para atomicidade das operacoes wait e signal
}

func NewSemaphore(init int) *Semaphore {
	s := &Semaphore{
		v:    init,                   // valor inicial de creditos
		fila: make(chan struct{}),    // canal sincrono para bloquear processos
		sc:   make(chan struct{}, 1), // usaremos este como semaforo para SC, somente 0 ou 1
	}
	return s
}

func (s *Semaphore) Wait() {
	s.sc <- struct{}{} // SC do semaforo feita com canal
	s.v--              // decrementa valor
	if s.v < 0 {       // se negativo era 0 ou menor, tem que bloquear
		<-s.sc               // antes de bloq, libera acesso
		s.fila <- struct{}{} // bloqueia proc
	} else {
		<-s.sc // libera acesso
	}
}

func (s *Semaphore) Signal() {
	s.sc <- struct{}{} // entra sc
	s.v++
	if s.v <= 0 { // tem processo bloqueado ?
		<-s.fila // desbloqueia
	}
	<-s.sc // libera SC para outra op
}

const N = 1000
const waitingChairs = 10

var customersLeft = 0
var customersDone = 0
var customers = 0
var mutex = sync.Mutex{}
var customerSemaphore = NewSemaphore(0)
var barberSemaphore = NewSemaphore(0)

func customer(id int) {
	mutex.Lock()
	if customers == waitingChairs+1 {
		customersLeft++
		mutex.Unlock()
		return
	}
	customers++
	mutex.Unlock()

	customerSemaphore.Signal()
	barberSemaphore.Wait()
	getHaircut(id)

	mutex.Lock()
	customers--
	customersDone++
	mutex.Unlock()
}

func getHaircut(id int) {
	fmt.Printf("Customer %d is getting a haircut\n", id)
}

func barber() {
	for {
		customerSemaphore.Wait()
		barberSemaphore.Signal()
		fmt.Println("Barber is cutting hair")
	}
}

func main() {
	wg := sync.WaitGroup{}

	go barber()
	for i := 0; i < N; i++ {
		time.Sleep(1 * time.Nanosecond)
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			customer(id)
		}(i)
	}

	wg.Wait()
	fmt.Printf("\n\n%d customers left because the room was full", customersLeft)
	fmt.Printf("\n%d customers got a haircut", customersDone)
}
