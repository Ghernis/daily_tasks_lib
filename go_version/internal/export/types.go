package export

type Author struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	SNID  string `json:"sn_id"`
}

type Subtask struct {
	Name string `json:"name"`
}

type Task struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Subtasks    []Subtask `json:"subtasks"`
	Time        float64   `json:"time"`
}

type TicketDaily struct {
	Title       string   `json:"title"`
	Date        string   `json:"date"`
	Description string   `json:"description"`
	Author      []Author `json:"author"`
	Tasks       []Task   `json:"tasks"`
	RITMNumber  *string  `json:"ritm_number"`
}

type Document struct {
	ExportedAt string        `json:"exported_at"`
	Tickets    []TicketDaily `json:"tickets"`
}
