# SP1 Jump Game 🎮

Jumping game supported by [SP1](https://github.com/succinctlabs/sp1).  
After completing the challenge, a unique proof will be generated for you.

## Preview

![](https://github.com/walirt/sp1-jump-game/blob/main/game/game.gif?raw=true)

## How to play

Use the left mouse button to control Freakybird to bounce, the longer you press, the farther you jump, when you score reach 10 points, congratulations on your successful challenge.

## Requirements

- [Rust](https://rustup.rs/)
- [SP1](https://docs.succinct.xyz/getting-started/install.html)
- [Go](https://go.dev/dl/)

## Running the Game

```sh
cd game
go run main.go
```

## Note

Generating proofs defaults to mock mode, if you want to generate proofs locally, change `game/index.html`.
```js
const mock = false;
```
It communicates with the proof server using websocket.

## Have fun！
