package cmpId

type StateId int

const (
    State = iota
    EntityList
    Position
)

type ActionId int

const (
    Action = iota
    AddEntity
    Move
)

type EntityId int

const (
    Entity = iota
    Game
    Player
)
