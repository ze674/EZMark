package web

import (
	"FileMarker/internal/config"
	"FileMarker/internal/database"
	"FileMarker/internal/filemanager"
	"FileMarker/internal/models"
	"FileMarker/internal/view/pages"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
)

// TaskViewModel представляет модель представления для задания
type TaskViewModel struct {
	ID          string
	GTIN        string
	Date        string
	BatchNumber string
	TotalCodes  int
	Status      string
	FilePath    string
}

// Server представляет веб-сервер
type Server struct {
	config     config.Config
	templates  *template.Template
	dirScanner *filemanager.DirectoryScanner
	db         *database.DB
}

// NewServer создает новый экземпляр сервера
func NewServer(cfg config.Config, db *database.DB) (*Server, error) {
	// Загружаем шаблоны
	tmpl, err := template.ParseGlob("templates/*.html")
	if err != nil {
		return nil, err
	}

	// Создаем сканер директории
	scanner := filemanager.NewDirectoryScanner(cfg.IncomingDir)

	return &Server{
		config:     cfg,
		templates:  tmpl,
		dirScanner: scanner,
		db:         db,
	}, nil
}

// Start запускает веб-сервер
func (s *Server) Start() error {
	// Настраиваем обработчики маршрутов
	http.HandleFunc("/", s.handleHome)
	http.HandleFunc("/tasks", s.handleTasksList)
	http.HandleFunc("/tasks/", s.handleTasksActions)
	http.HandleFunc("/active-task", s.handleActiveTask)
	http.HandleFunc("/scan-code", s.handleScanCode)
	http.HandleFunc("/complete-task", s.handleCompleteTask)

	// Статические файлы
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Запускаем HTTP-сервер
	log.Printf("Запуск сервера на порту %s", s.config.ServerPort)
	return http.ListenAndServe(":"+s.config.ServerPort, nil)
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	homePage := pages.Home()
	homePage.Render(r.Context(), w)
}

// handleTasksList обрабатывает запрос на список заданий
func (s *Server) handleTasksList(w http.ResponseWriter, r *http.Request) {
	// Получаем список файлов только при запросе страницы
	files, err := s.dirScanner.ListMarkFiles()
	if err != nil {
		http.Error(w, "Ошибка при получении списка файлов", http.StatusInternalServerError)
		return
	}

	// Преобразуем файлы в список задач для отображения
	var tasks []TaskViewModel
	for _, filePath := range files {
		task, _, err := filemanager.ParseMarkFile(filePath)
		if err != nil {
			log.Printf("Ошибка при парсинге файла %s: %v", filePath, err)
			continue
		}

		// Добавляем задачу в список
		tasks = append(tasks, TaskViewModel{
			ID:          task.ID,
			GTIN:        task.GTIN,
			Date:        task.Date,
			BatchNumber: task.BatchNumber,
			TotalCodes:  task.TotalCodes,
			Status:      task.Status,
			FilePath:    filePath,
		})
	}

	data := map[string]interface{}{
		"Title": "Список заданий",
		"Tasks": tasks,
	}

	s.render(w, "tasks", data)
}

// handleTasksActions обрабатывает действия с заданиями
func (s *Server) handleTasksActions(w http.ResponseWriter, r *http.Request) {
	// Извлекаем ID задания из URL
	path := strings.TrimPrefix(r.URL.Path, "/tasks/")
	parts := strings.Split(path, "/")

	if len(parts) < 1 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}

	taskID := parts[0]

	// Находим файл по ID задания - сканируем директорию каждый раз
	files, err := s.dirScanner.ListMarkFiles()
	if err != nil {
		http.Error(w, "Ошибка при получении списка файлов", http.StatusInternalServerError)
		return
	}

	var filePath string
	for _, file := range files {
		// Парсим файл для получения ID
		tempTask, _, err := filemanager.ParseMarkFile(file)
		if err != nil {
			continue
		}

		if tempTask.ID == taskID {
			filePath = file
			break
		}
	}

	if filePath == "" {
		http.NotFound(w, r)
		return
	}

	// Парсим файл для получения задания и кодов
	task, codes, err := filemanager.ParseMarkFile(filePath)
	if err != nil {
		http.Error(w, "Ошибка при парсинге файла", http.StatusInternalServerError)
		return
	}

	// Определяем действие в зависимости от URL
	if len(parts) == 1 {
		// Просмотр деталей задания
		s.handleTaskDetails(w, r, task, codes)
	} else if len(parts) == 2 && parts[1] == "start" {
		// Начало обработки задания
		s.handleTaskStart(w, r, task, filePath)
	} else {
		http.NotFound(w, r)
	}
}

