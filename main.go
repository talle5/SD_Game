package main

import (
	"fmt"
	"net"
	"raylib-game/game"
	"raylib-game/network"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	rl.InitWindow(800, 600, "P2P Game - Go Client")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)

	udpConn, _ := net.ListenUDP("udp", &net.UDPAddr{Port: 0})
	defer udpConn.Close()

	room := network.NewRoom("sala1", udpConn, func(id string, p *net.UDPAddr) {
		network.Players[id] = game.New(game.Player{ // <--- Usar o id aqui
			W: 50, H: 50, Color: rl.Blue,
		})
	},
		func(p *net.UDPAddr) {})
	go network.ReceiverLoop(udpConn)
	go room.Connect()

	game.New(game.Player{
		X:         0,
		Y:         0,
		W:         50,
		H:         50,
		Name:      "Voce",
		Color:     rl.Red,
		Direction: game.Stop,
		Input: func(p *game.Player) {
			if rl.IsKeyDown(rl.KeyLeft) {
				p.Direction = game.Left
				go network.Broadcast([]byte(fmt.Sprintf("move %s %d %d %d", network.GetClientID(), p.X, p.Y, p.Direction)), udpConn)
			}
			if rl.IsKeyDown(rl.KeyRight) {
				p.Direction = game.Right
				go network.Broadcast([]byte(fmt.Sprintf("move %s %d %d %d", network.GetClientID(), p.X, p.Y, p.Direction)), udpConn)
			}
			if rl.IsKeyDown(rl.KeyUp) {
				p.Direction = game.Up
				go network.Broadcast([]byte(fmt.Sprintf("move %s %d %d %d", network.GetClientID(), p.X, p.Y, p.Direction)), udpConn)
			}
			if rl.IsKeyDown(rl.KeyDown) {
				p.Direction = game.Down
				go network.Broadcast([]byte(fmt.Sprintf("move %s %d %d %d", network.GetClientID(), p.X, p.Y, p.Direction)), udpConn)
			}
		},
	})

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.DarkGray)
		game.RunAll()
		rl.EndDrawing()
	}
}
