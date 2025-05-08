package scanning

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

const (
	connectTimeout = 5 * time.Second
	readTimeout    = 2 * time.Second
	writeTimeout   = 2 * time.Second
)

// TCPScanner представляет сканер, работающий по TCP
type TCPScanner struct {
	address     string
	scanCommand string
	conn        net.Conn
	reader      *bufio.Reader
}

// NewTCPScanner создает новый экземпляр TCP-сканера
func NewTCPScanner(address, scanCommand string) *TCPScanner {
	return &TCPScanner{
		address:     address,
		scanCommand: scanCommand,
	}
}

// Connect устанавливает соединение со сканером
func (s *TCPScanner) Connect() error {
	conn, err := net.DialTimeout("tcp", s.address, connectTimeout)
	if err != nil {
		return fmt.Errorf("ошибка подключения к сканеру: %w", err)
	}

	s.conn = conn
	s.reader = bufio.NewReader(conn)
	return nil
}

// Close закрывает соединение со сканером
func (s *TCPScanner) Close() error {
	if s.conn != nil {
		err := s.conn.Close()
		s.conn = nil
		s.reader = nil
		return err
	}
	return nil
}

// Scan выполняет сканирование и возвращает результат
func (s *TCPScanner) Scan() (string, error) {
	if s.conn == nil {
		return "", fmt.Errorf("соединение со сканером не установлено")
	}

	// Устанавливаем таймаут для записи
	if err := s.conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
		return "", fmt.Errorf("ошибка установки таймаута записи: %w", err)
	}

	// Отправляем команду сканирования
	if _, err := s.conn.Write([]byte(s.scanCommand)); err != nil {
		return "", fmt.Errorf("ошибка отправки команды: %w", err)
	}

	// Устанавливаем таймаут для чтения
	if err := s.conn.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
		return "", fmt.Errorf("ошибка установки таймаута чтения: %w", err)
	}

	// Читаем ответ
	response, err := s.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	// Убираем символы конца строки
	response = strings.TrimSpace(response)

	return response, nil
}
