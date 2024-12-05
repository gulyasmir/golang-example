package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Info struct {
	ID        int       `json:"id"`
	EventTime time.Time `json:"event_time"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
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
			ID:        i + 1,
			EventTime: time.Now(),
			Level:     fmt.Sprintf("Level %d", i+1),
			Message:   fmt.Sprintf("Message %d", i+1),
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
		_, err = stmt.ExecContext(ctx, p.ID, p.EventTime, p.Level, p.Message)
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

func readFromPostgres() {
	// Подключение к базе данных.
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Проверка подключения
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	// Запрос к базе данных
	rows, err := db.Query("SELECT id, name_user, email FROM users")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// Обработка результатов
	users := []User{}
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Name, &user.Email)
		if err != nil {
			log.Fatal(err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	// Вывод результатов
	fmt.Println("users:")
	for _, user := range users {
		fmt.Printf("ID: %d, Name: %s, Email: %s\n", user.ID, user.Name, user.Email)
	}
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, Docker!")
	})

	fmt.Println("Server is running on port 8080...")
	http.ListenAndServe(":8080", nil)

	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")

	fmt.Printf("DB_HOST: %s\n", dbHost)
	fmt.Printf("DB_PORT: %s\n", dbPort)

	//writeToClickhouse()
	//readFromPostgres()
}
