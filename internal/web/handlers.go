package web

import (
	"FileMarker/internal/config"
	"FileMarker/internal/filemanager"
	"FileMarker/internal/models"
	"html/template"
	"log"
	"net/http"
	"strings"
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
}

// NewServer создает новый экземпляр сервера
func NewServer(cfg config.Config) (*Server, error) {
	// Загружаем шаблоны
	tmpl, err := template.ParseGlob("templates/*.html")
	if err != nil {
		return nil, err
	}

	// Создаем сканер директории
	scanner := filemanager.NewDirectoryScanner(cfg.IncomingDir, cfg.ProcessingDir)

	return &Server{
		config:     cfg,
		templates:  tmpl,
		dirScanner: scanner,
	}, nil
}

// Start запускает веб-сервер
func (s *Server) Start() error {
	// Настраиваем обработчики маршрутов
	http.HandleFunc("/", s.handleHome)
	http.HandleFunc("/tasks", s.handleTasksList)
	http.HandleFunc("/tasks/", s.handleTasksActions)

	// Статические файлы
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Запускаем HTTP-сервер
	log.Printf("Запуск сервера на порту %s", s.config.ServerPort)
	return http.ListenAndServe(":"+s.config.ServerPort, nil)
}

// handleHome обрабатывает запрос на главную страницу
func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	data := map[string]interface{}{
		"Title": "Главная",
	}

	s.render(w, "layout.html", data)
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

	s.render(w, "layout.html", data)
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

	s.render(w, "layout.html", data)
}

// handleTaskStart обрабатывает начало обработки задания
func (s *Server) handleTaskStart(w http.ResponseWriter, r *http.Request, task *models.Task, filePath string) {
	// Перемещаем файл из директории входящих в директорию обработки
	newPath, err := filemanager.MoveToProcessing(filePath, s.config.ProcessingDir)
	if err != nil {
		http.Error(w, "Ошибка при перемещении файла: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Файл перемещен для обработки: %s -> %s", filePath, newPath)

	// Пока что просто перенаправляем обратно на список заданий
	// В дальнейшем здесь будет логика начала обработки
	http.Redirect(w, r, "/tasks", http.StatusSeeOther)
}

// render рендерит шаблон с данными
func (s *Server) render(w http.ResponseWriter, tmpl string, data map[string]interface{}) {
	err := s.templates.ExecuteTemplate(w, tmpl, data)
	if err != nil {
		log.Printf("Ошибка при рендеринге шаблона: %v", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
	}
}
