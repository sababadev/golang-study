package main

import (
	"fmt"
	"log"
	"strings"
)

type Room struct {
	Name        string
	Description string
	Greeting    string
	Paths       []string
	Lock        map[string]bool
	Lootboxs    []*Contaiter
	Triggers
}

func (r *Room) Enable() string {
	return fmt.Sprintf("можно пройти - %s", strings.Join(r.Paths, ", "))
}

func (r *Room) GetItems() string {
	content := []string{}
	for _, lootbox := range r.Lootboxs {
		if len(lootbox.Content) != 0 {
			content = append(content, lootbox.String())
		}
	}
	res := strings.Join(content, ", ")

	if len(res) != 0 {
		return res
	}
	return "пустая комната"
}

func (r *Room) Unlock(item string) string {
	if _, ok := r.Lock[item]; ok {
		r.Lock[item] = false
		return fmt.Sprint("дверь открыта")
	}
	return fmt.Sprintln("не к чему применить")
}

type World struct {
	Ways map[string]*Room
}

func NewWorld() *World {
	return &World{
		Ways: map[string]*Room{},
	}
}

func (w *World) GetRoom(name string) *Room {
	if room, ok := w.Ways[name]; ok {
		return room
	}
	return nil
}

func (w *World) AddRoom(name, desc, welc string) *Room {
	if _, ok := w.Ways[name]; !ok {
		room := &Room{Name: name,
			Description: desc,
			Greeting:    welc,
			Paths:       []string{},
			Lock:        map[string]bool{},
			Lootboxs:    []*Contaiter{},
		}
		w.Ways[name] = room
		return room
	}
	log.Printf("name %v - already exist; name must be unique!", name)
	return nil
}

func (w *World) AddPath(from, to string, lock bool) {
	entry, exit := w.Ways[from], w.Ways[to]

	switch {
	case entry == nil:
		log.Printf("<from> room: %v not exist. path not added!", from)
		return
	case exit == nil:
		log.Printf("<to> room: %v not exist. path not added!", to)
		return
	}

	for _, val := range entry.Paths {
		if val == exit.Name {
			log.Printf("path: %v <-> %v : exist. doing nothng", from, to)
			return
		}
	}

	entry.Paths = append(entry.Paths, exit.Name)

	if entry.Name != exit.Name {
		exit.Paths = append(exit.Paths, entry.Name)
	}

	w.Ways[entry.Name], w.Ways[exit.Name] = entry, exit

	entry.Lock[exit.Name] = lock
	exit.Lock[entry.Name] = lock
}

func (w *World) SpawnItems(room string, where string, items []string) {
	space := w.GetRoom(room)
	lootbox := &Contaiter{Name: where}
	lootbox.Content = items
	space.Lootboxs = append(space.Lootboxs, lootbox)
}
