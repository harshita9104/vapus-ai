---
id: local
title: Local Setup
sidebar_label: Local
---

# VapusAI Local Development Setup

Set up and run **VapusAI** on your machine by following the instructions below.

---

## 🔧 Prerequisites

:::note
Ensure the following dependencies are installed on your system before proceeding.
:::

- Python **3.10+**
- Node.js **18+** and npm
- Go **1.20+**
- Poetry (Python dependency manager)
- Make (used for Go services)
- Git

---

## Services Overview

| Service       | Port(s)           | Purpose                          | Tech Stack     |
|---------------|------------------|----------------------------------|----------------|
| **AI Tools**  | 9028             | NLP tools & text processing      | Python / gRPC  |
| **AI Studio** | 9019 (gRPC), 9020 | Model management & completions   | Go / gRPC      |
| **AI Gateway**| 9017             | Authentication & API gateway     | Go / HTTP      |
| **Webapp**    | 9014             | Frontend dashboard               | Next.js / React|

---

##  Step 1: Clone the Repository

```bash
git clone https://github.com/vapusdata-ecosystem/vapus-ai
cd vapus-ai
```

---

## ⚙️ Step 2: Setup Local Config

```bash
cp -r config/dev/config local-configs/config
cp -r config/dev/network local-configs/network
```

:::tip
Make sure the `local-configs` directory exists or is created manually.
:::

---

## 🧬 Step 3: Generate Protobuf Code

```bash
cd apis
make sync
```

---

##  Step 4: Setup AI Tools (Python)

```bash
cd app/aitools
poetry install
```

:::info
Dependencies like SpaCy, NLTK, and Transformers will auto-install during first run.
:::

---

## 🖥️ Step 5: Setup Webapp Frontend

```bash
cd app/webapp
npm install
```

---

## 🚀 Step 6: Start All Services

Start each service in a **separate terminal**:

```bash
# Terminal 1 - AI Tools
cd app/aitools
poetry run python server.py --conf=../../local-configs
```

```bash
# Terminal 2 - AI Studio
cd app/aistudio
make runmain CONF=../../local-configs DEBUG=true
```

```bash
# Terminal 3 - AI Gateway
cd app/aigateway
make runmain CONF=../../local-configs DEBUG=true
```

```bash
# Terminal 4 - Webapp Frontend
cd app/webapp
npm run dev -- -p 9014
```

---

## Step 7: Verify Services Are Running

```bash
ss -tlnp | grep -E ":(9014|9017|9019|9028)"
```

You should see:

- 9028 → AI Tools  
- 9019 → AI Studio  
- 9017 → AI Gateway  
- 9014 → Webapp

---

## 🌐 Access the App

- App: [http://localhost:9014](http://localhost:9014)  
- Login: [http://localhost:9014/login](http://localhost:9014/login)

---

## Service Descriptions

### AI Tools
Text processing, embeddings, and entity recognition  
_Built with Python + gRPC_

---

### AI Studio
AI model configuration and completions  
_Built with Go + gRPC / HTTP_

---

### AI Gateway
API authentication and routing  
_Built with Go + HTTP_

---

### Webapp
Frontend dashboard for managing models, chat, prompts  
_Built with Next.js + React_

---

## 🛠️ Troubleshooting

### Poetry Errors

```bash
cd app/aitools
poetry install
```

---

### Port Conflicts

```bash
ss -tlnp | grep 90
kill <PID>
```

---

## Daily Workflow

- Visit [http://localhost:9014](http://localhost:9014)  
- Run/stop services as needed using `Ctrl+C`

---
