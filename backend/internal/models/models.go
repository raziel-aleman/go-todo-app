package models

type Todo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
	Body  string `json:"body"`
}

type NewTodo struct {
	Title       string `json:"title"`
	Description string `json:"body"`
}
