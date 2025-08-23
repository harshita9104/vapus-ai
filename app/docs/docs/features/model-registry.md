---
id: model-registry
title: Model Registry
sidebar_label: Model Registry
---

# VapusAI Model Registry: Complete Guide

The **Model Registry** in VapusAI allows you to connect, configure, authenticate, and test AI models from providers like OpenAI, Anthropic, Azure, AWS Bedrock, or even local services — all from a **unified control panel**.

---

## What Is Model Registry?

:::note
The Model Registry acts as your AI integration layer — abstracting away provider complexity while managing configuration, credentials, and access in one place.
:::

It enables:
- Model registration with metadata
- Scoped access management
- API key & credential configuration
- Optional guardrails (e.g., safety filters)
- One-click test and auto-discovery of models

---

## Step-by-Step: Registering a New Model

### Step 1: Open the Registry

Navigate to:

```
AI Center → Model Registry → Add New Model
```

<!-- ![Add New Model](../../static/img/add-new-model.png) -->

---

### Step 2: Fill in Model Metadata

| Field         | Description                            |
|---------------|----------------------------------------|
| Model Name    | Identifier (e.g., `gpt-4-production`)  |
| Scope         | Who can access this model              |

**Scopes:**
- `USER_SCOPE` → Personal use
- `DOMAIN_SCOPE` → Shared across your workspace
- `ORG_SCOPE` → Global access across organizations

---

### Step 3: Choose a Provider

Select one from the dropdown list:

- OpenAI
- Anthropic
- Ollama
- Azure
- AWS Bedrock
- Local or custom backend

<!-- ![Select Service Provider](../../static/img/select-service-provider.png) -->

---

### Step 4: Enter Network Parameters

| Field          | Example                        |
|----------------|--------------------------------|
| **Endpoint**   | `https://api.openai.com`       |
| **API Version**| `v1` or `2023-12-01-preview`   |
| **Model Path** | `chat/completions`             |

<!-- ![Network Credential](../../static/img/network-credential.png) -->

---

### Step 5: Authentication Setup

Depending on provider, choose:

- **API Key** (OpenAI, Azure)
- **Access Key + Secret** (AWS)
- **x-api-key** (Anthropic)
- **Basic Auth / None** (Local)

:::tip
Custom headers can also be added for advanced providers.
:::

---

### Step 6: (Optional) Attach Guardrails

Guardrails can help:

- Enforce content policies
- Apply safety filters
- Track usage limits

<!-- ![Select Guardrail](../../static/img/select-guardrail.png) -->

---

### Step 7: Submit and Test

Click **Submit** to run a validation and test connection.

If valid, VapusAI:
- Authenticates the config
- Discovers available models
- Registers & stores them securely

<!-- ![Final Submit](../../static/img/test-final.png) -->

---

## What Does “Discover Models” Do?

When toggled ON:

1. Connects to the selected provider  
2. Pulls available model list (e.g., GPT-4, Claude 3)  
3. Categorizes them by type (chat, embedding, etc.)  
4. Stores metadata securely  

:::info
This feature saves you from manually inputting each model's name and capabilities.
:::

---

## Behind the Scenes: Model Storage Format

Each model is stored with:

- Model name and ID
- Supported capabilities
- Endpoint and credentials (🔐 encrypted)
- Guardrail associations
- Scope and permissions

---

## Benefits of Using Model Registry

:::info
Why centralize your model config?
:::

- **Single Interface** across all providers  
- **Encrypted Credentials** in storage  
- **Scoped Access Control**  
- **Integrated Guardrails**  
- **Supports Prompt Studio, Playground, and Chat**  
- **Model Discovery** with zero manual input  

---

## Advanced Options

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

<Tabs>
  <TabItem value="dev" label="For Developers" default>
    <>
      <ul>
        <li>Programmatically register models via SDK</li>
        <li>Fetch registered model metadata for pipelines</li>
        <li>Use model aliases in the completion APIs</li>
      </ul>
    </>
  </TabItem>
  <TabItem value="admin" label="For Admins">
    <>
      <ul>
        <li>Restrict who can register models</li>
        <li>Control provider availability (OpenAI only, etc.)</li>
        <li>Predefine guardrail templates for org/domain</li>
      </ul>
    </>
  </TabItem>
</Tabs>

---

## Summary

- Use the **Model Registry** as the central point to manage LLM providers
- Works with OpenAI, Anthropic, Azure, AWS, and local endpoints
- Supports scopes, guardrails, credential encryption, and auto-discovery

:::caution
Do not share API keys directly in prompts or environment variables. Always use the secure Model Registry.
:::
