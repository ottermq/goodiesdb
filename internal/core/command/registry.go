package command

import (
	"strings"
	"sync"
)

type Registry struct {
	commands map[string]Command
	mu       sync.RWMutex
}

func NewRegistry() *Registry {
	r := &Registry{
		commands: make(map[string]Command),
	}

	r.Register(NewGetCommand())
	r.Register(NewSetCommand())
	r.Register(NewDelCommand())

	return r
}

// Register adds a command to the registry
func (r *Registry) Register(cmd Command) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.commands[cmd.Name()] = cmd
}

// Get retrieves a command by name
func (r *Registry) Get(name string) (Command, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cmd, ok := r.commands[strings.ToUpper(name)]
	return cmd, ok
}
