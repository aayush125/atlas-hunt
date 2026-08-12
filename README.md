# Atlas Hunt

## Gameplay

This is a web-based multi-player game where players join a room to hunt treasures to collect points. Over 3 rounds, the player with the highest score in at least 2 rounds wins the game.

## Architecture

The server is made in Go using the Gorilla websocket package, following an actor-based model where a single Goroutine handles all transactions of an ongoing game.

The following images illustrate the high-level and the real-time architectures of the project.

<br />
<br />

**High-Level Architecture:**

<p align="center">
  <img src="./images/architecture-high-level.png" alt="High-Level Architecture">
</p>

<br />
<br />

**Real-time Architecture:**

<p align="center">
  <img src="./images/architecture-realtime.png" alt="Real-time Architecture">
</p>

## WebSocket Events

All WebSocket message types are defined in [messages.go](./server/internal/wsmessage/messages.go), with direction annotated per event.
