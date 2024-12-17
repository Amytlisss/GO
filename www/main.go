package main

func main() {
	initDB()        // Инициализация базы данных
	defer closeDB() // Закрытие базы данных
	handleRequest() // Обработка запросов
}
