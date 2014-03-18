# Bomberman!

![thiiiiis](https://f.cloud.github.com/assets/1189716/2439669/2770d5a0-adff-11e3-9c53-af0e3a59171c.png)

## Right now

* Hard-coded players.
* Can use many types of players.
  * AI player.
  * Keyboard player.
  * [TCP player](https://github.com/aybabtme/bombertcp/blob/master/player.go)
  * Websocket player.
* Sample clients:
  * Sample [websocket](https://github.com/aybabtme/bomberweb) based UI/client.
* Real clients:
  * Python client: https://github.com/uiri/bombermanpy.
  * Java client: https://github.com/aybabtme/bombjava.
  * ... make your own client!

## Making your own client.

You have two choices to implement a client for the language of your choice. Both are usable at this time, however 
the TCP interface should be prefered, simply because I don't forecast major changes in it's working.

* Use the TCP interface, docs [here](https://github.com/aybabtme/bombertcp). 
  [`bombermanpy`](https://github.com/uiri/bombermanpy) uses this.
* Use the websocket interface.

## Implementing native players.

### Go

You can implement a native Go player if you respect the `player.Player` interface:

```go
type Player interface {
	Name() string
	Move() <-chan Move
	Update() chan<- State
}
```

[Details of `Move`, `State` and `Player`](https://github.com/aybabtme/bomberman/blob/master/player/player.go).

### Lua

If there's enough demand for it, I might be able to embed a Lua VM in the bomberman server and make it possible to run native Lua players.
