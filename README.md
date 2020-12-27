# Сборщик балансов и не только

Сервис собирает по расписанию или запросу балансы телефонов и прочую подобную информацию. Хранит историю. Позволяет просматривать информацию через web интерфейс.

Под linux (ubuntu) работает сервисом. Скрипты запуска в [cmd/linux/](cmd/linux/).

Под форточками не собирал, но должно работать. Стандартный скрипт сборки приложен. Тоже должен работать как сервис. Скрипты запуска в [cmd/windows/](cmd/windows/).

Перед запуском необходимо создать файл _config/entities.toml_ с описанием того, что требуется собирать. Пример в [config/entities-sample.toml](config/entities-sample.toml).

Поддерживаемые операторы в [config/operators/](config/operators/). Можно создавать своих по аналогии.

Общие настройки в файле [config/balance-collector.toml](config/balance-collector.toml).