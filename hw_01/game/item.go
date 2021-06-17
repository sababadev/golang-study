package main

import (
	"strings"
)

type Contaiter struct {
	Name    string
	Content []string
}

func (c *Contaiter) AddItem(name string) {
	c.Content = append(c.Content, name)
}

func (c *Contaiter) FindItem(name string) bool {
	if c != nil {
		for _, item := range c.Content {
			if item == name {
				return true
			}
		}
		return false
	}
	return false
}

func (c *Contaiter) String() string {
	items := strings.Join(c.Content, ", ")
	if len(items) != 0 {
		return c.Name + ": " + items
	}
	return ""
}

func (c *Contaiter) DeleteItem(name string) {
	for idx, item := range c.Content {
		if item == name {
			copy(c.Content[idx:], c.Content[idx+1:])
			c.Content[len(c.Content)-1] = ""
			c.Content = c.Content[:len(c.Content)-1]
		}
	}
}
