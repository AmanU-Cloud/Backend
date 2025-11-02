package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Время обработки одного PDF-файла (в секундах)
	fileProcessingTimeSeconds = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "pdf_service",
		Name:      "file_processing_time_seconds",
		Help:      "Время обработки одного PDF-файла (в секундах)",
	}, []string{"result"})

	// Размер загруженных файлов (в байтах)
	fileSizeBytes = promauto.NewSummary(prometheus.SummaryOpts{
		Namespace: "calculator_service",
		Name:      "file_size_bytes",
		Help:      "Размер загруженных файлов (в байтах)",
	})

	// Количество успешно загруженных файлов
	fileUploadSuccessCount = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "pdf_service",
		Name:      "file_upload_success_count",
		Help:      "Количество успешно загруженных файлов",
	})

	// Количество ошибок при загрузке файлов
	fileUploadErrorCount = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "pdf_service",
		Name:      "file_upload_error_count",
		Help:      "Количество ошибок при загрузке файлов",
	})

	// Текущее количество файлов, находящихся в процессе обработки
	currentFilesInProgress = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "pdf_service",
		Name:      "current_files_in_progress",
		Help:      "Текущее количество файлов, находящихся в процессе обработки",
	})

	// Количество операций в секунду
	operationsPerSecond = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "pdf_service",
		Name:      "operations_per_second",
		Help:      "Количество операций в секунду",
	})

	// Длина очереди задач, ожидающих обработки.
	queueLength = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "pdf_service",
		Name:      "queue_length",
		Help:      "Длина очереди задач, ожидающих обработки",
	})

	// Задержка между постановкой задачи в очередь и началом её обработки (в секундах)
	workerQueueDelaySeconds = promauto.NewSummary(prometheus.SummaryOpts{
		Namespace: "pdf_service",
		Name:      "worker_queue_delay_seconds",
		Help:      "Задержка между постановкой задачи в очередь и началом её обработки (в секундах)",
	})

	// Время работы сервера (в секундах)
	serverUptimeSeconds = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "pdf_service",
		Name:      "server_uptime_seconds",
		Help:      "Время работы сервера (в секундах)",
	})

	// Использование оперативной памяти (в байтах)
	memoryUsageBytes = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "pdf_service",
		Name:      "memory_usage_bytes",
		Help:      "Использование оперативной памяти (в байтах)",
	})

	// Средняя загрузка ЦПУ за последнюю минуту
	cpuLoadAverage = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "pdf_service",
		Name:      "cpu_load_average",
		Help:      "Средняя загрузка ЦПУ за последнюю минуту",
	})

	// Количество попыток повторной обработки при возникновении ошибок
	retryAttempts = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "pdf_service",
		Name:      "retry_attempts",
		Help:      "Количество попыток повторной обработки при возникновении ошибок",
	})

	// Счётчик статусов операций (NEW, PROGRESS, DONE, ERROR)
	operationStatusCounts = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "pdf_service",
		Name:      "operation_status_counts",
		Help:      "Счётчик статусов операций (NEW, PROGRESS, DONE, ERROR)",
	}, []string{"status"})

	// Количество успешных обращений к Memcached
	cacheHits = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "pdf_service",
		Name:      "cache_hits",
		Help:      "Количество успешных обращений к Memcached",
	})

	// Количество пропущенных записей в Memcached.
	cacheMisses = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "pdf_service",
		Name:      "cache_misses",
		Help:      "Количество пропущенных записей в Memcached",
	})

	// Количество превышений лимита запросов
	rateLimitExceededCount = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "pdf_service",
		Name:      "rate_limit_exceeded_count",
		Help:      "Количество превышений лимита запросов",
	})

	// Количество запросов от каждого IP-адреса
	requestCountByIP = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "pdf_service",
		Name:      "request_count_by_ip",
		Help:      "Количество запросов от каждого IP-адреса",
	}, []string{"ip"})

	// Количество успешных извлечений данных из PDF-файлов
	dataExtractionSuccessCount = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "pdf_service",
		Name:      "data_extraction_success_count",
		Help:      "Количество успешных извлечений данных из PDF-файлов",
	})

	// Количество ошибок при извлечении данных
	dataExtractionErrorCount = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "pdf_service",
		Name:      "data_extraction_error_count",
		Help:      "Количество ошибок при извлечении данных",
	})

	// Время извлечения данных из одного PDF-файла (в секундах)
	dataExtractionTimeSeconds = promauto.NewSummary(prometheus.SummaryOpts{
		Namespace: "pdf_service",
		Name:      "data_extraction_time_seconds",
		Help:      "Время извлечения данных из одного PDF-файла (в секундах)",
	})

	// Количество успешных сравнений "до/после"
	comparisonSuccessCount = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "pdf_service",
		Name:      "comparison_success_count",
		Help:      "Количество успешных сравнений до/после",
	})

	// Количество ошибок при сравнении "до/после"
	comparisonErrorCount = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "pdf_service",
		Name:      "comparison_error_count",
		Help:      "Количество ошибок при сравнении до/после",
	})

	// Время выполнения сравнения "до/после" (в секундах)
	comparisonTimeSeconds = promauto.NewSummary(prometheus.SummaryOpts{
		Namespace: "pdf_service",
		Name:      "comparison_time_seconds",
		Help:      "Время выполнения сравнения до/после (в секундах)",
	})

	// Количество успешных экспортов отчётов
	exportSuccessCount = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "pdf_service",
		Name:      "export_success_count",
		Help:      "Количество успешных экспортов отчётов",
	})

	// Количество ошибок при экспорте отчётов
	exportErrorCount = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "pdf_service",
		Name:      "export_error_count",
		Help:      "Количество ошибок при экспорте отчётов",
	})

	// Время выполнения экспорта отчётов (в секундах)
	exportTimeSeconds = promauto.NewSummary(prometheus.SummaryOpts{
		Namespace: "pdf_service",
		Name:      "export_time_seconds",
		Help:      "Время выполнения экспорта отчётов (в секундах)",
	})
)

