[![CI (Go tests + RPC test)](https://github.com/KenueYy/wallet-api-test-work/actions/workflows/main.yml/badge.svg?event=push)](https://github.com/KenueYy/wallet-api-test-work/actions/workflows/main.yml)

Можно запукать тесты из ide(но нужно сначала запустить контейнер с тестовой db "docker compose up -d postgres_test") или запускать используя "make test-go"

Можно прогнать тест на нагрузку с помощью k6 "make test-rpc"