# 🚀 Go Command Server

A lightweight, extensible HTTP server that executes commands via JSON payloads, now with streaming support! Perfect for creating microservices, API gateways, or real-time data processing services.

## ✨ Features

- 🔌 Plugin-based architecture - add new commands easily
- 🌊 Streaming support via Server-Sent Events (SSE)
- 🌐 Built-in CORS support
- 🔒 JSON-based communication
- 📦 Includes powerful `fetch` command out of the box
- 🛠️ Easy to extend and customize

## 🏃‍♂️ Quick Start

### Installation

```bash
git clone <your-repo>
cd github.com/tluyben/command-server
go build -o server
```

### Running the Server

```bash
# Run on default port 8080
./server

# Run on custom port with CORS enabled
./server -port 3000 -cors "*"
```

## 💻 API Usage

All commands use POST requests with JSON payloads in this format:

```json
{
  "cmd": "command-name",
  "args": {
    // command-specific arguments
  }
}
```

### Response Formats

#### Regular JSON Response

```json
{
  "statuscode": 200,
  "headers": {
    "Content-Type": ["application/json"]
  },
  "body": {
    // command-specific response
  }
}
```

#### Streaming Response (SSE)

```
event: start
data: {"statuscode":200,"headers":{...}}

event: data
data: {"some":"data"}

event: data
data: {"more":"data"}

event: end
data: {}
```

## 📚 Built-in Commands

### 🔄 Fetch Command

Makes HTTP requests to external services with full control over method, headers, and body. Supports both regular and streaming responses.

#### Regular Request Example:

```bash
curl -X POST http://localhost:8080 -d '{
  "cmd": "fetch",
  "args": {
    "method": "POST",
    "url": "https://api.example.com/data",
    "headers": {
      "Authorization": "Bearer token123",
      "Content-Type": "application/json"
    },
    "body": {
      "key": "value"
    }
  }
}'
```

#### Streaming Request Example:

```bash
curl -X POST http://localhost:8080 -H "Accept: text/event-stream" -d '{
  "cmd": "fetch",
  "args": {
    "method": "GET",
    "url": "https://api.example.com/stream",
    "stream": true
  }
}'
```

#### Additional Test Commands:

```bash
# Test GET request to public API
curl -X POST http://localhost:8080 -d '{
  "cmd": "fetch",
  "args": {
    "method": "GET",
    "url": "https://jsonplaceholder.typicode.com/posts/1"
  }
}'

# Test with custom headers
curl -X POST http://localhost:8080 -d '{
  "cmd": "fetch",
  "args": {
    "method": "GET",
    "url": "https://api.github.com/users/octocat",
    "headers": {
      "Accept": "application/vnd.github.v3+json",
      "User-Agent": "command-server"
    }
  }
}'

# Test POST with JSON body
curl -X POST http://localhost:8080 -d '{
  "cmd": "fetch",
  "args": {
    "method": "POST",
    "url": "https://jsonplaceholder.typicode.com/posts",
    "headers": {
      "Content-Type": "application/json"
    },
    "body": {
      "title": "Test Post",
      "body": "This is a test post",
      "userId": 1
    }
  }
}'

# Test streaming with SSE
curl -X POST http://localhost:8080 -H "Accept: text/event-stream" -d '{
  "cmd": "fetch",
  "args": {
    "method": "GET",
    "url": "https://stream.data.alpaca.markets/v2/iex",
    "stream": true,
    "headers": {
      "Content-Type": "application/json"
    }
  }
}'
```

### 🌐 Testing CORS

To test CORS functionality:

1. Start the server with CORS enabled:
```bash
./server -cors "*"
```

2. Test from different origins:

```javascript
// In a web browser console or HTML file served from a different domain
async function testCORS() {
  try {
    const response = await fetch('http://localhost:8080', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        cmd: 'fetch',
        args: {
          method: 'GET',
          url: 'https://jsonplaceholder.typicode.com/posts/1'
        }
      })
    });
    const data = await response.json();
    console.log('CORS test successful:', data);
  } catch (error) {
    console.error('CORS test failed:', error);
  }
}

// Test streaming with CORS
const eventSource = new EventSource('http://localhost:8080');
eventSource.onmessage = (event) => {
  console.log('Received:', event.data);
};
eventSource.onerror = (error) => {
  console.error('SSE Error:', error);
};
```

