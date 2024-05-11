package handler_test

import (
	"os"
	"testing"
	"time"

	"github.com/Truenya/dota-search-back/cache"
	"github.com/Truenya/dota-search-back/data"
	"github.com/Truenya/dota-search-back/db"
	"github.com/Truenya/dota-search-back/handlers"
	"github.com/go-playground/assert/v2"
)

func TestMain(m *testing.M) {
	pg := db.Init("dotasearch_test")
	cache.Init()

	code := m.Run()

	pg.Close()
	os.Exit(code)
}

func TestHandlers(t *testing.T) {
	ip := "172.31.36.0"
	player := data.Player{
		IP:   ip,
		Data: "TestData",
	}

	defer db.DeleteP(ip)
	defer db.DeleteC(ip)
	t.Parallel()

	handlers.TooEarly = 0
	result := handlers.AddPlayer(player)
	assert.Equal(t, result, handlers.JustAdded)
	result = handlers.AddPlayer(player)
	assert.Equal(t, result, handlers.Updated)
	result = handlers.AddCommand(data.Command{Player: player})
	assert.Equal(t, result, handlers.DeletedAndAdded)
	handlers.TooEarly = 1 * time.Second
	result = handlers.AddCommand(data.Command{Player: player})
	assert.Equal(t, result, handlers.TooManyRequests)
}
