# Books API

[![CI](https://github.com/LucasLevingston/api-books-golang/actions/workflows/ci.yml/badge.svg)](https://github.com/LucasLevingston/api-books-golang/actions/workflows/ci.yml)

REST API em Go para gerenciamento de livros, com banco de dados PostgreSQL, containerizada com Docker.

## Stack

| Camada        | Tecnologia                             |
|---------------|----------------------------------------|
| Linguagem     | Go 1.22                                |
| Router        | [chi v5](https://github.com/go-chi/chi)|
| Banco de dados| PostgreSQL 16                          |
| SQL           | [sqlx](https://github.com/jmoiron/sqlx) + `lib/pq` |
| Migrações     | [golang-migrate](https://github.com/golang-migrate/migrate) |
| Validação     | [go-playground/validator v10](https://github.com/go-playground/validator) |
| Documentação  | [swaggo/swag](https://github.com/swaggo/swag) + [Scalar](https://scalar.com) |
| Testes        | `testing` + [testify](https://github.com/stretchr/testify) |
| Config        | [godotenv](https://github.com/joho/godotenv) |

## Arquitetura

```
cmd/api/                          # Entrypoint: wiring de dependências
docs/                             # Spec OpenAPI gerada pelo swag (não editar)
internal/
  book/                           # Feature: domínio de livros
    domain.go                     # Entidade, interfaces (ports)
    service.go                    # Regras de negócio (application layer)
    repository.go                 # Adapter PostgreSQL (infrastructure)
    handler.go                    # Adapter HTTP (infrastructure)
    service_test.go               # Testes unitários do service
    handler_test.go               # Testes unitários do handler
  platform/
    database/
      postgres.go                 # Conexão + migrations
      migrations/                 # Arquivos SQL (embedded)
    server/
      server.go                   # HTTP server com middlewares
```

O fluxo de dados segue o padrão hexagonal:  
`Handler → Service (regras) → Repository (DB)`

## Pré-requisitos

- [Docker](https://docs.docker.com/get-docker/) 24+
- [Docker Compose](https://docs.docker.com/compose/) v2+ (incluído no Docker Desktop)

Para desenvolvimento local (testes sem Docker):
- [Go](https://go.dev/dl/) 1.22+

## Rodando o projeto

### 1. Clone / acesse o diretório

```bash
cd api-books-golang
```

### 2. Suba os containers

```bash
docker compose up --build
```

O Docker irá:
1. Subir o PostgreSQL 16 com health check
2. Instalar o `swag` CLI e **regenerar a documentação OpenAPI** a partir das anotações do código
3. Compilar a API Go (multi-stage build)
4. Executar as migrações automaticamente (criação de tabela + seed de ~100 livros)
5. Iniciar o servidor na porta **8080**

### 3. Verifique que está funcionando

```bash
curl.exe http://localhost:8080/health
# HTTP 200
```

### 4. Acesse a documentação interativa

Abra no browser:

```
http://localhost:8080/docs
```

A interface [Scalar](https://scalar.com) é carregada automaticamente com todos os endpoints, parâmetros e exemplos de resposta.

O spec OpenAPI bruto (JSON) também está disponível em:

```
http://localhost:8080/swagger/doc.json
```

### Parar os containers

```bash
docker compose down
```

Para remover também o volume do banco:

```bash
docker compose down -v
```

---

## Rodando testes

### Com Go instalado localmente

```powershell
go test ./...
```

#### Cobertura das rotas (handler)

Exibe a cobertura por função apenas do `handler.go`, ignorando pacotes sem testes:

```powershell
go test ./internal/book/... -coverprofile=coverage.out
go tool cover -func=coverage.out | Select-String "handler.go"
```

Saída esperada:

```
internal/book/handler.go:20:    NewHandler       100.0%
internal/book/handler.go:24:    RegisterRoutes   100.0%
internal/book/handler.go:50:    list              81.0%
internal/book/handler.go:98:    create            83.3%
internal/book/handler.go:131:   update           100.0%
internal/book/handler.go:170:   delete            81.8%
...
```

Para abrir o relatório visual no browser:

```powershell
go test ./internal/book/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

> **Linux/macOS** — use `grep` no lugar de `Select-String`:
> ```bash
> go test ./internal/book/... -coverprofile=coverage.out
> go tool cover -func=coverage.out | grep "handler.go"
> ```
>
> Se tiver `make` instalado, os atalhos `make cover` e `make cover-html` já encapsulam esses comandos.

### Sem Go instalado (via Docker)

```powershell
docker run --rm -v "${PWD}:/app" -w /app golang:1.22-alpine go test ./...
```

> Linux/macOS:
> ```bash
> docker run --rm -v "$(pwd)":/app -w /app golang:1.22-alpine go test ./...
> ```

---

## API Reference

### Base URL

```
http://localhost:8080
```

### Endpoints

| Método   | Rota                  | Descrição                      |
|----------|-----------------------|--------------------------------|
| `GET`    | `/books`              | Listar com filtros             |
| `POST`   | `/books`              | Criar livro                    |
| `PUT`    | `/books/{id}`         | Atualizar livro                |
| `DELETE` | `/books/{id}`         | Deletar livro                  |
| `GET`    | `/health`             | Health check                   |
| `GET`    | `/docs`               | Documentação interativa Scalar |
| `GET`    | `/swagger/doc.json`   | Spec OpenAPI 2.0 (JSON)        |

---

### `GET /books` — Listar livros

#### Query parameters

| Parâmetro     | Tipo    | Default | Descrição                                          |
|---------------|---------|---------|----------------------------------------------------|
| `page`        | int     | `1`     | Número da página                                   |
| `page_size`   | int     | `10`    | Itens por página (máx 100)                         |
| `author`      | string  | —       | Filtro parcial por autor (case-insensitive)        |
| `year`        | int     | —       | Filtro exato por ano                               |
| `masterpiece` | bool    | —       | Filtro por obra-prima (`true` / `false`)           |
| `sort_by`     | string  | `name`  | Campo de ordenação: `name`, `author`, `year`, `masterpiece` |
| `sort_dir`    | string  | `asc`   | Direção: `asc` ou `desc`                           |

#### Resposta `200`

```json
{
  "data": [
    {
      "id": 1,
      "name": "1984",
      "author": "George Orwell",
      "year": 1949,
      "masterpiece": false,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 100,
  "page": 1,
  "page_size": 10,
  "total_pages": 10
}
```

---

### `POST /books` — Criar livro

#### Body

```json
{
  "name": "O Senhor dos Anéis",
  "author": "J.R.R. Tolkien",
  "year": 1954,
  "masterpiece": true
}
```

| Campo         | Tipo    | Obrigatório | Regras             |
|---------------|---------|-------------|--------------------|
| `name`        | string  | sim         | 1–255 caracteres   |
| `author`      | string  | sim         | 1–255 caracteres   |
| `year`        | int     | sim         | 1–2100             |
| `masterpiece` | bool    | não         | default `false`    |

#### Resposta `201` — livro criado

---

### `PUT /books/{id}` — Atualizar livro

Todos os campos são opcionais (patch semântico).

#### Body

```json
{
  "masterpiece": true
}
```

#### Resposta `200` — livro atualizado  
#### Resposta `404` — livro não encontrado

---

### `DELETE /books/{id}` — Deletar livro

#### Resposta `204` — deletado  
#### Resposta `404` — livro não encontrado

---

## Exemplos com curl.exe

### Listar todos (paginação padrão)

```bash
curl.exe -s "http://localhost:8080/books"
```

### Listar página 2 com 5 itens

```bash
curl.exe -s "http://localhost:8080/books?page=2&page_size=5"
```

### Filtrar por autor

```bash
curl.exe -s "http://localhost:8080/books?author=Machado"
```

### Filtrar por ano

```bash
curl.exe -s "http://localhost:8080/books?year=1949"
```

### Filtrar obras-primas

```bash
curl.exe -s "http://localhost:8080/books?masterpiece=true"
```

### Ordenar por autor (descendente)

```bash
curl.exe -s "http://localhost:8080/books?sort_by=author&sort_dir=desc"
```

### Combinar filtros

```bash
curl.exe -s "http://localhost:8080/books?author=jorge&sort_by=year&sort_dir=asc&page_size=5"
```

### Criar livro

```bash
curl.exe -s -X POST http://localhost:8080/books \
  -H "Content-Type: application/json" \
  -d '{"name":"O Senhor dos Anéis","author":"J.R.R. Tolkien","year":1954,"masterpiece":true}'
```

### Atualizar livro (ID 1) — marcar como obra-prima

```bash
curl.exe -s -X PUT http://localhost:8080/books/1 \
  -H "Content-Type: application/json" \
  -d '{"masterpiece":true}'
```

### Atualizar múltiplos campos

```bash
curl.exe -s -X PUT http://localhost:8080/books/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Nineteen Eighty-Four","year":1949}'
```

### Deletar livro

```bash
curl.exe -s -X DELETE http://localhost:8080/books/1 -w "%{http_code}"
# 204
```

### Tentar buscar livro inexistente

```bash
curl.exe -s http://localhost:8080/books/9999 2>&1
# (DELETE/PUT retornam 404)
```

### Erro de validação (campo ausente)

```bash
curl.exe -s -X POST http://localhost:8080/books \
  -H "Content-Type: application/json" \
  -d '{"author":"Alguém","year":2020}'
# HTTP 422
```

```json
{
  "errors": [
    { "field": "name", "message": "Name is required" }
  ]
}
```

### Erro de validação (valor fora do range)

```bash
curl.exe -s -X PUT http://localhost:8080/books/1 \
  -H "Content-Type: application/json" \
  -d '{"year":9999}'
# HTTP 422
```

```json
{
  "errors": [
    { "field": "year", "message": "Year must be at most 2100" }
  ]
}
```

---

## Variáveis de ambiente

Copie `.env.example` para `.env` para rodar localmente sem Docker:

```bash
cp .env.example .env
```

| Variável      | Default      | Descrição              |
|---------------|--------------|------------------------|
| `DB_HOST`     | `localhost`  | Host do PostgreSQL     |
| `DB_PORT`     | `5432`       | Porta do PostgreSQL    |
| `DB_USER`     | `postgres`   | Usuário                |
| `DB_PASSWORD` | `postgres`   | Senha                  |
| `DB_NAME`     | `books`      | Nome do banco          |
| `DB_SSLMODE`  | `disable`    | SSL mode               |
| `PORT`        | `8080`       | Porta da API           |

---

## Documentação OpenAPI

A spec é gerada automaticamente a partir das anotações nos handlers pelo [swaggo/swag](https://github.com/swaggo/swag).

- **Durante o build Docker**: o `swag init` é executado antes da compilação — docs sempre em sincronia com o código.
- **Localmente**: após instalar o CLI, rode:

```bash
go install github.com/swaggo/swag/cmd/swag@v1.16.6
swag init -g cmd/api/main.go -o docs
```

> Os arquivos em `docs/` são gerados — **não edite manualmente**.

---

## Migrações

As migrações são **embarcadas no binário** (`embed.FS`) e executadas automaticamente na inicialização.

| Migração | Descrição                          |
|----------|------------------------------------|
| `000001` | Criação da tabela `books` + índices|
| `000002` | Seed inicial com ~100 livros       |

Para reverter manualmente (requer `migrate` CLI):

```bash
migrate -path internal/platform/database/migrations \
        -database "postgres://books:books@localhost:5432/books?sslmode=disable" \
        down 1
```
