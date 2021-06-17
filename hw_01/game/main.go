package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Action func(*Game, ...string) string
type Triggers map[string]func(*Game)

type Game struct {
	*World
	*Player
}

var (
	quest    *Game
	commands map[string]Action
	triggers Triggers
)

func main() {
	initGame()
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		fmt.Println(handleCommand(s.Text()))
	}
}

func handleCommand(command string) string {
	args := strings.Fields(command)
	if cmd, ok := commands[args[0]]; ok {
		msg := cmd(quest, args[1:]...)
		return msg
	}
	return "неизвестная команда"
}

func initGame() {
	world := NewWorld()

	world.AddRoom("комната", "", "ты в своей комнате")
	world.AddRoom("кухня", "ты находишься на кухне", "кухня, ничего интересного")
	world.AddRoom("коридор", "", "ничего интересного")
	world.AddRoom("улица", "", "на улице весна")

	world.SpawnItems("комната", "на столе", []string{"ключи", "конспекты"})
	world.SpawnItems("комната", "на стуле", []string{"рюкзак"})
	world.SpawnItems("кухня", "на столе", []string{"чай"})

	world.AddPath("коридор", "кухня", false)
	world.AddPath("коридор", "комната", false)
	world.AddPath("коридор", "улица", true)

	startRoom := world.GetRoom("кухня")
	mission := []string{"собрать рюкзак", "идти в универ"}
	player := NewPlayer(startRoom, mission)

	quest = &Game{World: world, Player: player}
	commands = InitActions()

	triggers = map[string]func(*Game){
		"взять конспекты": func(g *Game) {
			g.CompliteMission("собрать рюкзак")
		},
		"применить ключи дверь": func(g *Game) {
			for i, path := range g.Stying.Paths {
				if path == "коридор" {
					g.Stying.Paths[i] = "домой"
				}
			}
		},
	}
}

func InitActions() map[string]Action {
	Move := func(g *Game, args ...string) string {
		to := args[0]
		close, ok := g.Stying.Lock[to]
		if !ok {
			return fmt.Sprintf("нет пути в %s", to)
		}

		if !close {
			next := g.Ways[to]
			g.Stying = next
			if g.Stying.Name == "улица" {
				if event, ok := triggers["применить ключи дверь"]; ok {
					event(g)
				}
			}
			return fmt.Sprintf("%s. %s", next.Greeting, next.Enable())
		}
		return "дверь закрыта"
	}

	Around := func(g *Game, args ...string) string {
		description := g.Stying.Description
		items := g.Stying.GetItems()

		var tasks string
		if g.Stying.Name == "кухня" {
			tasks = g.PrintMission()
		} else {
			tasks = ""
		}

		output := []string{}
		for _, info := range []string{description, items, tasks} {
			if info != "" {
				output = append(output, info)
			}
		}

		return fmt.Sprintf("%s. %s", strings.Join(output, ", "), g.Stying.Enable())
	}

	Wear := func(g *Game, args ...string) string {
		name := args[0]
		for _, box := range g.Stying.Lootboxs {
			if ok := box.FindItem(name); ok {
				box.DeleteItem(name)
				g.Inventory = &Contaiter{}
				return fmt.Sprintf("вы надели: %s", name)
			}
		}
		return "нет такого"
	}

	Take := func(g *Game, args ...string) string {
		name := args[0]
		if g.Inventory != nil {
			for _, box := range g.Stying.Lootboxs {
				if ok := box.FindItem(name); ok {
					box.DeleteItem(name)
					g.Inventory.AddItem(name)

					if event, ok := triggers["взять "+name]; ok {
						event(g)
					}
					return fmt.Sprintf("предмет добавлен в инвентарь: %s", name)
				}
			}
			return "нет такого"
		}
		return "некуда класть"
	}

	Apply := func(g *Game, args ...string) string {
		sub, _ := args[0], args[1]
		if ok := g.Inventory.FindItem(sub); !ok {
			return fmt.Sprintf("нет предмета в инвентаре - %s", sub)
		}

		for room, lock := range g.Stying.Lock {
			if lock {
				return g.Stying.Unlock(room)
			}
		}
		return "не к чему применить"
	}

	return map[string]Action{
		"идти":        Move,
		"надеть":      Wear,
		"взять":       Take,
		"применить":   Apply,
		"осмотреться": Around,
	}
}
