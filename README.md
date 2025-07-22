# S3 Browser 🗂️

A modern, web-based file manager for S3-compatible storage systems. Built with Go backend and Vue.js frontend, featuring session-based authentication, real-time operations, and a beautiful user interface.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/go-1.24.4-blue.svg)](https://golang.org/)
[![Vue Version](https://img.shields.io/badge/vue-3.5.17-green.svg)](https://vuejs.org/)

## 🚀 Features

### 🔐 **Session-Based Authentication**
- Secure session management with HTTP-only cookies
- Connection testing before establishing sessions
- Automatic session cleanup (24-hour expiration)
- Support for custom S3 endpoints, regions, and credentials

### 🪣 **Bucket Management**
- **List Buckets**: View all your S3 buckets with creation dates
- **Create Buckets**: Create new buckets with validation
- **Delete Buckets**: Safe deletion with confirmation dialogs
- **Real-time Error Handling**: User-friendly error messages with auto-dismiss

### 📁 **Object Operations**
- **Browse Objects**: Navigate through your bucket contents
- **Upload Files**: Drag-and-drop or click-to-select file uploads
- **Download Objects**: Direct download with proper filenames
- **View Objects**: Preview files directly in the browser
- **Delete Objects**: Remove objects with confirmation

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Vue.js SPA    │    │   Go Backend    │    │  S3 Compatible  │
│                 │    │                 │    │     Storage     │
│ • Bucket List   │◄──►│ • Session Mgmt  │◄──►│                 │
│ • Object List   │    │ • S3 API Proxy  │    │ • MinIO         │
│ • File Upload   │    │ • Error Handling│    │ • AWS S3        │
│ • Modals        │    │ • Static Files  │    │ • Other S3 APIs │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 🛠️ Technology Stack

### Backend (Go)
- **Framework**: Gorilla Mux for HTTP routing
- **S3 SDK**: AWS SDK for Go v2
- **Session Management**: UUID-based sessions with in-memory storage
- **Logging**: Structured logging with slog
- **Documentation**: Swagger API documentation

### Frontend (Vue.js)
- **Framework**: Vue 3 with Composition API
- **Routing**: Vue Router 4
- **Styling**: Tailwind CSS
- **Build Tool**: Vite
- **Language**: TypeScript

## 📦 Installation

### Prerequisites
- Go 1.24.4 or later
- Node.js 18+ and npm (for frontend development)
- Access to S3-compatible storage (AWS S3, MinIO, etc.)

### Option 1: Download Binary (Recommended)

```bash
# Download the latest release for your platform
# Linux
wget https://github.com/cksidharthan/s3-browser/releases/latest/download/s3-browser-linux-amd64
chmod +x s3-browser-linux-amd64
./s3-browser-linux-amd64

# macOS
wget https://github.com/cksidharthan/s3-browser/releases/latest/download/s3-browser-darwin-amd64
chmod +x s3-browser-darwin-amd64
./s3-browser-darwin-amd64

You can also use homebrew to install the latest release
brew tap cksidharthan/homebew-tap
brew install s3-browser

# Windows
Download s3-browser-windows-amd64.exe and run it
```

### Option 2: Build from Source

```bash
# Clone the repository
git clone https://github.com/cksidharthan/s3-browser.git
cd s3-browser

# Build the frontend
cd frontend
npm install
npm run build
cd ..

# Build the Go binary
go build -o s3-browser main.go

# Run the application
./s3-browser
```

## 🚀 Quick Start

1. **Start the application**:
   ```bash
   ./s3-browser
   ```
   The server will start on `http://localhost:8080`

2. **Access the web interface**: Open your browser and navigate to `http://localhost:8080`

3. **Connect to your S3 storage**:
   - Enter your S3 endpoint (e.g., `https://s3.amazonaws.com` or `http://localhost:9000` for MinIO)
   - Provide your region (e.g., `us-east-1`, `eu-west-1`)
   - Enter your Access Key ID and Secret Access Key
   - Choose whether to use SSL/TLS
   - Click "Test Connection and Continue"

4. **Start managing your buckets and objects**!

## 📱 Usage

### Connection Setup
- On first visit, you'll see the connection form
- Fill in your S3 credentials and endpoint details
- The system will test the connection before proceeding
- Successful connections create a secure session

### Managing Buckets
- **View Buckets**: See all buckets with creation dates
- **Create Bucket**: Click the "Create Bucket" button and enter a name
- **Delete Bucket**: Click the delete button (requires empty bucket)
- **Access Bucket**: Click "Open" to browse objects

### Managing Objects
- **Upload Files**: Click "Upload" button or drag files to the upload area
- **View Objects**: Click on object names to preview in browser
- **Download Objects**: Use the download button for each object
- **Delete Objects**: Click delete button with confirmation
- **Navigation**: Use breadcrumbs to navigate folder structures

### Session Management
- **View Connection**: Click "Connection Info" in the navbar
- **Logout**: Click "Logout" to end session and clear data
- **Auto-Logout**: Sessions expire after 24 hours

## 🔧 API Documentation

The application includes built-in Swagger documentation available at:
`http://localhost:8080/api/swagger/`

### Key Endpoints
- `POST /api/connect` - Establish S3 connection and create session
- `GET /api/session/status` - Check current session status
- `POST /api/logout` - Destroy current session
- `GET /api/buckets` - List all buckets
- `PUT /api/buckets/{name}` - Create new bucket
- `DELETE /api/buckets/{name}` - Delete bucket
- `GET /api/objects` - List objects in bucket
- `POST /api/objects/{key}` - Upload object
- `GET /api/objects/{key}` - Download/view object
- `DELETE /api/objects/{key}` - Delete object


## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 📊 Project Status

🟢 **Active Development** - Regular updates and maintenance

---

**Made with ❤️ for the S3 community**
