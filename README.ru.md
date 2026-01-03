# Instafix

**Instafix** готовит изображения к публикации в Instagram: подбирает формат, добавляет фон, паддинги, рамки и вотермарк. Основная логика реализована как Go‑пакет, сверху есть CLI и web‑service.

## Возможности

- Ресайз под форматы Instagram (square, portrait, landscape, story).
- Автовыбор формата по соотношению сторон.
- Фоны: solid, blur, stretch, average.
- Паддинги и рамки.
- Стиль вотермарка (текст передается при запуске).
- Поддержка DNG/RAW через встроенный JPEG preview.
- Профили обработки в конфиге.

## Структура проекта

- `config/` схема конфига и дефолтный `profiles.toml`.
- `cmd/cli/` CLI.
- `cmd/service/` HTTP сервис.
- `docs/` техническое описание и интерфейсы.
- `assets/` шрифты для вотермарка.

### Конфиг

Дефолтный конфиг — `config/profiles.toml`. Если `--config` не передан, поиск идет в таком порядке:

1. `INSTAFIX_CONFIG` (если задан)
2. `./profiles.toml`
3. `./config/profiles.toml`
4. `profiles.toml` рядом с исполняемым файлом

Текст вотермарка не хранится в конфиге. Его нужно передавать явно в CLI или HTTP‑запросе. Если текст не передан, вотермарк не рисуется.

## CLI

**Сборка:**

```shell
go build -o instafix ./cmd/cli
```

**Примеры:**

```shell
./instafix --profile default --watermark "@name" input.jpg
./instafix --config config/profiles.toml --profile white_passepartout --out output.jpg input.jpg
```

## Web‑service

**Сборка:**

```shell
go build -o instafix-server ./cmd/service
```

**Запуск:**

```shell
API_KEY=secret ./instafix-server --config config/profiles.toml --addr :8080
```

**Railway:**

- Используется `PORT` (Railway подставляет сам), если `--addr` не задан.
- Опциональные переменные: `API_KEY`, `INSTAFIX_CONFIG`.

**API:**

- `POST /fix` (multipart form‑поле `image`)
- Query params:
  - `profile` (по умолчанию `default`)
  - `watermark` (опционально)
- Header:
  - `X-API-Key` (обязателен, если задан `API_KEY`)

**Пример:**

```shell
curl -X POST \
  -H "X-API-Key: secret" \
  -F "image=@testdata/1.jpg" \
  "http://localhost:8080/fix?profile=black&watermark=@name"
```

## Документация

- Техническое описание: `docs/technical.md`
- Proposal по формату конфига: `docs/profiles.md`

## Тесты

```shell
go test ./...
```
