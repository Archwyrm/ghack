package core

type ServiceMsg interface {
    Name() string
}

type Service interface {
    Run(input chan ServiceMsg)
}
