# 📚 Microservices Library API

API REST para gerenciamento de uma biblioteca, desenvolvida como Trabalho de Conclusão de Curso (TCC).
Esta versão implementa a **Arquitetura de Microsserviços**, onde os domínios (Gateway, Usuários, Livros e Empréstimos) são separados em projetos distintos, rodando de forma isolada, comunicando-se via HTTP e possuindo bancos de dados próprios independentes (pattern *Database-per-service*).

## Stack Tecnológica

| Componente | Tecnologia |
|---|---|
| Linguagem | Go |
| Router HTTP | [Chi](https://github.com/go-chi/chi) |
| Banco de dados | 3 instâncias de PostgreSQL 16 |
| Driver SQL | [pgx](https://github.com/jackc/pgx) + [sqlx](https://github.com/jmoiron/sqlx) |
| Comunicação Inter-serviços | Chamadas HTTP Síncronas |
| Container | Docker Compose |

---

## System Design

Na arquitetura de Microsserviços, o Cliente faz chamadas para o **API Gateway**, que atua como proxy reverso para os microsserviços do backend, escondendo a complexidade da rede. O serviço de empréstimos, antes de confirmar uma inserção, valida com o **User Service** se o usuário existe, e com o **Book Service** se o livro existe e tem estoque disponível, demonstrando a comunicação síncrona inter-serviços.

### Visão Geral da Arquitetura

```mermaid
flowchart TD
    Client([Cliente HTTP / Postman]) --> Gateway

    subgraph "Camada de Gateway / Proxy"
        Gateway[API Gateway (porta: 8080)]
    end

    subgraph "Camada de Microsserviços"
        US[User Service (porta: 8081)]
        BS[Book Service (porta: 8082)]
        LS[Loan Service (porta: 8083)]
    end

    subgraph "Camada de Dados"
        DB_US[(PostgreSQL\nusers_db)]
        DB_BS[(PostgreSQL\nbooks_db)]
        DB_LS[(PostgreSQL\nloans_db)]
    end

    Gateway -- "/api/users/*" --> US
    Gateway -- "/api/books/*" --> BS
    Gateway -- "/api/loans/*" --> LS

    US --> DB_US
    BS --> DB_BS
    LS --> DB_LS

    %% Comunicação entre microserviços
    LS -- "1. Valida Usuário" --> US
    LS -- "2. Valida Livro/Estoque" --> BS
```

### Detalhamento dos Componentes

1. **API Gateway (`/gateway`)**
   - Porta `8080`.
   - Único ponto de entrada. Encaminha rotas iniciadas com `/api/users`, `/api/books` e `/api/loans` para os microsserviços correspondentes usando proxy reverso (`httputil.NewSingleHostReverseProxy`).

2. **User Service (`/user-service`)**
   - Porta `8081`. 
   - Gerencia exclusivamente entidades de "Usuário".
   - Banco de dados isolado: `users_db` (banco de dados via docker).

3. **Book Service (`/book-service`)**
   - Porta `8082`. 
   - Gerencia a entidade "Livro" do acervo.
   - Banco de dados isolado: `books_db` (banco de dados via docker).

4. **Loan Service (`/loan-service`)**
   - Porta `8083`. 
   - Gerencia "Empréstimos".
   - Banco de dados isolado: `loans_db` (banco de dados via docker).
   - Possui o `client.ServiceClient` para realizar GET em `User Service` e `Book Service` antes de confirmar um novo empréstimo (evitando a falta de exemplares).

---

## Endpoints Expostos Pelo Gateway

A comunicação de fora deve ser feita **sempre através das rotas do API Gateway (porta 8080)**.

### Usuários (`/api/users`) - Direcionados ao User Service
- `POST /api/users` - Criar usuário (`name`, `email`).
- `GET /api/users` - Listar todos os usuários.
- `GET /api/users/{id}` - Buscar usuário por ID.

### Livros (`/api/books`) - Direcionados ao Book Service
- `POST /api/books` - Cadastrar livro (`title`, `author`, `isbn`, `quantity`).
- `GET /api/books` - Listar livros (busca opcional com `?q=termo`).
- `GET /api/books/{id}` - Buscar livro por ID.

### Empréstimos (`/api/loans`) - Direcionados ao Loan Service
- `POST /api/loans` - Registrar empréstimo (`user_id`, `book_id`). O Loan Service vai validar as informações com os outros dois microsserviços sob o capô.
- `GET /api/loans` - Listar empréstimos (filtro opcional com `?user_id=X`).
- `PATCH /api/loans/{id}/return` - Devolver livro atualizando status do empréstimo.

---

## Como Rodar

O diretório traz um `docker-compose.yml` raiz que contém as **três instâncias de banco de dados** necessárias (`postgres-users`, `postgres-books`, `postgres-loans`).

1. **Subir os 3 bancos de dados PostgreSQL:**
```bash
docker compose up -d
```
> Os bancos `postgres-users`, `postgres-books` e `postgres-loans` ficarão disponíveis respectivamente nas portas `5433`, `5434` e `5435`.

2. **Rodar cada serviço separadamente:**
Em quatro terminais diferentes, a partir do diretório do serviço base (`/micro`), inicie os processos Go. As `migrations` rodarão de forma autônoma na inicialização de cada microsserviço.

```bash
# Terminal 1 - User Service
cd user-service
go run ./cmd/server

# Terminal 2 - Book Service
cd book-service
go run ./cmd/server

# Terminal 3 - Loan Service
cd loan-service
go run ./cmd/server

# Terminal 4 - API Gateway
cd gateway
go run ./cmd/main.go
```

A API final, agregada, estará rodando via API Gateway em:
**`http://localhost:8080/api/`**
