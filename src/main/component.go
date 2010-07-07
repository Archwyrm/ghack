// Contains everything necessary for component system

package main

type State interface {
    Id() int
    Name() string
}

// Transactional state change
type Action interface {
    Id() int
    Name() string
    Act(entStates map[string]State)
}

type Entity interface {
    Id() int
    Name() string
}

// Contains all the data that each component needs
type CmpData struct {
    // Use maps for easy/add remove for now
    states  map[string]State
    actions map[string]Action
    input   chan CmpMsg
}

func NewCmpData() *CmpData {
    states := make(map[string]State)
    actions := make(map[string]Action)
    return &CmpData{states, actions, nil}
}

func (cd CmpData) getState(state State) State {
    ret := cd.states[state.Name()]
    return ret
}

func (cd CmpData) setState(state State) {
    cd.states[state.Name()] = state
}

func (cd CmpData) addAction(action Action) {
    cd.actions[action.Name()] = action
}

// Main loop which handles all component tasks
func (cd CmpData) Run(input chan CmpMsg) {
    cd.input = input

    for {
        m := <-input
        switch {
        case m.Id == MsgTick:
            cd.update()
        case m.Id == MsgGetState:
            cd.sendState(m)
        case m.Id == MsgAddAction:
            cd.addAction(m.Action)
        }
    }
}

func (cd CmpData) update() {
    for _, v := range cd.actions {
        v.Act(cd.states)
    }
}

func (cd CmpData) sendState(msg CmpMsg) {
    state := cd.states[msg.StateId]
    msg.StateReply <- state
}
