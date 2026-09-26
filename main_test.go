package main

import (
	"testing"
	"time"
)

func waitForCompletion(t *testing.T, done <-chan struct{}) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("station did not finish")
	}
}

func TestTimeoutStopsWorkersAndDispatcher(t *testing.T) {
	timeout := make(chan struct{})
	s := &Station{
		Pumps:    make(chan Pump, 2),
		Vehicles: make(chan Vehicle),
		Timeout:  timeout,
	}
	// With no vehicles, both workers must exit through cancellation.
	wg.Add(2)
	go s.FillNext(Pump{Id: 0, Fillups: 3})
	go s.FillNext(Pump{Id: 1, Fillups: 4})
	close(timeout)
	workersDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(workersDone)
	}()
	waitForCompletion(t, workersDone)
	if len(s.Pumps) != 2 {
		t.Fatalf("got %d pumps back, want 2", len(s.Pumps))
	}
	for i := 0; i < 2; i++ {
		p := <-s.Pumps
		if p.Fillups != p.Id+3 {
			t.Errorf("pump %d fillups = %d, want %d", p.Id, p.Fillups, p.Id+3)
		}
	}
	// Cancellation remains observable after both workers receive it.
	dispatcherDone := make(chan struct{})
	go func() {
		s.RunPumps()
		close(dispatcherDone)
	}()
	waitForCompletion(t, dispatcherDone)
}

func TestFillNextReturnsUpdatedVehicleAndPump(t *testing.T) {
	s := &Station{
		Pumps:    make(chan Pump, 1),
		Vehicles: make(chan Vehicle, 1),
	}
	s.Vehicles <- Vehicle{Id: 7, Fillups: 2}
	wg.Add(1)
	done := make(chan struct{})
	go func() {
		s.FillNext(Pump{Id: 5, Fillups: 4})
		wg.Wait()
		close(done)
	}()
	waitForCompletion(t, done)
	if len(s.Pumps) != 1 || len(s.Vehicles) != 1 {
		t.Fatal("fillup did not return both the pump and vehicle")
	}
	if p := <-s.Pumps; p != (Pump{Id: 5, Fillups: 5}) {
		t.Errorf("unexpected pump: %+v", p)
	}
	if v := <-s.Vehicles; v != (Vehicle{Id: 7, Fillups: 3}) {
		t.Errorf("unexpected vehicle: %+v", v)
	}
}
