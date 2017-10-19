package main

import (
	"fmt"
	"os"
	"sync"
	"text/tabwriter"
	"time"
)

type Station struct {
	Pumps    chan Pump
	Vehicles chan Vehicle
	Timeout  <-chan time.Time
}

type Vehicle struct {
	Id      int
	Fillups int
}

type Pump struct {
	Id      int
	Fillups int
}

// Wait group to ensure all cars and pumps have finished executiong
// before attempting to run the final report
var wg sync.WaitGroup

// RunPumps is the main execution loop for Station
// It will either receive an available pump and initiate a fillup on it
// or it will return when the maximum execution time has been reached
func (s *Station) RunPumps() {
	for {
		select {
		case <-s.Timeout:
			return
		case p := <-s.Pumps:
			wg.Add(1)
			go s.FillNext(p)
		}
	}

}

// Fillup fills a car and sleeps for 50ms
func (c *Vehicle) Fillup() {
	time.Sleep(time.Millisecond * 50)
	c.Fillups++
}

// FillNext will receive a car from the vehicles channel, fill it,
// incrememet the number of fillups on the car and pump, and then return
// the car and pump to their respective channels
func (s *Station) FillNext(p Pump) {
	select {
	case <-s.Timeout:
		return
	case current := <-s.Vehicles:
		current.Fillup()
		p.Fillups++
		//fmt.Printf("Filling car %v\n", current.Id)
		s.Vehicles <- current
		s.Pumps <- p
		wg.Done()
		return

	}
}

// PrintTotals nicely formats a report detailing how many fillups a pump gave
// and how many fillups each car received
func (s *Station) PrintTotals() {
	w := tabwriter.NewWriter(os.Stdout, 15, 15, 0, '\t', tabwriter.AlignRight)

	row := "===============\t==============="
	fmt.Fprintln(w, row)
	fmt.Fprintln(w, "Pump\tFills Given")
	fmt.Fprintln(w, row)

	close(s.Pumps)
	for p := range s.Pumps {
		fmt.Fprintf(w, "%v\t%v\n", p.Id, p.Fillups)
	}

	fmt.Fprintln(w, "")

	fmt.Fprintln(w, row)
	fmt.Fprintln(w, "Car\tFills Received")
	fmt.Fprintln(w, row)

	close(s.Vehicles)
	for c := range s.Vehicles {
		fmt.Fprintf(w, "%v\t%v\n", c.Id, c.Fillups)
	}
	w.Flush()
}

// NewStation creates a new instance of a Station struct
func NewStation(maxCars int, maxPumps int) *Station {

	// Buffered channels for up to maxCars and maxPumps
	pumps := make(chan Pump, maxPumps)
	vehicles := make(chan Vehicle, maxCars)

	// Timeout station execution loop after 30 seconds
	timeout := time.After(time.Second * 30)

	// Load up buffered channels
	for i := 0; i < maxPumps; i++ {
		pumps <- Pump{Id: i}
	}

	for i := 0; i < maxCars; i++ {
		vehicles <- Vehicle{Id: i}
	}

	// New up a Station and return
	s := &Station{
		Pumps:    pumps,
		Vehicles: vehicles,
		Timeout:  timeout}

	return s
}

func main() {
	s := NewStation(10, 4)
	s.RunPumps()
	wg.Wait()
	s.PrintTotals()
}
