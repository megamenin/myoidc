# Пример реализации клиента Open Id Connect

## Запуск

### Запуск локально

Для запуска потребуется
- golang 1.21+

```shell
go run cmd/app/main.go
```

### Запуск в docker контейнере

Для запуска потребуется:
- docker 20.10+

```shell
docker-compose up
```

## Использование

http://localhost:8080/oauth/login - перенаправляет на авторизацию в провайдер

http://localhost:8080/oauth/callback - принимает код от провайдера, авторизует пользователя в приложении и создает пользовательскую сессию

http://localhost:8080/oauth/userinfo - возвращает информацию об авторизованном пользователе

http://localhost:8080/ - перенаправляет пользователя на логин или главную страницу в зависимости от наличия сессии

> В настройках приложения указан тестовый OpenId Connect сервер, созданный с помощью [auth0.com](https://auth0.com/)