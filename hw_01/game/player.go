package main

import (
	"fmt"
	"strings"
)

type Player struct {
	Stying    *Room
	Inventory *Contaiter
	Missions  []string
}

func NewPlayer(start *Room, tasks []string) *Player {
	return &Player{
		Stying:   start,
		Missions: tasks,
	}
}

func (p *Player) PrintMission() string {
	var missions = []string{}
	missions = append(missions, p.Missions...)
	return fmt.Sprintf("надо %s", strings.Join(missions, " и "))
}

func (p *Player) CompliteMission(complited string) {
	for idx, mission := range p.Missions {
		if mission == complited {
			copy(p.Missions[idx:], p.Missions[idx+1:])
			p.Missions[len(p.Missions)-1] = ""
			p.Missions = p.Missions[:len(p.Missions)-1]
		}
	}
}
