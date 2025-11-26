package entity

type Task struct {
	ID      uint64
	IsCheck bool
	Links   []Link
}

type Link struct {
	Url    string
	Status string
}

type RequestTask struct {
	Links []string `json:"links"`
}
type RequestTaskList struct {
	Ids []uint64 `json:"links_list"`
}
