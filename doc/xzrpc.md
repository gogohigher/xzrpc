## xzrpc     

### 消息池化
```go
var MessagePool sync.Pool

func init() {
	MessagePool = sync.Pool{New: func() any {
		return NewMessage()
	}}
}
```
