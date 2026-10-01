# Benchmark Suite

Este diretório contém o fluxo de benchmark para comparar a versão monolítica e a versão em microsserviços da aplicação.

## O que ele gera

- `request_metrics.csv`: latência e status code por requisição.
- `summary.csv`: agregados por cenário e endpoint.
- `container_metrics.csv`: métricas de CPU e memória coletadas via `docker stats`.
- `plots/`: gráficos prontos para inserir no TCC.

## Instalação

```bash
python3 -m venv .venv
source .venv/bin/activate
python -m pip install --upgrade pip
pip install -r benchmarks/requirements.txt
```

Se o `venv` ainda não existir ou o comando acima falhar por falta do módulo, instale o pacote do sistema para ambientes virtuais:

```bash
sudo apt install python3-venv
```

## Execução

> Se estiver usando `venv`, ative-o antes de rodar qualquer comando abaixo:
>
> ```bash
> source .venv/bin/activate
> ```

### 1. Suba a stack que será testada

Monólito:

```bash
cd library/mono
docker compose up -d
go run ./cmd/server
```

Microsserviços:

```bash
cd library/micro
docker compose up -d
cd user-service && go run ./cmd/server
cd book-service && go run ./cmd/server
cd loan-service && go run ./cmd/server
cd gateway && go run ./cmd/main.go
```

### 2. Rode o benchmark

Monólito:

```bash
python3 benchmarks/benchmark.py \
  --stack mono \
  --base-url http://localhost:8080 \
  --compose-dir mono \
  --output-dir benchmarks/results
```

Microsserviços:

```bash
python3 benchmarks/benchmark.py \
  --stack micro \
  --base-url http://localhost:8180 \
  --compose-dir micro \
  --output-dir benchmarks/results
```

### 3. Gere os gráficos

```bash
python3 benchmarks/plot_results.py --input-dir benchmarks/results/mono
python3 benchmarks/plot_results.py --input-dir benchmarks/results/micro
```

Se aparecer `ModuleNotFoundError: No module named 'matplotlib'`, instale as dependências dentro do ambiente virtual:

```bash
python -m pip install -r benchmarks/requirements.txt
```

## Observações

- O benchmark faz uma fase de `seed` antes das medições para garantir dados suficientes nos `GET`.
- Os cenários cobrem leitura e escrita nos três domínios da API.
- Para comparação justa, rode a mesma quantidade de requisições e concorrência nas duas arquiteturas.
