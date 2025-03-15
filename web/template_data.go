package web

import "sync"

type TemplateData struct {
	mu   sync.RWMutex
	data map[string]any
}

func NewTemplateData() *TemplateData {
	return &TemplateData{
		data: make(map[string]any),
	}
}

func (td *TemplateData) SetIndex(data any) {
	td.mu.Lock()
	defer td.mu.Unlock()
	td.data["index"] = data
}

func (td *TemplateData) GetIndex() any {
	td.mu.RLock()
	defer td.mu.RUnlock()
	return td.data["index"]
}

func (td *TemplateData) SetLang(key string, data any) {
	td.mu.Lock()
	defer td.mu.Unlock()
	td.data["lang-"+key] = data
}

func (td *TemplateData) GetLang(key string) any {
	td.mu.RLock()
	defer td.mu.RUnlock()
	return td.data["lang-"+key]
}
