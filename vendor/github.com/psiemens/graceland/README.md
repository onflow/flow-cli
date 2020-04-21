# graceland

Graceful shutdown for multi-routine Go programs.

![Paul & Chevy](https://media.giphy.com/media/uJRZ6N5ZlYRvq/source.gif)

> The Mississippi Delta was shining like a national guitar 🎸

## Usage

```go
import (
    "net/http"

    "github.com/psiemens/graceland"
)

// Server implements the graceland.Routine interface.
type Server struct {
	server *http.Server
}

func NewServer() *Server {
	server := &http.Server{
		Addr: ":8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "Hello, World!")
		}),
	}

	return &Server{server: server}
}

func (h *Server) Start() error {
	err := h.server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err
}

func (h *Server) Stop() {
	h.server.Shutdown(context.Background())
}

// Worker implements the graceland.Routine interface.
type Worker struct {
	ticker *time.Ticker
	done   chan bool
}

func NewWorker() *Worker {
	return &Worker{
		ticker: time.NewTicker(time.Second),
		done:   make(chan bool, 1),
	}
}

func (t *Worker) Start() error {
	for {
		select {
		case <-t.ticker.C:
			// do work
		case <-t.done:
			return nil
		}
	}
}

func (t *Worker) Stop() {
	t.done <- true
}

func main() {
	// create a new routine group
	g := graceland.NewGroup()

	server := NewServer()
	worker := NewWorker()

	// add server and worker to group
	g.Add(server)
	g.Add(worker)

	// stop group after 3 seconds
	go func() {
		time.Sleep(time.Second * 3)
		g.Stop()
	}()

	// start group and block until shutdown
	err := g.Start()
	if err != nil {
		fmt.Printf("Shut down with error: %s\n", err.Error())
		return
	}

	fmt.Print("Shut down with no error\n")
}
```
