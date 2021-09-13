# Сборщик балансов и не только

Для работы требуется браузер Google Chrome!

Сервис собирает по расписанию или запросу балансы телефонов и прочую подобную информацию. Хранит историю. Позволяет просматривать информацию через web интерфейс (по умолчанию [http://localhost:10802/](http://localhost:10802/)).

Под linux (ubuntu) работает сервисом. Скрипты запуска в [cmd/linux/](cmd/linux/).

Под форточками не собирал, но должно работать. Стандартный скрипт сборки приложен. Тоже должен работать как сервис. Скрипты запуска в [cmd/windows/](cmd/windows/).

Перед запуском необходимо создать файл _config/entities.toml_ с описанием того, что требуется собирать. Пример в [config/entities-sample.toml](config/entities-sample.toml).

Общие настройки в файле [config/balance-collector.toml](config/balance-collector.toml).

Поддерживаемые операторы в [config/operators/](config/operators/).

К сожалению, почти ни у кого из операторов нет API для получения данных о счёте. Из представленных здесь он имеется только у 1vds. Поэтому приходится получать информацию имитируя вход в личный кабинет и распарсивая полученные страницы. Не секрет, что большинство операторов любят менять свои личные кабинеты, кто-то реже, кто-то чаще. Из-за этого перестает работать соответствующий скрипт. Возможно, это уже случилось с частью из представленных.

Я актуализирую только тех, которые имеются в моем текущем личном наборе. Если у вас какой-то из операторов перестал работать и исправление не пришло в течение недели, значит я им уже не пользуюсь. Попробуйте скорректировать соответствующий скрипт самостоятельно. Как правило, там нет ничего сложного. Посмотрите в [config/operators/](config/operators/). После внесения изменений необходимо перезапустить сервис. Если он не запускается, то обычно причину можно увидеть в появившемся в рабочей директории файле _balance-collector_unsaved.log_. Исправьте ошибку и попробуйте запустить сервис снова.

Аналогично вы можете написать скрипт и для почти любого другого оператора, отсутствующего в имеющемся наборе.
