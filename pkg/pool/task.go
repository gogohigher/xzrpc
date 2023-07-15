package pool

type ITask interface {
	ID() string
}

type Task func()