// handleTaskDetails обрабатывает просмотр деталей задания
func (s *Server) handleTaskDetails(w http.ResponseWriter, r *http.Request, task *models.Task, codes []*models.Code) {
	// Подготавливаем данные для отображения
	taskVM := TaskViewModel{
		ID:          task.ID,
		GTIN:        task.GTIN,
		Date:        task.Date,
		BatchNumber: task.BatchNumber,
		TotalCodes:  task.TotalCodes,
		Status:      task.Status,
		FilePath:    task.FilePath,
	}

	// Для предпросмотра берем только первые 10 кодов
	maxPreview := 10
	if len(codes) < maxPreview {
		maxPreview = len(codes)
	}
	previewCodes := codes[:maxPreview]

	data := map[string]interface{}{
		"Title":        "Детали задания",
		"Task":         taskVM,
		"Codes":        codes,
		"PreviewCodes": previewCodes,
	}

	s.render(w, "task_details", data)
}

// render рендерит шаблон с данными
func (s *Server) render(w http.ResponseWriter, content string, data map[string]interface{}) {
	// Указываем, какой контент-шаблон использовать
	data["ContentTemplate"] = content

	// Всегда рендерим layout.html, который включает нужный контент-шаблон
	err := s.templates.ExecuteTemplate(w, "layout.html", data)
	if err != nil {
		log.Printf("Ошибка при рендеринге шаблона: %v", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
	}
}

// Обновленная функция handleTaskStart для перемещения файла сразу в архив
func (s *Server) handleTaskStart(w http.ResponseWriter, r *http.Request, task *models.Task, filePath string) {
	// Перемещаем файл из директории входящих сразу в директорию архива
	archivePath, err := filemanager.MoveToArchive(filePath, s.config.ArchiveDir)
	if err != nil {
		http.Error(w, "Ошибка при архивации файла: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Обновляем путь к файлу
	task.FilePath = archivePath
	task.Status = models.TaskStatusProcessing
	task.ProcessedAt = time.Now()

	// Сохраняем задание в БД
	if err := s.db.SaveTask(task); err != nil {
		http.Error(w, "Ошибка при сохранении задания: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Получаем коды из файла и сохраняем в БД
	_, codes, err := filemanager.ParseMarkFile(archivePath)
	if err != nil {
		http.Error(w, "Ошибка при парсинге файла: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := s.db.SaveCodes(codes); err != nil {
		http.Error(w, "Ошибка при сохранении кодов: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Устанавливаем это задание как активное
	if err := s.db.SetActiveTask(task.ID); err != nil {
		log.Printf("Ошибка при установке активного задания: %v", err)
	}

	log.Printf("Задание %s начато, файл архивирован: %s -> %s", task.ID, filePath, archivePath)

	// Перенаправляем на страницу активного задания
	http.Redirect(w, r, "/active-task", http.StatusSeeOther)
}

// handleActiveTask обрабатывает страницу с активным заданием (обновление)
func (s *Server) handleActiveTask(w http.ResponseWriter, r *http.Request) {
	// Получаем ID активного задания
	activeTaskID, err := s.db.GetActiveTask()
	if err != nil || activeTaskID == "" {
		// Если запрашивается JSON-формат, возвращаем ошибку в JSON
		if r.URL.Query().Get("format") == "json" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"error": "Нет активного задания"}`))
			return
		}

		// Иначе перенаправляем на список заданий
		http.Redirect(w, r, "/tasks", http.StatusSeeOther)
		return
	}

	// Получаем информацию о задании из БД
	task, err := s.db.GetTaskByID(activeTaskID)
	if err != nil {
		if r.URL.Query().Get("format") == "json" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"error": "Ошибка при получении задания"}`))
			return
		}

		http.Error(w, "Ошибка при получении задания: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Получаем коды для этого задания
	codes, err := s.db.GetCodesByTaskID(activeTaskID)
	if err != nil {
		if r.URL.Query().Get("format") == "json" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"error": "Ошибка при получении кодов"}`))
			return
		}

		http.Error(w, "Ошибка при получении кодов: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Считаем статистику
	totalCodes := len(codes)
	scannedCodes := 0
	validCodes := 0

	for _, code := range codes {
		if code.Scanned {
			scannedCodes++
			if code.Valid {
				validCodes++
			}
		}
	}

	progress := 0
	if totalCodes > 0 {
		progress = int(float64(scannedCodes) / float64(totalCodes) * 100)
	}

	// Обновляем статистику в БД
	s.db.UpdateTaskStatistics(activeTaskID)

	// Если запрашивается JSON-формат, возвращаем данные в JSON
	if r.URL.Query().Get("format") == "json" {
		w.Header().Set("Content-Type", "application/json")
		jsonData := map[string]interface{}{
			"task_id":  task.ID,
			"total":    totalCodes,
			"scanned":  scannedCodes,
			"valid":    validCodes,
			"progress": progress,
		}

		jsonBytes, _ := json.Marshal(jsonData)
		w.Write(jsonBytes)
		return
	}

	// Иначе рендерим HTML-шаблон
	data := map[string]interface{}{
		"Title":        "Активное задание",
		"Task":         task,
		"TotalCodes":   totalCodes,
		"ScannedCodes": scannedCodes,
		"ValidCodes":   validCodes,
		"Progress":     progress,
	}

	s.render(w, "active_task", data)
}

// handleScanCode обрабатывает сканирование кода
func (s *Server) handleScanCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Получаем данные формы
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ошибка при парсинге формы", http.StatusBadRequest)
		return
	}

	codeValue := r.FormValue("code")
	if codeValue == "" {
		http.Error(w, "Код не может быть пустым", http.StatusBadRequest)
		return
	}

	// Получаем ID активного задания
	activeTaskID, err := s.db.GetActiveTask()
	if err != nil || activeTaskID == "" {
		http.Error(w, "Нет активного задания", http.StatusBadRequest)
		return
	}

	// Проверяем, существует ли код для этого задания и не отсканирован ли он уже
	codes, err := s.db.GetCodesByTaskID(activeTaskID)
	if err != nil {
		http.Error(w, "Ошибка при получении кодов: "+err.Error(), http.StatusInternalServerError)
		return
	}

	codeFound := false
	codeAlreadyScanned := false
	for _, code := range codes {
		if code.Value == codeValue {
			codeFound = true
			if code.Scanned {
				codeAlreadyScanned = true
			}
			break
		}
	}

	if !codeFound {
		// Код не найден в задании
		http.Error(w, "Код не найден в текущем задании", http.StatusBadRequest)
		return
	}

	if codeAlreadyScanned {
		// Код уже отсканирован
		http.Error(w, "Код уже отсканирован", http.StatusBadRequest)
		return
	}

	// Обновляем статус кода
	if err := s.db.UpdateCodeStatus(activeTaskID, codeValue, true, true, ""); err != nil {
		http.Error(w, "Ошибка при обновлении статуса кода: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Обновляем статистику задания
	if err := s.db.UpdateTaskStatistics(activeTaskID); err != nil {
		log.Printf("Ошибка при обновлении статистики: %v", err)
	}

	// Возвращаем JSON с результатом
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success": true, "message": "Код успешно отсканирован"}`))
}

// Обновленная функция handleCompleteTask
func (s *Server) handleCompleteTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Получаем ID активного задания
	activeTaskID, err := s.db.GetActiveTask()
	if err != nil || activeTaskID == "" {
		http.Error(w, "Нет активного задания", http.StatusBadRequest)
		return
	}

	// Получаем информацию о задании
	task, err := s.db.GetTaskByID(activeTaskID)
	if err != nil {
		http.Error(w, "Ошибка при получении задания: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Получаем отсканированные коды
	scannedCodes, err := s.db.GetScannedCodesByTaskID(activeTaskID)
	if err != nil {
		http.Error(w, "Ошибка при получении кодов: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Генерируем файл с результатами
	resultFilePath, err := filemanager.GenerateSerializationFile(task, scannedCodes, s.config.OutgoingDir)
	if err != nil {
		http.Error(w, "Ошибка при создании файла результатов: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Обновляем статус задания
	if err := s.db.UpdateTaskStatus(activeTaskID, models.TaskStatusCompleted); err != nil {
		http.Error(w, "Ошибка при обновлении статуса задания: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Удаляем активное задание
	if err := s.db.ClearActiveTask(); err != nil {
		log.Printf("Ошибка при удалении активного задания: %v", err)
	}

	log.Printf("Задание %s завершено, создан файл результатов: %s", activeTaskID, resultFilePath)

	// Перенаправляем на список заданий
	http.Redirect(w, r, "/tasks", http.StatusSeeOther)
}
