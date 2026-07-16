# Сборка и тесты

Использовать `task`, а не прямые вызовы `go`.

## Команды

```bash
# Линтер
task lint                           # golangci-lint run --config=.golangci.pipeline.yaml
task install:lint                   # установка golangci-lint v2

# Unit-тесты
task test:unit                      # все unit-тесты с покрытием

# Smoke-тесты (build tag: smoke)
task test:smoke                     # smoke-тесты с покрытием

# Concurrent-тесты (build tag: concurrent, race detector)
task test:concurrent                # тесты многопоточности

# Functional-тесты (build tag: functional)
task test:functional                # functional-тесты с покрытием

# Все тесты (unit + concurrent + smoke + functional)
task test                       # полный прогон всех тестов

# Покрытие
task test:coverage                  # показать сохранённое покрытие
task test:cover:html                # HTML-отчёт
task test:clean:coverage            # удалить файлы покрытия

# Моки (minimock)
task test:gen:mocks                 # go generate для моков

# Генерация кода
task gen                            # полная генерация
task gen:go                         # go generate ./...

# PostgreSQL миграции (goose)
task test:db:create                 # создание тестовой БД
task test:db:drop                   # удаление тестовой БД
task pg:up                          # применить миграции к основной БД
task pg:down                        # откатить последнюю миграцию
```

## Стандартная проверка

```bash
go vet ./... && go build ./... && go test ./... && task lint
```
