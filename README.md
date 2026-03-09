# shorten-go

A fast, minimal URL shortener written in **Go**.

## Stack

- **[Fiber](https://gofiber.io/)** HTTP framework
- **[BadgerDB](https://github.com/dgraph-io/badger)** embedded key/value store (no external DB **needed**)

## Project Structure

```
shorten-go/
├── cmd/
│   └── shorten/
│       └── main.go       # Entry point
├── internal/
│   ├── handler/
│   │   └── handler.go    # HTTP handlers
│   ├── middleware/
│   │   └── auth.go       # API key middleware
│   ├── store/
│   │   └── store.go      # BadgerDB logic
│   └── model/
│       └── url.go        # URL model
├── .env                  # Environment variables
├── .gitignore
├── README.md
└── go.mod
```

## API

| Method | Route | Description |
|--------|-------|-------------|
| `POST` | `/shorten` | Create a short URL |
| `GET` | `/:code` | Redirect to original URL |
| `DELETE` | `/:code` | Delete a short URL (API key required) |

### POST /shorten

**Request:**
```json
{
  "url": "https://example.com/very/long/url"
}
```

**Response:**
```json
{
  "short": "https://example.com/abc123",
  "code": "abc123"
}
```

### DELETE /:code

**Headers:**
```
X-API-Key: your-api-key
```

## Security

- **Rate limiting**: 10 requests/minute per **IP** on `POST /shorten`
- **URL validation**: only `http://` and `https://` accepted
- **API key**: required for `DELETE /:code`

## Getting Started

### Prerequisites

- Go 1.21+

### Install

```bash
git clone https://github.com/UneBaguette/shorten-go
cd shorten-go
go mod tidy
```

### Configure

```bash
cp .env.example .env
```

```env
PORT=3000
BASE_URL=https://example.com
DB_PATH=./data/badger
API_KEY=your-api-key
```

### Run

```bash
go run ./cmd/shorten
```

### Build

```bash
go build -o shorten ./cmd/shorten
./shorten
```

## License

This project is **licensed** under the **MIT License**. See the [LICENSE](LICENSE) file for details.
