package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type Info struct {
	ID    int       `json:"id"`
	Name  string    `json:"name"`
	Price float64   `json:"price"`
	Date  time.Time `json:"date"`
}

func writeToClickhouse() {
	// Инициализация логгера
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel) // Установите уровень логирования по необходимости

	// Подключение к ClickHouse. Замените на ваши данные.
	connStr := fmt.Sprintf("tcp://%s:%d?debug=true", os.Getenv("CLICKHOUSE_HOST"), os.Getenv("CLICKHOUSE_PORT"))

	conn, err := sql.Open("clickhouse", connStr)
	if err != nil {
		logger.Fatalf("Ошибка при подключении к ClickHouse: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			logger.Printf("Ошибка при закрытии соединения с ClickHouse: %v", err)
		}
	}()

	// Проверка подключения
	err = conn.Ping()
	if err != nil {
		logger.Fatalf("Ошибка при проверке подключения к ClickHouse: %v", err)
	}
	logger.Info("Подключение к ClickHouse успешно установлено.")

	// Данные для вставки (симулируем большой объем данных)
	products := generateProducts(100000) // 100 000 продуктов

	// Вставка данных пакетно
	batchSize := 1000 // Размер пакета
	for i := 0; i < len(products); i += batchSize {
		end := i + batchSize
		if end > len(products) {
			end = len(products)
		}
		batch := products[i:end]
		err := insertBatch(conn, batch, logger)
		if err != nil {
			logger.Errorf("Ошибка при вставке пакета данных: %v", err)
			// Обработка ошибки, например, повторная попытка
		}
	}

	logger.Info("Данные успешно вставлены в ClickHouse.")
}

func generateProducts(count int) []Info {
	products := make([]Info, count)
	for i := 0; i < count; i++ {
		products[i] = Info{
			ID:    i + 1,
			Name:  fmt.Sprintf("Info %d", i+1),
			Price: float64(i+1) * 0.99,
			Date:  time.Now(),
		}
	}
	return products
}

func insertBatch(conn *sql.DB, products []Info, logger *logrus.Logger) error {
	ctx := context.Background()
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("ошибка при начале транзакции: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			logger.Printf("Ошибка при откате транзакции: %v", err)
		}
	}()

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO products (id, name, price, date) VALUES (?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("ошибка при подготовке запроса: %w", err)
	}
	defer stmt.Close()

	for _, p := range products {
		_, err = stmt.ExecContext(ctx, p.ID, p.Name, p.Price, p.Date)
		if err != nil {
			return fmt.Errorf("ошибка при вставке данных: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("ошибка при фиксации транзакции: %w", err)
	}
	return nil
}
