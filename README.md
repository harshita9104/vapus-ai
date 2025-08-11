# vapus-ai

## Prerequisites

- **Python 3.10+** (for AI Tools service)

- **Node.js 18+ and npm** (for Webapp frontend)

- **Go 1.20+** (for AI Studio, AI Gateway, Webapp backend)

- **Poetry** (for Python dependency management)

- **Make** (for building Go services)

- **git** (for version control)



---



## Services Overview



| Service | Port | Purpose | Technology |

|---------|------|---------|------------|

| **AI Tools** | 9028 | Text processing, embeddings, NLP utilities | Python/gRPC |

| **AI Studio** | 9019 (gRPC), 9020 (HTTP) | AI model management, chat completions | Go/gRPC |

| **AI Gateway** | 9017 | Authentication, routing, API gateway | Go/HTTP |

| **Webapp** | 9014 | Web interface, user dashboard | Next.js/React |



---


## 1. Clone and Setup Repository



```bash

git clone https://github.com/vapusdata-ecosystem/vapus-ai

cd vapus-ai

```



---



## 2. Setup Local Configuration


Ensure your `local-configs` directory contains:



```bash

cp -r config/dev/config local-configs/config

cp -r config/dev/network local-configs/network

```

*(Copy from `config/dev/` if needed, or use existing `local-configs/`)*

---


## 3. Generate Protobuf Code


```bash

cd apis

make sync

```


---


## 4. Setup AI Tools Service (Python)


```bash

cd app/aitools

poetry install

# Note: ML models and NLTK data download automatically on first run

```


---

## 5. Setup Webapp Frontend



```bash

cd app/webapp

npm install

```



---


## 6. Start All Services


### Start services individually


**Terminal 1 - AI Tools Service:**

```bash

cd app/aitools

# If using virtual environment:

source venv/bin/activate

python server.py --conf=/path/to/your/vapus-ai/local-configs

# Or with Poetry:

poetry run python server.py --conf=/path/to/your/vapus-ai/local-configs

```


**Terminal 2 - AI Studio Service:**

```bash

cd app/aistudio

make runmain CONF=/path/to/your/vapus-ai/local-configs DEBUG=true

```


**Terminal 3 - AI Gateway Service:**

```bash

cd app/aigateway

make runmain CONF=/path/to/your/vapus-ai/local-configs DEBUG=true

```


**Terminal 4 - Webapp Frontend:**

```bash

cd app/webapp

npm run dev -- -p 9014

```


---



## 7. Verify Services Are Running



```bash

# Check all services are up

ss -tlnp | grep -E ":(9014|9017|9019|9028)"

```


Expected output:

- **Port 9028** - AI Tools Service (Python gRPC server)

- **Port 9019** - AI Studio Service (Go gRPC server)  

- **Port 9017** - AI Gateway Service (Go HTTP server)

- **Port 9014** - Webapp Service (Next.js frontend)



---


## 8. Access the Application


- **Main Application**: [http://localhost:9014](http://localhost:9014)

- **Login Page**: [http://localhost:9014/login](http://localhost:9014/login)


---


## Service Details



### AI Tools Service (Port 9028)

- **Purpose**: Handles text processing, embeddings, NLP tasks

- **Technology**: Python with gRPC

- **Dependencies**: SpaCy, NLTK, sentence-transformers, presidio

- **Function**: Provides utilities for text analysis, summarization, entity recognition



### AI Studio Service (Port 9019)

- **Purpose**: Core AI model management and chat completions

- **Technology**: Go with gRPC/HTTP gateway

- **Function**: Manages AI model configurations, handles chat requests, applies guardrails



### AI Gateway Service (Port 9017)

- **Purpose**: Authentication and API routing

- **Technology**: Go with HTTP

- **Function**: Handles user authentication, request routing, security



### Webapp (Port 9014)

- **Purpose**: User interface and dashboard

- **Technology**: Next.js/React frontend

- **Function**: Provides web interface for managing AI models, prompts, and chat



---


## Troubleshooting

### AI Tools Service Issues:

```bash

cd app/aitools

poetry install  # Reinstall dependencies

poetry run python server.py --conf=/path/to/local-configs

```


### Port Conflicts:

```bash

# Check what's using ports

ss -tlnp | grep 90

# Kill specific processes if needed

kill <PID>

```



---

## Daily Development Workflow


1. **Access webapp**: [http://localhost:9014](http://localhost:9014)

2. **Stop all services**: `Ctrl+C` in terminal



---