3. Quick HTML test file (save as test-cors.html):
```html
<!DOCTYPE html>
<html>
<head>
    <title>CORS Test</title>
</head>
<body>
    <h1>CORS Test</h1>
    <button onclick="testCORS()">Test CORS</button>
    <pre id="result"></pre>
    <script>
        async function testCORS() {
            const result = document.getElementById('result');
            try {
                const response = await fetch('http://localhost:8080', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({
                        cmd: 'fetch',
                        args: {
                            method: 'GET',
                            url: 'https://jsonplaceholder.typicode.com/posts/1'
                        }
                    })
                });
                const data = await response.json();
                result.textContent = JSON.stringify(data, null, 2);
            } catch (error) {
                result.textContent = 'Error: ' + error.message;
            }
        }
    </script>
</body>
</html>
```

To test the HTML file:
1. Save it somewhere outside the project
2. Start a simple HTTP server in that directory:
```bash
# Python 3
python3 -m http.server 8081
# or Python 2
python -m SimpleHTTPServer 8081
```
3. Open http://localhost:8081/test-cors.html in your browser
4. Click the "Test CORS" button to verify CORS is working

### 🌊 Streaming Support

The server supports real-time data streaming using Server-Sent Events (SSE). To use streaming:

1. Set `"stream": true` in your command args
2. Handle SSE events on the client side
3. Events types include:
   - `start`: Initial response with status and headers
   - `data`: Streamed data chunks
   - `error`: Any errors during streaming
   - `end`: Stream completion

#### Browser Client Example:

```javascript
const eventSource = new EventSource("/your-streaming-endpoint");

eventSource.addEventListener("start", (e) => {
  const data = JSON.parse(e.data);
  console.log("Stream started:", data);
});

eventSource.addEventListener("data", (e) => {
  const data = JSON.parse(e.data);
  console.log("Received data:", data);
});

eventSource.addEventListener("error", (e) => {
  const error = JSON.parse(e.data);
  console.error("Stream error:", error);
});

eventSource.addEventListener("end", () => {
  console.log("Stream ended");
  eventSource.close();
});
```

## 🔧 Adding New Commands

1. Create a new file in the `commands` directory (e.g., `commands/mycommand.go`)
2. Implement the Command interface:

```go
type Command interface {
    Execute(args map[string]interface{}, writer ResponseWriter) error
}
```

3. Use the ResponseWriter interface for output:

```go
type ResponseWriter interface {
    // For regular JSON responses
    WriteJSON(statusCode int, headers map[string][]string, body interface{}) error

    // For streaming responses
    Stream(eventType string, data interface{}) error
    End() error
}
```

4. Register your command in the `init()` function:

```go
func init() {
    RegisterCommand("mycommand", &MyCommand{})
}
```

## 🚨 Error Handling

### Regular Responses

- All errors are returned as JSON responses
- Appropriate HTTP status codes are used
- Error messages are human-readable

Example error response:

```json
{
  "statuscode": 400,
  "headers": {
    "Content-Type": ["application/json"]
  },
  "body": {
    "error": "Invalid request format"
  }
}
```

### Streaming Errors

- Errors during streaming are sent as error events
- Stream continues if possible
- Automatic stream termination on fatal errors

Example streaming error:

```
event: error
data: {"error": "Connection timeout"}
```

## 🔐 Security Considerations

- CORS can be disabled for production use
- No authentication included by default - add middleware as needed
- Be cautious with the fetch command in production environments
- Consider rate limiting for streaming responses
- Monitor memory usage with long-running streams

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch
3. Add your command to the `commands` directory
4. Submit a pull request

### Streaming Best Practices

When adding streaming support to commands:

- Use appropriate chunk sizes
- Include proper error handling
- Implement timeout mechanisms
- Clean up resources properly
- Document streaming behavior

## 📄 License

MIT License - feel free to use for any project!

## 🐛 Known Issues

- Some browsers limit the number of concurrent SSE connections
- Long-running streams may require connection keep-alive handling

## 🙋‍♂️ Need Help?

- Submit an issue for bug reports
- Start a discussion for feature requests
- Check existing issues before submitting new ones

Happy streaming! 🌊
