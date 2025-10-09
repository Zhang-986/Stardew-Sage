# 🌾 Stardew Sage - AI-Powered Stardew Valley Assistant

<div align="center">

[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Spring Boot](https://img.shields.io/badge/Spring%20Boot-3.4.3-brightgreen.svg)](https://spring.boot.io)
[![Vue](https://img.shields.io/badge/Vue-2.6.12-brightgreen.svg)](https://vuejs.org)
[![Spring AI](https://img.shields.io/badge/Spring%20AI-1.0.0--M8-orange.svg)](https://spring.io/projects/spring-ai)

**An intelligent AI agent system that brings the world of Stardew Valley to life through conversational AI, RAG vector databases, and multimodal analysis.**

[Features](#-features) • [Architecture](#-architecture) • [Quick Start](#-quick-start) • [Documentation](#-documentation) • [Contributing](#-contributing)

</div>

---

## 📖 Overview

**Stardew Sage** is an innovative AI-powered assistant designed for Stardew Valley fans. It combines the power of modern AI technologies to create an interactive experience where users can:

- 🎂 **Get daily birthday stories** - Discover which Stardew Valley character is celebrating their birthday with engaging AI-generated narratives
- 🎯 **Receive daily missions** - Get personalized daily tasks and challenges inspired by the game
- 🦐 **Query game database** - Ask questions about game mechanics, items, characters, and get intelligent responses powered by MCP (Model Context Protocol)
- 📚 **RAG-powered search** - Leverage vector database search for semantic understanding of game content
- 🖼️ **Image analysis** - Upload game screenshots and get AI-powered analysis and insights using multimodal vision models

## ✨ Features

### 🧠 AI Agent System
- **Intelligent Conversational Agent**: Built with Spring AI and Model Context Protocol (MCP) for context-aware responses
- **Multi-Prompt Strategy**: Different specialized prompts for birthdays, missions, database queries, and RAG searches
- **Streaming Responses**: Real-time SSE (Server-Sent Events) streaming for smooth user experience

### 🗄️ Data Management
- **Comprehensive Game Database**: MySQL database with 20+ tables covering all aspects of Stardew Valley:
  - Characters (NPCs) and their preferences
  - Items (crops, cooking, crafting, minerals, fish, etc.)
  - Game mechanics (seasons, events, quests)
  - Animals, buildings, and locations
- **RAG Vector Store**: Redis-based vector database for semantic search using embeddings
- **Flexible Data Upload**: API endpoints for uploading and managing vector embeddings

### 🎨 Modern Web Interface
- **Vue.js Frontend**: Responsive, game-themed UI with Stardew Valley aesthetics
- **Real-time Updates**: SSE integration for live AI response streaming
- **Multiple Modules**:
  - MCP Chat Interface
  - RAG-powered Q&A
  - Image Analysis Tool
  - Data Management Dashboard

### 🔧 Technical Capabilities
- **Spring AI Integration**: Leverages Spring AI framework for AI operations
- **OpenAI-Compatible APIs**: Works with various LLM providers (GLM-4.5, GPT, etc.)
- **Vector Embeddings**: Text embedding generation and similarity search
- **Multimodal Support**: Vision model integration for image understanding
- **MyBatis Plus**: Efficient database operations with automatic code generation

## 🏗️ Architecture

```
Stardew-Sage/
├── aurora-admin/          # Backend admin module (RuoYi framework)
│   └── Spring Boot application for system management
├── aurora-mcp/            # AI MCP Server module
│   ├── Controller         # REST API endpoints
│   ├── Service            # Business logic and AI integration
│   ├── Model              # Database entities (Stardew Valley data)
│   ├── Mapper             # MyBatis database mappers
│   └── Prompt             # AI prompt templates
├── aurora-ui/             # Frontend Vue.js application
│   ├── src/
│   │   ├── views/        # Vue components
│   │   │   ├── stardew/  # Stardew-specific pages
│   │   │   │   ├── mcp/     # MCP chat interface
│   │   │   │   ├── rag/     # RAG Q&A interface
│   │   │   │   ├── image/   # Image analysis
│   │   │   │   └── show/    # Data display
│   │   ├── api/          # API service layer
│   │   └── assets/       # Static resources
└── database.sql          # Complete database schema and data
```

### Technology Stack

**Backend:**
- Spring Boot 3.4.3
- Spring AI 1.0.0-M8 (MCP Server, OpenAI integration)
- MyBatis Plus 3.5.5
- MySQL 8.0
- Redis (Vector Store)
- Maven

**Frontend:**
- Vue.js 2.6.12
- Element UI 2.15.14
- Axios
- Server-Sent Events (SSE)

**AI/ML:**
- Model Context Protocol (MCP)
- OpenAI-compatible API support
- Vector embeddings (2048 dimensions)
- GLM-4.5V vision model
- Redis vector store for RAG

## 🚀 Quick Start

### Prerequisites

- **Java 17+** (for aurora-mcp)
- **Java 8+** (for aurora-admin)
- **Node.js 8.9+** and npm
- **MySQL 8.0+**
- **Redis 6.0+**
- **AI API Key** (OpenAI, GLM, or compatible provider)

### 1. Database Setup

```bash
# Create and import database
mysql -u root -p
source database.sql
# This creates the 'ry-vue' database with all Stardew Valley data
```

### 2. Backend Setup - MCP Server (aurora-mcp)

```bash
# Navigate to MCP server directory
cd aurora-mcp

# Configure application settings
# Edit src/main/resources/application.yml:
#   - Database connection (MySQL)
#   - Redis connection
#   - AI API key and base URL

# Example configuration:
# spring:
#   datasource:
#     url: jdbc:mysql://localhost:3306/ry-vue
#     username: root
#     password: your_password
#   data:
#     redis:
#       host: localhost
#       port: 6379
#   ai:
#     openai:
#       api-key: your_api_key
#       base-url: your_api_endpoint

# Build and run
mvn clean install
mvn spring-boot:run
# Server starts on port 8099
```

### 3. Backend Setup - Admin Server (aurora-admin)

```bash
# Navigate to admin directory
cd aurora-admin

# Configure application settings
# Edit src/main/resources/application.yml

# Build and run
cd ..
mvn clean install
cd aurora-admin
mvn spring-boot:run
# Server starts on port 8080
```

### 4. Frontend Setup (aurora-ui)

```bash
# Navigate to UI directory
cd aurora-ui

# Install dependencies
npm install
# For faster installation in China:
# npm install --registry=https://registry.npmmirror.com

# Start development server
npm run dev
# Application runs on http://localhost:80
```

### 5. Access the Application

- **Frontend**: http://localhost:80
- **MCP API**: http://localhost:8099
- **Admin API**: http://localhost:8080

Default admin credentials (if using RuoYi system):
- Username: `admin`
- Password: `admin123`

## 📚 Documentation

### API Endpoints

#### MCP Agent APIs (Port 8099)

**Birthday Story Stream**
```http
GET /api/agent/getBirthInfoStream
Content-Type: text/event-stream

Returns: SSE stream with birthday character story
```

**Daily Mission Stream**
```http
GET /api/agent/getMissionInfoStream
Content-Type: text/event-stream

Returns: SSE stream with daily mission
```

**Database Query Stream**
```http
GET /api/agent/getInfoDetail?question=<your_question>
Content-Type: text/event-stream

Returns: SSE stream with query results using MCP tools
```

**RAG Query Stream**
```http
GET /api/agent/getInfoRagDetail?question=<your_question>
Content-Type: text/event-stream

Returns: SSE stream with RAG-enhanced query results
```

**Image Analysis**
```http
POST /api/image/analyze
Content-Type: multipart/form-data

Parameters:
  - file: Image file (JPEG/PNG, max 5MB)
  - question: Optional analysis question

Returns: Stream with image analysis
```

#### Resource Management APIs

**Get All Tables**
```http
GET /api/table/list
Returns: List of all database tables
```

**Get Table Info**
```http
GET /api/table/info?tableName=<table_name>
Returns: Table structure and data
```

**Upload to RAG**
```http
POST /api/table/upload
Body: { tableName, columnSimpleMetas }
Returns: Upload status
```

### Database Schema

The project includes 20+ tables for Stardew Valley data:

- `stardew_people` - NPCs and characters
- `stardew_people_favorite` - Character gift preferences
- `stardew_crop` - Crop information
- `stardew_food`, `stardew_food_material` - Cooking recipes
- `stardew_fish`, `stardew_fish_task` - Fishing data
- `stardew_mineral`, `stardew_antique` - Collectibles
- `stardew_animal`, `stardew_husbandry` - Farm animals
- `stardew_craft`, `stardew_craft_material` - Crafting
- And many more...

### Vector Store

The system uses Redis as a vector database for RAG:
- **Index Name**: `stardew_index`
- **Embedding Model**: embedding-3
- **Dimensions**: 2048
- **Purpose**: Semantic search across game content

## 🎮 Usage Examples

### Example 1: Ask About a Character

```javascript
// Frontend code example
import { getInfoDetailSSE } from '@/api/ai'

const eventSource = getInfoDetailSSE('Tell me about Abigail')
eventSource.onmessage = (event) => {
  console.log(event.data) // Streaming response
}
```

### Example 2: Upload Data to RAG

```javascript
import { uploadToRAG } from '@/api/ai'

const result = await uploadToRAG('stardew_crop', [
  { columnName: 'name_ch', isChose: true },
  { columnName: 'remark', isChose: true }
])
```

### Example 3: Analyze Game Screenshot

```javascript
import { analyzeImage } from '@/api/ai'

const file = document.querySelector('input[type="file"]').files[0]
const response = await analyzeImage(file, 'What crop is this?')
```

## 🔧 Configuration

### AI Model Configuration

Edit `aurora-mcp/src/main/resources/application.yml`:

```yaml
spring:
  ai:
    openai:
      base-url: https://open.bigmodel.cn/api/paas/v4/  # Your AI provider
      api-key: your_api_key_here
      chat:
        options:
          model: GLM-4.5V  # Change to your preferred model
      embedding:
        options:
          model: embedding-3
          dimensions: 2048
    vectorstore:
      redis:
        index-name: stardew_index
        initialize-schema: true
```

### Frontend API Configuration

Edit `aurora-ui/vue.config.js`:

```javascript
const baseUrl = 'http://localhost:8080'  // Backend URL
const port = 80  // Frontend port
```

## 🤝 Contributing

Contributions are welcome! Here's how you can help:

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/AmazingFeature`)
3. **Commit** your changes (`git commit -m 'Add some AmazingFeature'`)
4. **Push** to the branch (`git push origin feature/AmazingFeature`)
5. **Open** a Pull Request

### Development Guidelines

- Follow existing code style and conventions
- Add comments for complex logic
- Update documentation for new features
- Test thoroughly before submitting PR

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **Stardew Valley** by ConcernedApe for the amazing game
- **RuoYi** framework for the admin system foundation
- **Spring AI** team for the excellent AI integration framework
- **OpenAI** and **GLM** for AI model capabilities
- The Stardew Valley community for game data and inspiration

## �� Contact & Support

- **GitHub Issues**: For bug reports and feature requests
- **Discussions**: For questions and community support

---

<div align="center">

**Made with ❤️ for the Stardew Valley community**

⭐ Star this repository if you find it helpful!

</div>
