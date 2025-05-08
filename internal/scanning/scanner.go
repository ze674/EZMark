package scanning

// Scanner определяет интерфейс для устройств сканирования
type Scanner interface {
	Connect() error
	Close() error
	Scan() (string, error)
}
