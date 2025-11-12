package filereviewer

import (
	"sync"

	"github.com/Caritas-Team/reviewer/internal/logger"
)

// Канал для ключей файлов
var bufferCh = make(chan string, 10)

// Канал для удаления файлов
var cleanUpCh = make(chan string, 10)

// Условный файл
type UserData struct {
	Name     string
	Age      int
	Keywords []string
	Value    float64
}

// Глобальная переменная логгера
var GlobalLogger *logger.Logger

// Начнём обработку
func StartProcessing() {
	var wg sync.WaitGroup
	wg.Add(2)

	// Запустим рабочую горутину
	go worker(&wg)

	// Запустим горутину очистки
	go cleanupWorker(&wg)

	// Заполним канал несколькими примерами ключей
	go func() {
		bufferCh <- "randomkey1"
		bufferCh <- "randomkey2"
	}()

	// Ждём завершение обработки
	wg.Wait()

	// Закрываем каналы
	close(bufferCh)
	close(cleanUpCh)
}

// Рабочая горутина
func worker(wg *sync.WaitGroup) {
	defer wg.Done()
	for key := range bufferCh {
		// Берём файлы из Memcached (фиктивная реализация)
		data1, data2 := fetchUserDataFromCache(key)

		// Проверяем имена пользователей и сравниваем остальные поля
		result := compareUserData(data1, data2)

		// Отправляем результат на фронтенд
		err := sendResponse(result)
		if err != nil {
			GlobalLogger.Error("Ошибка отправки:", "key", key, "err", err)
		}

		// Сообщение об необходимости удаления файла
		cleanUpCh <- key
	}
}

// Горутина для очистки данных
func cleanupWorker(wg *sync.WaitGroup) {
	defer wg.Done()
	for key := range cleanUpCh {
		// Вызываем фиктивную процедуру очистки
		fakeCleanup(key)
	}
}

// Заглушка получения данных пользователя из кэша
func fetchUserDataFromCache(key string) (*UserData, *UserData) {
	return &UserData{
			Name:     "Иван",
			Age:      30,
			Keywords: []string{"hello", "world"},
			Value:    42.0,
		},
		&UserData{
			Name:     "Иван",
			Age:      31,
			Keywords: []string{"goodbye", "universe"},
			Value:    43.0,
		}
}

// Метод сравнения данных пользователей, просто чтобы прикинуть принцип работы
func compareUserData(user1, user2 *UserData) map[string]interface{} {
	comparisonResult := make(map[string]interface{})

	if user1.Name == user2.Name {
		comparisonResult["Name"] = true
	} else {
		comparisonResult["Name"] = false
	}

	// Сравнение возраста
	comparisonResult["AgeDifference"] = abs(user1.Age - user2.Age)

	// Разница ключевых слов
	commonKeywords := findCommonElements(user1.Keywords, user2.Keywords)
	differentKeywords := append(difference(user1.Keywords, user2.Keywords),
		difference(user2.Keywords, user1.Keywords)...)

	comparisonResult["CommonKeywords"] = commonKeywords
	comparisonResult["DifferentKeywords"] = differentKeywords

	// Разница значений
	valueDiff := user1.Value - user2.Value
	comparisonResult["ValueDifference"] = valueDiff

	return comparisonResult
}

// Вспомогательные функции
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func difference(a, b []string) []string {
	mb := make(map[string]bool)
	for _, item := range b {
		mb[item] = true
	}
	var diff []string
	for _, item := range a {
		if !mb[item] {
			diff = append(diff, item)
		}
	}
	return diff
}

func findCommonElements(a, b []string) []string {
	mb := make(map[string]bool)
	for _, item := range b {
		mb[item] = true
	}
	var common []string
	for _, item := range a {
		if mb[item] {
			common = append(common, item)
		}
	}
	return common
}

// Фиктивный метод отправки данных на фронтенд
func sendResponse(data interface{}) error {
	GlobalLogger.Info("Отправлено на фронтенд:", "data", data)
	return nil
}

// Фиктивный метод очистки
func fakeCleanup(key string) {
	GlobalLogger.Info("Фиктивное удаление данных для ключа:", "key", key)
}
