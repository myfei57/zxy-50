package dispatch

import (
	"encoding/json"
	"sync"

	"drainnet/internal/store"
)

const QueueKind = "commands"

type CommandQueue struct {
	store *store.Store
	mu    sync.Mutex
}

func NewCommandQueue(st *store.Store) *CommandQueue {
	return &CommandQueue{store: st}
}

func (q *CommandQueue) Enqueue(command Command) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.store.AppendJSON(QueueKind, command)
}

func (q *CommandQueue) Pending() ([]Command, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	rows, err := q.store.ReadJournal(QueueKind)
	if err != nil {
		return nil, err
	}
	commands := make([]Command, 0, len(rows))
	for _, row := range rows {
		var command Command
		if err := json.Unmarshal(row, &command); err != nil {
			return nil, err
		}
		commands = append(commands, command)
	}
	return commands, nil
}

func (q *CommandQueue) Clear() error {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.store.ReplaceJournal(QueueKind, nil)
}

func (d *Dispatcher) EnqueueCommand(command Command) error {
	if command.ID == "" {
		command.ID = newID()
	}
	if command.At.IsZero() {
		command.At = now()
	}
	return d.queue.Enqueue(command)
}

func (d *Dispatcher) PendingCommands() ([]Command, error) {
	return d.queue.Pending()
}
