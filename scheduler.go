package gotaskscheduler

import (
	"errors"
	"fmt"
	"runtime"
	"time"
)

//Function type definition
type Fn func()

//Timer type definition
type timer struct {
	name     string
	once     bool
	seconds  uint32
	stop     bool
	function Fn
}

//List type definition
type TList struct {
	id      int
	name    string
	seconds uint32
	once    bool
}

//Timers (internal list)
var timers = map[int]*timer{}

//true while scheduler is stopping
var doStop bool

//last task id
var lastTimer int

//max tasks
var tasksLimit int = 50

//Set max tasks, default is 50. Warning: too many tasks can dangerously increase CPU usage
func SetTasksLimit(limit int) {
	tasksLimit = limit
}

//Add new Task to execute every specified seconds (once=true for only one execution).
//Name can be an empty string.
//seconds will be 1 to 100years (in seconds).
//function is a func task
//once determine if it's an one time task
func AddTask(name string, seconds uint32, function Fn, once bool) (id int, err error) {
	if len(timers) >= tasksLimit {
		err = errors.New("too many tasks")
		return 0, err
	}

	//Seconds is uint32, the limit is 4294967295, but 3155760000 seconds is 100 years, should be enough
	if seconds >= 3155760001 {
		err = errors.New("too much time")
		return 0, err
	}

	//this should never happen
	if lastTimer >= 4294967290 {
		err = errors.New("too many tasks, danger of overflow")
		return 0, err
	}

	//increase id
	lastTimer++
	id = lastTimer

	//if no name, set default "Task" + ID
	if name == "" {
		name = fmt.Sprintf("Task %d", id)
	}

	if !once {
		timers[id] = &timer{name, once, seconds, false, function}

		go func(id int) {
			for {
				time.Sleep(time.Duration(seconds) * time.Second)
				if timers[id] == nil {
					break
				}

				// Verify if schedule is being stopped
				if doStop {
					break
				}

				// Verify if task is being stopped
				if timers[id].stop {
					break
				}

				function()
			}
		}(id)

		err = nil
		return id, err
	}

	go func() {
		function()
	}()

	err = nil
	return id, err
}

//Manually task execution
func ExecTask(id int) (err error) {
	if _, ok := timers[id]; ok {
		timers[id].function()
		return nil
	}
	err = errors.New("task does not exist")
	return err
}

//Count total tasks.
func CountTasks() int {
	return len(timers)
}

//List Tasks (ID, NAME, SECONDS)
func ListTasks() (list map[int]*TList) {
	list = make(map[int]*TList)

	for key, value := range timers {
		list[key] = &TList{key, value.name, value.seconds, value.once}
	}

	return list
}

func StopTask(id int) {
	if _, ok := timers[id]; ok {
		timers[id].stop = true
	}
}

//Stop Scheduler (and optionally Delete all Tasks) (if prev started)
func StopAllTasks(DelTasks bool) {
	doStop = true

	if DelTasks {
		timers = map[int]*timer{}
		go func() {
			runtime.GC()
		}()
	}
}

func DeleteTask(id int) {
	delete(timers, id)
	go func() {
		runtime.GC()
	}()
}
