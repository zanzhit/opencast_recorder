# Сервис записи и хранения видео с камер видеонаблюдения

# Общие вводные
Для записи видео с камер, сервис использует утилиту gstreamer. Сервис способен записывать в двух режимах: одиночный (идет обыкновенная запись с камеры) и смешанный (на экране появляются две картинки, аудио берется из первого потока).
В качестве сервиса для хранения и компоновки видеофайлов используется Opencast.

# Getting Started

- Иметь устновленный sqlite.
- Подключиться к сети камер.
- Развернуть Opencast.
- Изменить файл config/config.go под свои параметры.

# Usage

Запустить сервис можно с помощью команды go cmd/main.go


## Examples

Некоторые примеры запросов через Postman
- [Добавление камеры](#create-camera)
- [Начало одиночной записи](#single-start)
- [Начало смешанной записи](#mixed-start)
- [Остановка одиночной записи](#single-stop)
- [Остановка смешанной записи](#mixed-stop)
- [Запланированная запись](#schedule)
- [Информация о последней записи с камеры](#stats)

### Добавление камеры <a name="create-camera"></a>

Добавление камеры в базу данных:
```curl
POST http://localhost:8000/cameras
```

Body:
```json
{
	"camera_ip": "192.168.1.2:554",
	"room_number": "101",
	"has_audio": true
}
```

Пример ответа:
200

### Начало одиночной записи <a name="single-start"></a>

Начало обычной одиночной записи:
```curl
POST http://localhost:8000/cameras/192.168.1.2:554/start
```

Пример ответа:
200

### Начало смешанной записи <a name="mixed-start"></a>

Начало смешанной записи:
```curl
POST http://localhost:8000/cameras/192.168.1.2:554,192.168.1.3:554/start
```

Пример ответа:
200

### Остановка одиночной записи <a name="single-stop"></a>

Остановка одиночной записи:
```curl
POST http://localhost:8000/cameras/192.168.1.2:554/stop
```

Пример ответа:
200

### Остановка смешанной записи <a name="mixed-stop"></a>
(P.S на самом деле берется лишь второй ip адрес, поэтому смешанную запись можно остановить и одиночной остановкой, указав ip камеры, которая была второй в url)
Остановка смешанной записи:
```curl
POST http://localhost:8000/cameras/192.168.1.2:554,192.168.1.3:554/stop
```

Пример ответа:
200

### Запланированная запись <a name="schedule"></a>

Запланированная запись с указанием времени и продолжнительности (поддерживается как одиночная, так и смешанная запись):
```curl
POST http://localhost:8000/cameras/192.168.1.2:554/schedule
```
Body:
```json
{
	"start_time": "2024-05-22T15:00:00+03:00",
	"duration": "00:00:30"
}
В КОНЦЕ УКАЗЫВАЕТСЯ +3:00 КАК МОСКОВСКИЙ ЧАСОВОЙ ПОЯС

```
Пример ответа:
200

### Информация о последней записи с камеры <a name="stats"></a>

Получение баннеров по указанным тегу и/или фиче:
```curl
GET http://localhost:8000/cameras/192.168.1.2:554
```

Пример ответа:
200
```json
{
    "camera_ip": "192.168.1.2",
    "start_time": "2024-05-22T15:00:00Z",
    "stop_time": "2024-05-22T16:00:00Z",
    "file_path": "videos/192.168.1.2..554.mkv",
    "is_moved": true
}
```
