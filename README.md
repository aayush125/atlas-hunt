# Atlas Hunt

## Gameplay

This is a web-based multi-player game where players join a room to hunt treasures to collect points. Over 3 rounds, the player with the highest score in at least 2 rounds wins the game.

## Architecture

The server was made in Go using the Gorilla websocket package, following an actor-based model where a single Goroutine handles all transactions of an ongoing game.
