package main

import (
	"log"
	"os"

	"github.com/jingen11/gmail-crawler/internal/command"
)

type commands struct {
	Commands map[string]func(*command.Command) error
}

func newCommands() commands {
	c := commands{
		Commands: map[string]func(*command.Command) error{},
	}
	return c
}

func (c *commands) register(name string, f func(*command.Command) error) {
	c.Commands[name] = f
}

func (c *commands) run(name string, command *command.Command) error {
	err := c.Commands[name](command)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	c := newCommands()

	c.register("auth", command.HandleAuth)
	c.register("scrap", command.HandleScrap)

	comm := os.Args[1]
	args := os.Args[2:]
	err := c.run(comm, &command.Command{
		Name:      comm,
		Arguments: args,
	})
	if err != nil {
		log.Fatalf("error running command: %v", err)
		os.Exit(1)
	}
}
