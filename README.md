## A hub is initialized by main.go

Hub contains the different channels for events (subscribe, unsubscribe, broadcast)

It also contains the map of clients (starts as empty; more on this later)

## A go routine with an infinite loop is called

The go routine listens to (hub's) different channels

## Hub instance is passed to function handler servews

## servews creates an instance of client

Client contains values related to Client, user or the connection

The Client also contains an instance of the Hub (really more for accessing the channels while you manage the ws connection)

## servews launches 2 go routines, 1 for reading and writing

These allow you to read and write from/to the clients then broadcast to the approriate channel
