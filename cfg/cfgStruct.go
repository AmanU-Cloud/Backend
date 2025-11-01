package cfg

// Структура для представления конфигурации логгера
type Config struct {
	Logging struct {
		Level  string `yaml:"level"`  // Уровень логирования
		Format string `yaml:"format"` // Формат логирования
	} `yaml:"logging"`
}
