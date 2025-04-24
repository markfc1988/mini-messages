package main

import (
	"fmt"
	"html/template"
	"net/http"
	"sync"
)

// 定义了每条消息的结构样式
type Message struct {
	Name    string
	Content string
	age     int
}

var (
	messages []Message
	stats    = make(map[string]int)
	mu       sync.Mutex
)

// 做一个中间件 计数用的,所有请求都先请求这，记录访问次数
func countAndHandle(path string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		stats[path]++
		mu.Unlock()
		handler(w, r)
	}
}
// 加载templates/index.html模板
func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	mu.Lock()
	defer mu.Unlock()
	tmpl.Execute(w, messages)
}

// 页面提交数据
func submitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	name := r.FormValue("name")
	content := r.FormValue("message")
	mu.Lock()
	messages = append(messages, Message{Name: name, Content: content})
	mu.Unlock()
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// 查看页面状态，各自访问次数
func statsHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	for path, count := range stats {
		fmt.Fprintf(w, "%s: %d\n", path, count)
	}
}

// 重制访问次数
func resetHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	messages = nil
	mu.Unlock()
	fmt.Fprintln(w, "All messages cleared.")
}

// 主函数，调用各个函数
func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", countAndHandle("/", homeHandler))
	mux.HandleFunc("/submit", countAndHandle("/submit", submitHandler))
	mux.HandleFunc("/stats", countAndHandle("/stats", statsHandler))
	mux.HandleFunc("/reset", countAndHandle("/reset", resetHandler))
	// 打印出信息 并监听80端口
	fmt.Println("🚀 Server running at http://localhost:80/")
	http.ListenAndServe(":80", mux)
}
