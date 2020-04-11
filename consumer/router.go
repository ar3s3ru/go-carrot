package consumer

type Router interface {
	Register(Handler, Binding)

	Use(middlewares ...func(Handler) Handler)
	With(middlewares ...func(Handler) Handler) Router
	Group(func(Router)) Router
}

func NewRouterMux() Router {
	panic("implement me")
}
