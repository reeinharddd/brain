# Estrategia de Despliegue Completo - Brain + MCP + Skills

## 🎯 Objetivo
Sistema auto-contenido que funciona en cualquier dispositivo con un solo comando.

## 📋 Componentes a Integrar

### 1. **Brain Repo Components** ✅
- Rules y adapters (symlinks funcionan)
- Agents y commands (symlinks funcionan)
- Hooks y validaciones (funcionan)

### 2. **MCP Servers** 🔄
**Actuales (Docker):**
- ✅ memory (port 3001)
- ✅ filesystem (port 3002) 
- ✅ sequential-thinking (port 3003)
- ✅ context7 (port 3005)

**Faltantes (agregar a Docker):**
- ❌ github (port 3004) - requiere GITHUB_TOKEN
- ❌ postgres (port 3006) - requiere DATABASE_URL
- ❌ slack (port 3007) - requiere SLACK_BOT_TOKEN
- ❌ linear (port 3008) - requiere LINEAR_API_KEY

### 3. **Skills Externas** ❌
**Detectadas:**
- research skill (web search + citations)
- frontend-design skill (UI components)
- debugging-methodology skill
- architecture-patterns skill
- etc.

**Problema:** No están en este repo, solo en el dispositivo.

## 🚀 Estrategia Recomendada

### Opción A: **Todo en Docker** (Recomendado)
```
brain/
├── docker/
│   ├── docker-compose.yml (todos los MCPs)
│   ├── .env (todos los tokens)
│   └── skills/ (imágenes Docker de skills)
├── agents/ (symlinks a Claude Code)
├── commands/ (symlinks a Claude Code)
└── scripts/
    ├── install.sh (instala todo)
    ├── deploy.sh (despliega en nuevo dispositivo)
    └── backup.sh (backup de configuración)
```

### Opción B: **Híbrido Docker + Local**
```
brain/
├── docker/ (solo MCPs pesados)
├── local/ (skills livianas + MCPs simples)
└── install.sh (detecta y elige estrategia)
```

## 📦 Plan de Implementación

### Fase 1: **Completar MCP Docker**
1. Agregar GitHub MCP al docker-compose.yml
2. Agregar PostgreSQL MCP (opcional, por proyecto)
3. Agregar Slack/Linear MCPs (opcionales)
4. Configurar variables de entorno seguras

### Fase 2: **Skills Containerizadas**
1. Identificar skills críticas
2. Crear imágenes Docker para cada skill
3. Exponer skills como MCP servers
4. Integrar en docker-compose.yml

### Fase 3: **Deploy Automatizado**
1. Script `deploy.sh` que:
   - Clona brain repo
   - Configura Docker
   - Inicia todos los servicios
   - Verifica funcionamiento

### Fase 4: **Backup/Restore**
1. Exportar configuración y tokens
2. Backup de datos MCP (memory, etc.)
3. Script de restore para nuevo dispositivo

## 🔧 Configuración de Tokens

### Archivo `.env` Seguro:
```bash
# Required for full deployment
GITHUB_TOKEN=ghp_...
SLACK_BOT_TOKEN=xoxb-...
LINEAR_API_KEY=lin_api_...

# Optional (per project)
DATABASE_URL=postgresql://...

# Device-specific
HOST_HOME=/home/username
DEVICE_ID=unique-device-id
```

### Manejo de Secrets:
1. Nunca commitear `.env`
2. Usar `.env.example` como template
3. Cifrar tokens con `gpg` o `ansible-vault`
4. Generar `.env` en deploy time

## 🎪 Escenario de Nuevo Dispositivo

### Comando Único:
```bash
curl -fsSL https://raw.githubusercontent.com/reeinharrrd/brain/main/scripts/deploy.sh | bash
```

### Qué hace el script:
1. **Detecta entorno** (OS, Docker, etc.)
2. **Clona brain repo** en `~/.brain`
3. **Instala dependencias** (Docker, Node, etc.)
4. **Configura MCPs** con puertos únicos
5. **Inicia servicios** via docker-compose
6. **Verifica funcionamiento** de todos los componentes
7. **Configura Claude Code** con settings correctos

## 📊 Métricas de Éxito

### Funcionalidad Completa:
- ✅ Todos los MCPs corriendo
- ✅ Skills disponibles
- ✅ Agents y commands accesibles
- ✅ Memoria persistente
- ✅ Configuración segura

### Portabilidad:
- ✅ Un comando para deploy
- ✅ Funciona en Linux/macOS/WSL
- ✅ Backup/restore sencillo
- ✅ Actualizaciones automáticas

## 🔄 Próximos Pasos

1. **Analizar skills actuales** - Identificar cuáles son críticas
2. **Dockerizar skills** - Convertir a MCP servers
3. **Completar docker-compose** - Agregar todos los MCPs
4. **Crear deploy.sh** - Script de despliegue completo
5. **Testear en dispositivo limpio** - Validar portabilidad
