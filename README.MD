# 📚 Library API

API REST para gerenciamento de uma biblioteca, desenvolvida como Trabalho de Conclusão de Curso (TCC).

A aplicação permite o cadastro de **usuários** e **livros**, a **consulta** do acervo por título ou autor, e o controle de **empréstimos** com validação de disponibilidade de exemplares.

## Stack Tecnológica

| Componente | Tecnologia |
|---|---|
| Linguagem | Go |
| Router HTTP | [Chi](https://github.com/go-chi/chi) |
| Banco de dados | PostgreSQL 16 |
| Driver SQL | [pgx](https://github.com/jackc/pgx) + [sqlx](https://github.com/jmoiron/sqlx) |
| Migrations | SQL puro |
| Container | Docker Compose |

---

## System Design

### Visão Geral da Arquitetura

```
┌─────────────┐       HTTP/JSON        ┌──────────────────────────────────────┐
│             │ ◄────────────────────►  │            API Server (Go)           │
│   Cliente   │   POST/GET/PATCH       │                                      │
│  (Postman,  │                        │  ┌────────┐  ┌──────┐  ┌──────────┐ │
│   curl,     │                        │  │Router  │─►│Handle│─►│Repository│ │
│   Frontend) │                        │  │ (Chi)  │  │ rs   │  │          │ │
│             │                        │  └────────┘  └──────┘  └─────┬────┘ │
└─────────────┘                        │                              │      │
                                       └──────────────────────────────┼──────┘
                                                                      │
                                                                      │ SQL
                                                                      ▼
                                                               ┌─────────────┐
                                                               │ PostgreSQL  │
                                                               │   (Docker)  │
                                                               └─────────────┘
```

### Fluxo de Requisição

```mermaid
sequenceDiagram
    participant C as Cliente
    participant R as Router (Chi)
    participant H as Handler
    participant Repo as Repository
    participant DB as PostgreSQL

    C->>R: HTTP Request
    R->>H: Route match + middlewares (Logger, Recoverer)
    H->>H: Validação do JSON / parâmetros
    H->>Repo: Chamada ao repositório
    Repo->>DB: Query SQL (sqlx)
    DB-->>Repo: Resultado
    Repo-->>H: Struct Go
    H-->>C: JSON Response (200/201/4xx)
```

### Diagrama Entidade-Relacionamento

```mermaid
erDiagram
    USERS {
        int id PK
        varchar name
        varchar email UK
        timestamp created_at
    }

    BOOKS {
        int id PK
        varchar title
        varchar author
        varchar isbn UK
        int quantity
        timestamp created_at
    }

    LOANS {
        int id PK
        int user_id FK
        int book_id FK
        timestamp loaned_at
        timestamp returned_at
        varchar status
    }

    USERS ||--o{ LOANS : "realiza"
    BOOKS ||--o{ LOANS : "é emprestado"
```

### Fluxo de Empréstimo

```mermaid
flowchart LR
    A[POST /api/loans] --> B{user_id e book_id válidos?}
    B -- Não --> C[400 Bad Request]
    B -- Sim --> D[Buscar livro no banco]
    D --> E{Livro existe?}
    E -- Não --> F[404 Not Found]
    E -- Sim --> G[Contar empréstimos ativos do livro]
    G --> H{empréstimos ativos < quantity?}
    H -- Não --> I[409 Conflict - Sem exemplares]
    H -- Sim --> J[Criar empréstimo com status 'active']
    J --> K[201 Created]
```

---

## Endpoints da API

### Usuários

| Método | Rota | Descrição |
|---|---|---|
| `POST` | `/api/users` | Criar usuário |
| `GET` | `/api/users` | Listar todos os usuários |
| `GET` | `/api/users/{id}` | Buscar usuário por ID |

### Livros

| Método | Rota | Descrição |
|---|---|---|
| `POST` | `/api/books` | Cadastrar livro |
| `GET` | `/api/books` | Listar livros (busca com `?q=termo`) |
| `GET` | `/api/books/{id}` | Buscar livro por ID |

### Empréstimos

| Método | Rota | Descrição |
|---|---|---|
| `POST` | `/api/loans` | Registrar empréstimo |
| `PATCH` | `/api/loans/{id}/return` | Devolver livro |
| `GET` | `/api/loans` | Listar empréstimos (filtro com `?user_id=X`) |

---

## Como Rodar

```bash
# 1. Subir o PostgreSQL
docker compose up -d

# 2. Rodar a API (migrations executam automaticamente)
go run ./cmd/server
```

A API estará disponível em `http://localhost:8080/api`

---

## Estrutura do Projeto

```
├── cmd/
│   └── server/
│       └── main.go              # Entrypoint
├── internal/
│   ├── config/
│   │   └── config.go            # Carrega variáveis de ambiente
│   ├── database/
│   │   ├── connection.go        # Conexão com PostgreSQL
│   │   └── migrations.go        # Executa migrations
│   ├── models/
│   │   └── models.go            # Structs: User, Book, Loan
│   ├── handlers/
│   │   ├── user_handler.go      # Handlers de usuário
│   │   ├── book_handler.go      # Handlers de livro
│   │   └── loan_handler.go      # Handlers de empréstimo
│   ├── repository/
│   │   ├── user_repository.go   # CRUD de usuários
│   │   ├── book_repository.go   # CRUD de livros
│   │   └── loan_repository.go   # CRUD de empréstimos
│   └── router/
│       └── router.go            # Define rotas com chi
├── migrations/
│   └── 001_init.sql             # Schema inicial
├── .env                         # Variáveis de ambiente
├── docker-compose.yml           # PostgreSQL
├── go.mod
└── go.sum
```