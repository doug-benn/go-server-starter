package middleware

import "net/http"

type Middleware func(http.Handler) http.Handler

//A list of middlewares
type Chain struct {
	middlewares []Middleware
}

//Create a new list of middlewares
func NewChain(middlewares ...Middleware) Chain {
	return Chain{append(([]Middleware)(nil), middlewares...)}
}

//Builds the chain of middlewares and returns a http.Handler
func (chain Chain) Build(handler http.Handler) http.Handler {
	if handler == nil {
		handler = http.NewServeMux()
	}

	for i := range chain.middlewares {
		handler = chain.middlewares[len(chain.middlewares)-1-i](handler)
	}

	return handler
}

//Example of a wrapped middleware
// func NewExampleMiddleware(someThing string) func(http.Handler) http.Handler {
// 	return func(next http.Handler) http.Handler {
// 			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 					// Pre-processing logic
// 					defer func() {
// 							// Post-processing logic (if needed)
// 					}()
// 					next.ServeHTTP(w, r)
// 			})
// 	}
// }
