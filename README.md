# Log Processor

## Описание
Инструмент для анализа логов распределённых систем

## Возможности
- Сопоставление записей по request_id
- Обнаружение ошибок и построение хронологии
- Конкурентная обработка
- JSON-отчёты

## Установка и запуск
```bash
git clone https://github.com/ichuvilin/log-handler

cd log-handler

go build -o log-processor main.go
./log-processor --input-dir /var/log/services/
```