package mcp

import (
	"crypto/rand"
	"time"
)

type TaskId string
type TaskStatus string

var store *TaskStore = nil

const (
	WORKING   TaskStatus = "working"
	INPUT_REQ TaskStatus = "input_required"
	COMPLETED TaskStatus = "completed"
	FAILED    TaskStatus = "failed"
	CANCELLED TaskStatus = "cancelled"
)

type TaskStatusUpdate struct {
	TaskId        TaskId     `json:"taskId"`
	Status        TaskStatus `json:"status"`
	CreatedAt     time.Time  `json:"createdAt"`
	LastUpdatedAt time.Time  `json:"lastUpdatedAt"`
	Ttl           int        `json:"ttl,omitempty"`
	PollInterval  int        `json:"pollInterval,omitempty"`
}

type Task struct {
	Id     TaskId     `json:"taskId"`
	Status TaskStatus `json:"status"`

	result []any
	err    error

	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"lastUpdatedAt"`
	Ttl          int       `json:"ttl,omitempty"`
	PollInterval int       `json:"pollInterval,omitempty"`
}

func NewTask(s *TaskStore, f func() ([]any, error)) *Task {
	task := Task{
		Id:     TaskId(rand.Text()),
		Status: WORKING,

		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Ttl:          15 * 60 * 1000,
		PollInterval: 5 * 1000,
	}

	go task.Run(s, f)

	return &task
}

func (t *Task) Run(s *TaskStore, f func() ([]any, error)) {
	resp, err := f()

	update := Task{
		Id:        t.Id,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}

	if err != nil {
		update.Status = FAILED
		update.err = err
	} else {
		update.Status = COMPLETED
		update.result = resp
	}

	s.Update(&update)
}

// TODO: add a cleanup goroutine to clean up tasks older than ttl
type TaskStore struct {
	server  *Server
	channel chan *Task

	tasks map[TaskId]*Task
}

/*
Singleton task store:
  - Any function can call the thread safe Add and Update operations to start and update functions
    Concurrency is handled within the store itself
  - Completed tasks will send out notifications back to the attached client
*/
func NewTaskStore(s *Server) *TaskStore {
	if store != nil || s == nil {
		return store
	}

	channel := make(chan *Task)

	store = &TaskStore{
		tasks:   map[TaskId]*Task{},
		channel: channel,
		server:  s,
	}

	go store.Run()

	return store
}

func (t *TaskStore) Run() {
	for task := range t.channel {
		t.tasks[task.Id] = task

		t.server.sender.SendNotification("notifications/tasks/update", TaskStatusUpdate{
			TaskId:        task.Id,
			Status:        task.Status,
			CreatedAt:     task.CreatedAt,
			LastUpdatedAt: task.UpdatedAt,
			Ttl:           15 * 60 * 1000,
			PollInterval:  5 * 1000,
		})
	}
}

func (t *TaskStore) Close() {
	close(t.channel)
}

func (t *TaskStore) Get(id TaskId) *Task {
	return t.tasks[id]
}

func (t *TaskStore) Add(f func() ([]any, error)) *Task {
	task := NewTask(t, f)
	t.Update(task)

	return task
}

func (t *TaskStore) Update(task *Task) {
	t.channel <- task
}