// UpdateFileProcessingTime обновляет время обработки файла
func UpdateFileProcessingTime(result string, duration float64) {
	fileProcessingTimeSeconds.WithLabelValues(result).Observe(duration)
}

// UpdateFileSize обновляет размер загруженного файла
func UpdateFileSize(size float64) {
	fileSizeBytes.Observe(size)
}

// UpdateFileUploadSuccess увеличивает счётчик успешных загрузок файлов
func UpdateFileUploadSuccess() {
	fileUploadSuccessCount.Inc()
}

// UpdateFileUploadError увеличивает счётчик ошибок при загрузке файлов
func UpdateFileUploadError() {
	fileUploadErrorCount.Inc()
}

// UpdateCurrentFilesInProgress обновляет текущее количество файлов в процессе обработки
func UpdateCurrentFilesInProgress(count float64) {
	currentFilesInProgress.Set(count)
}

// UpdateOperationsPerSecond увеличивает счётчик операций в секунду
func UpdateOperationsPerSecond() {
	operationsPerSecond.Inc()
}

// UpdateQueueLength обновляет длину очереди задач
func UpdateQueueLength(length float64) {
	queueLength.Set(length)
}

// UpdateWorkerQueueDelay обновляет задержку между постановкой задачи в очередь и началом её обработки
func UpdateWorkerQueueDelay(delay float64) {
	workerQueueDelaySeconds.Observe(delay)
}

// UpdateServerUptime обновляет время работы сервера
func UpdateServerUptime() {
	go func() {
		startTime := time.Now()
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			uptime := time.Since(startTime).Seconds()
			serverUptimeSeconds.Set(uptime)
		}
	}()
}

// UpdateMemoryUsage обновляет использование оперативной памяти
func UpdateMemoryUsage(bytes float64) {
	memoryUsageBytes.Set(bytes)
}

// UpdateCPULoadAverage обновляет среднюю загрузку ЦПУ
func UpdateCPULoadAverage(load float64) {
	cpuLoadAverage.Set(load)
}

// UpdateRetryAttempts увеличивает счётчик попыток повторной обработки
func UpdateRetryAttempts() {
	retryAttempts.Inc()
}

// UpdateOperationStatus увеличивает счётчик статусов операций
func UpdateOperationStatus(status string) {
	operationStatusCounts.WithLabelValues(status).Inc()
}

// UpdateCacheHits увеличивает счётчик успешных обращений к Memcached
func UpdateCacheHits() {
	cacheHits.Inc()
}

// UpdateCacheMisses увеличивает счётчик пропущенных записей в Memcached
func UpdateCacheMisses() {
	cacheMisses.Inc()
}

// UpdateRateLimitExceeded увеличивает счётчик превышений лимита запросов
func UpdateRateLimitExceeded() {
	rateLimitExceededCount.Inc()
}

// UpdateRequestCountByIP увеличивает счётчик запросов от каждого IP-адреса
func UpdateRequestCountByIP(ip string) {
	requestCountByIP.WithLabelValues(ip).Inc()
}

// UpdateDataExtractionSuccess увеличивает счётчик успешных извлечений данных из PDF-файлов
func UpdateDataExtractionSuccess() {
	dataExtractionSuccessCount.Inc()
}

// UpdateDataExtractionError увеличивает счётчик ошибок при извлечении данных
func UpdateDataExtractionError() {
	dataExtractionErrorCount.Inc()
}

// UpdateDataExtractionTime обновляет время извлечения данных из одного PDF-файла
func UpdateDataExtractionTime(duration float64) {
	dataExtractionTimeSeconds.Observe(duration)
}

// UpdateComparisonSuccess увеличивает счётчик успешных сравнений "до/после"
func UpdateComparisonSuccess() {
	comparisonSuccessCount.Inc()
}

// UpdateComparisonError увеличивает счётчик ошибок при сравнении "до/после"
func UpdateComparisonError() {
	comparisonErrorCount.Inc()
}

// UpdateComparisonTime обновляет время выполнения сравнения "до/после"
func UpdateComparisonTime(duration float64) {
	comparisonTimeSeconds.Observe(duration)
}

// UpdateExportSuccess увеличивает счётчик успешных экспортов отчётов
func UpdateExportSuccess() {
	exportSuccessCount.Inc()
}

// UpdateExportError увеличивает счётчик ошибок при экспорте отчётов
func UpdateExportError() {
	exportErrorCount.Inc()
}

// UpdateExportTime обновляет время выполнения экспорта отчётов
func UpdateExportTime(duration float64) {
	exportTimeSeconds.Observe(duration)
}

func InitMetrics() {
	http.Handle("/metrics", promhttp.Handler())
}
