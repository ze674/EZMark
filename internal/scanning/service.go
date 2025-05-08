package scanning

import (
	"FileMarker/internal/models"
	"context"
	"log"
	"sync"
	"time"
)

// ScanningService управляет процессом сканирования
type ScanningService struct {
	mu           sync.Mutex
	scanner      Scanner
	running      bool
	activeTask   *models.Task
	scanInterval time.Duration
	codes        map[string]*models.Code
	cancelFunc   context.CancelFunc
	results      []ScanResult
}

// ScanResult представляет результат сканирования
type ScanResult struct {
	Code      string
	Valid     bool
	Error     string
	Timestamp time.Time
}

// NewScanningService создает новый сервис сканирования
func NewScanningService(scanner Scanner, scanInterval time.Duration) *ScanningService {
	return &ScanningService{
		scanner:      scanner,
		running:      false,
		scanInterval: scanInterval,
		codes:        make(map[string]*models.Code),
		results:      make([]ScanResult, 0),
	}
}

// StartScanning начинает процесс сканирования с указанным заданием
func (s *ScanningService) StartScanning(task *models.Task, codes []*models.Code) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil // Уже запущено
	}

	// Подключаемся к сканеру
	if err := s.scanner.Connect(); err != nil {
		return err
	}

	// Загружаем задание и коды
	s.activeTask = task
	s.codes = make(map[string]*models.Code)
	for _, code := range codes {
		s.codes[code.Value] = code
	}

	// Создаем контекст с возможностью отмены
	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFunc = cancel

	// Запускаем процесс сканирования в отдельной горутине
	go s.scanLoop(ctx)

	s.running = true
	return nil
}

// StopScanning останавливает процесс сканирования
func (s *ScanningService) StopScanning() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil // Уже остановлено
	}

	// Отменяем контекст, что приведет к завершению цикла сканирования
	if s.cancelFunc != nil {
		s.cancelFunc()
		s.cancelFunc = nil
	}

	// Закрываем соединение со сканером
	if err := s.scanner.Close(); err != nil {
		return err
	}

	s.running = false
	return nil
}

// IsRunning возвращает текущее состояние сервиса
func (s *ScanningService) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// GetResults возвращает текущие результаты сканирования
func (s *ScanningService) GetResults() []ScanResult {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Создаем копию, чтобы избежать проблем с конкурентным доступом
	results := make([]ScanResult, len(s.results))
	copy(results, s.results)

	return results
}

// GetActiveTask возвращает текущее активное задание
func (s *ScanningService) GetActiveTask() *models.Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.activeTask
}

// scanLoop выполняет цикл сканирования
func (s *ScanningService) scanLoop(ctx context.Context) {
	ticker := time.NewTicker(s.scanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Выполняем сканирование
			response, err := s.scanner.Scan()
			if err != nil {
				log.Printf("Ошибка сканирования: %v", err)
				continue
			}

			// Обрабатываем результат
			s.processResponse(response)

		case <-ctx.Done():
			// Контекст отменен, завершаем работу
			return
		}
	}
}

// processResponse обрабатывает ответ от сканера
func (s *ScanningService) processResponse(response string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Проверяем, что это не пустой ответ и не специальное значение "NoRead"
	if response == "" || response == "NoRead" {
		// Добавляем ошибочный результат
		s.results = append(s.results, ScanResult{
			Code:      response,
			Valid:     false,
			Error:     "Код не распознан",
			Timestamp: time.Now(),
		})
		return
	}

	// Проверяем наличие кода в списке кодов
	code, exists := s.codes[response]
	if !exists {
		// Код не найден в списке
		s.results = append(s.results, ScanResult{
			Code:      response,
			Valid:     false,
			Error:     "Код не найден в задании",
			Timestamp: time.Now(),
		})
		return
	}

	// Проверяем, не был ли код уже отсканирован
	if code.Scanned {
		// Код уже был отсканирован
		s.results = append(s.results, ScanResult{
			Code:      response,
			Valid:     false,
			Error:     "Код уже был отсканирован",
			Timestamp: time.Now(),
		})
		return
	}

	// Отмечаем код как отсканированный
	code.Scanned = true
	code.ScannedAt = time.Now()

	// Добавляем успешный результат
	s.results = append(s.results, ScanResult{
		Code:      response,
		Valid:     true,
		Error:     "",
		Timestamp: time.Now(),
	})

	// Логируем успешное сканирование
	log.Printf("Успешно отсканирован код: %s", response)
}
