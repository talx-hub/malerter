// Package pgcontainer предоставляет вспомогательные средства для запуска PostgreSQL в Docker-контейнере
// с помощью библиотеки dockertest. Этот пакет предназначен для использования в интеграционных тестах.
//
// Контейнер запускается с правами суперпользователя PostgreSQL, после чего автоматически создаётся
// тестовая база данных и отдельный пользователь. Полученный DSN можно использовать для подключения
// к изолированной среде тестирования.
//
// Использование:
//
//	c := pgcontainer.New(log)
//	if err := c.RunContainer(); err != nil {
//	    log.Fatal().Err(err).Msg("failed to run pg container")
//	}
//	defer c.Close()
//
//	db, err := sql.Open("pgx", c.GetDSN())
package pgcontainer

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"

	"github.com/talx-hub/malerter/internal/logger"
)

const (
	testDBName       = "test" // Имя создаваемой тестовой базы
	testUserName     = "test" // Имя пользователя для тестовой базы
	testUserPassword = "test" // Пароль пользователя
)

const (
	queryTimeOut = 5 * time.Second  // Время ожидания SQL-запросов
	retryTimeout = 10 * time.Second // Максимальное время ожидания запуска контейнера
)

// PGContainer управляет запуском, подключением и остановкой PostgreSQL-контейнера для тестирования.
type PGContainer struct {
	dockerPool  *dockertest.Pool
	pgContainer *dockertest.Resource
	log         *logger.ZeroLogger
	dsn         string
}

// New создаёт и возвращает новую структуру PGContainer.
//
// Требует передать логгер, который будет использоваться для вывода ошибок.
func New(log *logger.ZeroLogger) *PGContainer {
	return &PGContainer{
		log: log,
	}
}

// GetDSN возвращает строку подключения к тестовой базе данных.
//
// Можно использовать её для подключения к базе через pgx или database/sql.
func (c *PGContainer) GetDSN() string {
	return c.dsn
}

// Close останавливает и удаляет PostgreSQL-контейнер, если он был запущен.
//
// Все ресурсы Docker очищаются через dockertest.Pool.Purge.
func (c *PGContainer) Close() {
	if c.pgContainer == nil {
		return
	}

	if err := c.dockerPool.Purge(c.pgContainer); err != nil {
		c.log.Error().Err(err).
			Msg("failed to purge the postgres container")
	}
}

// RunContainer запускает PostgreSQL-контейнер, ожидает его готовности и создаёт тестовую БД.
//
// Также создаётся пользователь с правами доступа к тестовой базе.
// Используется dockertest для управления Docker-контейнером и pgx для подключения к базе данных.
//
// Возвращает ошибку, если контейнер не удалось запустить, подключиться или создать БД.
func (c *PGContainer) RunContainer() error {
	var err error
	c.dockerPool, err = dockertest.NewPool("")
	if err != nil {
		return fmt.Errorf("failed to initialize a docker pool: %w", err)
	}

	const pgPort = "5432/tcp"
	c.pgContainer, err = c.dockerPool.RunWithOptions(
		&dockertest.RunOptions{
			Repository: "postgres",
			Tag:        c.loadImageFromEnv(),
			Env: []string{
				"POSTGRES_USER=postgres",
				"POSTGRES_PASSWORD=postgres",
			},
			ExposedPorts: []string{pgPort},
		},
		func(config *docker.HostConfig) {
			config.AutoRemove = true
			config.RestartPolicy = docker.RestartPolicy{Name: "no"}
		},
	)
	if err != nil {
		return fmt.Errorf("failed to run postgres container: %w", err)
	}

	hostPort := c.pgContainer.GetHostPort(pgPort)

	c.dockerPool.MaxWait = retryTimeout
	var conn *pgx.Conn
	if err := c.dockerPool.Retry(func() error {
		conn, err = getSUConnection(hostPort)
		if err != nil {
			return fmt.Errorf("failed to connect to the DB: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("retry failed: %w", err)
	}
	defer func() {
		if err = conn.Close(context.TODO()); err != nil {
			c.log.Error().Err(err).
				Msg("failed to correctly close the DB connection")
		}
	}()

	if c.dsn, err = createTestDB(conn); err != nil {
		return fmt.Errorf("failed to create a test DB: %w", err)
	}
	return nil
}

// loadImageFromEnv загружает тег образа PostgreSQL из файла .env.
//
// Имя переменной окружения: POSTGRES_TAG.
func (c *PGContainer) loadImageFromEnv() string {
	if err := godotenv.Load(".env"); err != nil {
		c.log.Error().Err(err).
			Msg("error loading .env file: %v")
	}
	return os.Getenv("POSTGRES_TAG")
}

// getSUConnection создаёт соединение от имени суперпользователя postgres.
//
// Используется для выполнения CREATE USER и CREATE DATABASE.
func getSUConnection(hostPort string) (*pgx.Conn, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable",
		"postgres",
		"postgres",
		hostPort,
		"postgres",
	)
	conn, err := pgx.Connect(context.TODO(), dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to get a super user connection: %w", err)
	}

	return conn, nil
}

// createTestDB создаёт тестовую базу данных и пользователя,
// затем формирует DSN для подключения к ней.
func createTestDB(conn *pgx.Conn) (string, error) {
	const (
		createUser = `CREATE USER %s PASSWORD '%s';`
		createDB   = `CREATE DATABASE %s
		OWNER %s
		ENCODING 'UTF8'
		LC_COLLATE = 'en_US.utf8'
		LC_CTYPE = 'en_US.utf8';`
	)

	ctx, cancel1 := context.WithTimeout(context.Background(), queryTimeOut)
	defer cancel1()
	_, err := conn.Exec(ctx, fmt.Sprintf(createUser, testUserName, testUserPassword))
	if err != nil {
		return "", fmt.Errorf("failed to create a test user: %w", err)
	}

	ctx, cancel2 := context.WithTimeout(context.Background(), queryTimeOut)
	defer cancel2()
	_, err = conn.Exec(ctx, fmt.Sprintf(createDB, testDBName, testUserName))
	if err != nil {
		return "", fmt.Errorf("failed to create a test DB: %w", err)
	}

	testDatabaseDSN := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable",
		testUserName,
		testUserPassword,
		net.JoinHostPort(
			conn.Config().Host,
			strconv.FormatUint(uint64(conn.Config().Port), 10)),
		testDBName,
	)

	return testDatabaseDSN, nil
}
