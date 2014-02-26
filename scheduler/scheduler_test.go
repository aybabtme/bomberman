package scheduler_test

import (
	"fmt"
	"github.com/aybabtme/bomberman/scheduler"
)

type PrintAction string

func (p PrintAction) Duration() int { return 1 }

func ExampleScheduler_simple() {
	s := scheduler.NewScheduler()

	actions := []struct {
		action PrintAction
		time   int
	}{
		{PrintAction("Still there?"), 5},
		{PrintAction("Hello"), 1},
		{PrintAction("Bye"), 3},
		{PrintAction("Bonjour"), 2},
		{PrintAction("Bye"), 3},
	}

	for _, action := range actions {
		s.Register(action.action, action.time)
	}

	for s.HasNext() {
		s.NextTurn()
		s.DoTurn(func(a scheduler.Action, turn int) error {
			fmt.Println(a.(PrintAction))
			return nil
		})
	}

	// Output:
	// Hello
	// Bonjour
	// Bye
	// Bye
	// Still there?
}

// func ExampleScheduler_serialized() {
// 	s := scheduler.NewScheduler()

// 	actions := []struct {
// 		action PrintAction
// 		time   int
// 	}{
// 		{PrintAction("Still there?"), 5},
// 		{PrintAction("Hello"), 1},
// 		{PrintAction("Bye"), 3},
// 		{PrintAction("Bonjour"), 2},
// 		{PrintAction("Bye"), 3},
// 	}

// 	for _, action := range actions {
// 		s.Register(action.action, action.time)
// 	}

// 	buf := bytes.NewBuffer(nil)
// 	err := s.Encode(buf)
// 	if err != nil {
// 		panic(err)
// 	}

// 	log.Printf("Scheduler=%s", buf.String())

// 	back, err := scheduler.Decode(buf)
// 	if err != nil {
// 		panic(err)
// 	}

// 	for back.HasNext() {
// 		back.NextTurn()
// 		back.DoTurn(func(a scheduler.Action, turn int) error {
// 			fmt.Println(a.(PrintAction))
// 			return nil
// 		})
// 	}

// 	// Output:
// 	// Hello
// 	// Bonjour
// 	// Bye
// 	// Bye
// 	// Still there?
// }
