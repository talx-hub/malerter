// Package shutdown содержит вспомогательные функции для корректного завершения работы приложения
// по сигналу прерывания (например, Ctrl+C).
//
// Позволяет отреагировать на системные сигналы завершения (SIGINT),
// выполнить завершающие операции и корректно остановить сервис.
package shutdown

import (
	"os"
	"os/signal"

	"github.com/talx-hub/malerter/internal/logger"
)

// CancelFunc — функция, вызываемая при завершении работы приложения.
//
// Обычно используется для отмены контекста, закрытия соединений и других операций завершения.
type CancelFunc func(args ...any) error

// IdleShutdown блокирует выполнение до получения сигнала завершения (SIGINT) и
// затем вызывает переданную функцию отмены.
//
// Параметры:
//   - idleCh: канал, который будет закрыт при завершении (можно использовать для синхронизации);
//   - log: логгер для вывода сообщений;
//   - cancelFunc: функция, которая будет вызвана перед завершением.
//
// После получения сигнала SIGINT (обычно по Ctrl+C) логгирует сообщение,
// вызывает cancelFunc и закрывает канал `idleCh`.
//
// Пример использования:
//
//	func main() {
//	    idleCh := make(chan struct{})
//	    log := logger.New("my-service")
//	    go shutdown.IdleShutdown(idleCh, log, func(...any) error {
//	        // clean up
//	        return nil
//	    })
//	    <-idleCh // блокировка до завершения
//	}
func IdleShutdown(
	idleCh chan struct{},
	log *logger.ZeroLogger,
	cancelFunc CancelFunc,
) {
	defer close(idleCh)

	sigintCh := make(chan os.Signal, 1)
	signal.Notify(sigintCh, os.Interrupt)
	<-sigintCh

	log.Info().Msg("shutdown signal received. Exiting....")
	if err := cancelFunc(); err != nil {
		log.Error().Err(err).Msg("error during service shutdown")
	}
}
