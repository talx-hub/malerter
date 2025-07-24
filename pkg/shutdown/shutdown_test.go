package shutdown

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/talx-hub/malerter/internal/logger"
)

func TestIdleShutdown(t *testing.T) {
	idleCh := make(chan struct{})
	log := logger.NewNopLogger()

	called := false

	go IdleShutdown(idleCh, log, func(...any) error {
		called = true
		return nil
	})

	// Даем горутине время на установку обработчика
	time.Sleep(100 * time.Millisecond)

	// Симулируем SIGINT
	p, err := os.FindProcess(os.Getpid())
	assert.NoError(t, err)
	err = p.Signal(os.Interrupt)
	assert.NoError(t, err)

	// Ждем завершения shutdown
	select {
	case <-idleCh:
	case <-time.After(2 * time.Second):
		t.Fatal("IdleShutdown did not return in time")
	}

	assert.True(t, called, "cancelFunc should have been called")
}

func ExampleIdleShutdown() {
	idleCh := make(chan struct{})
	log := logger.NewNopLogger()

	go IdleShutdown(idleCh, log, func(...any) error {
		fmt.Println("Cleanup complete")
		return nil
	})

	// Симуляция SIGINT через секунду
	go func() {
		time.Sleep(1 * time.Second)
		p, _ := os.FindProcess(os.Getpid())
		_ = p.Signal(os.Interrupt)
	}()

	<-idleCh
	fmt.Println("Shutdown complete")

	// Output:
	// Cleanup complete
	// Shutdown complete
}
