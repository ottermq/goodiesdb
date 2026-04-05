package store

type AOFCommand []string

func NewAOFCommand(name string, args ...string) AOFCommand {
	cmd := make(AOFCommand, 0, len(args)+1)
	cmd = append(cmd, name)
	cmd = append(cmd, args...)
	return cmd
}
