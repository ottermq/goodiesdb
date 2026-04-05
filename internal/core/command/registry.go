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
	r.Register(NewExistsCommand())
	r.Register(NewExpireCommand())
	r.Register(NewIncrCommand())
	r.Register(NewDecrCommand())
	r.Register(NewTTLCommand())
	r.Register(NewSetNXCommand())
	r.Register(NewTypeCommand())
	r.Register(NewStrLenCommand())
	r.Register(NewGetRangeCommand())
	r.Register(NewLPushCommand())
	r.Register(NewRPushCommand())
	r.Register(NewLRangeCommand())
	r.Register(NewLTrimCommand())
	r.Register(NewLPopCommand())
	r.Register(NewRPopCommand())
	r.Register(NewHSetCommand())
	r.Register(NewHGetCommand())
	r.Register(NewHGetAllCommand())
	r.Register(NewHDelCommand())
	r.Register(NewHExistsCommand())
	r.Register(NewHLenCommand())
	r.Register(NewHMGetCommand())
	r.Register(NewHKeysCommand())
	r.Register(NewHValsCommand())
	r.Register(NewAuthCommand())
	r.Register(NewSelectCommand())
	r.Register(NewRenameCommand())
	r.Register(NewKeysCommand())
	r.Register(NewInfoCommand())
	r.Register(NewPingCommand())
	r.Register(NewEchoCommand())
	r.Register(NewQuitCommand())
	r.Register(NewFlushDBCommand())
	r.Register(NewFlushAllCommand())
	r.Register(NewScanCommand())

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
