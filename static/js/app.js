// static/js/app.js (обновление)
// JavaScript для FileMarker
console.log('FileMarker - приложение загружено');

// Глобальная функция для обновления статуса активного задания
function updateActiveTaskStatus() {
    // Если мы на странице активного задания
    if (window.location.pathname === '/active-task') {
        const progressBar = document.querySelector('.progress');
        const progressInfo = document.querySelector('.progress-info');

        if (progressBar && progressInfo) {
            // Запрашиваем обновление данных каждые 5 секунд
            setTimeout(() => {
                fetch('/active-task?format=json')
                    .then(response => response.json())
                    .then(data => {
                        progressBar.style.width = data.progress + '%';
                        progressInfo.innerHTML = `
                            <p>Отсканировано: <strong>${data.scanned}</strong> из <strong>${data.total}</strong> (${data.progress}%)</p>
                            <p>Валидных кодов: <strong>${data.valid}</strong></p>
                        `;
                    })
                    .catch(error => console.error('Ошибка обновления статуса:', error))
                    .finally(() => {
                        // Запускаем следующее обновление
                        updateActiveTaskStatus();
                    });
            }, 5000);
        }
    }
}

// Инициализируем обновление статуса при загрузке страницы
document.addEventListener('DOMContentLoaded', function() {
    updateActiveTaskStatus();

    // Обработка фокуса в поле ввода кода
    const codeInput = document.getElementById('code-input');
    if (codeInput) {
        // Автоматический фокус при клике на странице
        document.addEventListener('click', function() {
            codeInput.focus();
        });

        // Подавление стандартного поведения Enter
        document.addEventListener('keydown', function(e) {
            if (e.key === 'Enter' && document.activeElement !== codeInput) {
                e.preventDefault();
                codeInput.focus();
            }
        });
    }
});