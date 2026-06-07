package main

import (
	"net"

	rl "github.com/gen2brain/raylib-go/raylib"
	"raylib-game/game"
	"raylib-game/network"
)

func main() {
	rl.InitWindow(800, 600, "P2P Game - Go Client")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)

	udpConn, _ := net.ListenUDP("udp", &net.UDPAddr{Port: 0})
	defer udpConn.Close()

	room := network.NewRoom("sala1", udpConn)
	go room.Connect()
	go network.ReceiverLoop(udpConn)

	game.New(game.Player{
		X:     0,
		Y:     0,
		W:     50,
		H:     50,
		Name:  "Voce",
		Color: rl.Red,
		Direction: game.Stop,
		Input: func(p *game.Player) {
			if rl.IsKeyDown(rl.KeyLeft) {
				p.Direction = game.Left
			}
			if rl.IsKeyDown(rl.KeyRight) {
				p.Direction = game.Right
			}
			if rl.IsKeyDown(rl.KeyUp) {
				p.Direction = game.Up
			}
			if rl.IsKeyDown(rl.KeyDown) {
				p.Direction = game.Down
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