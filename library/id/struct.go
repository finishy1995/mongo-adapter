package id

// WithID 给结构体、对象赋予 ID 功能
type WithID struct {
	id ID
}

// InitID 初始化 ID
func (w *WithID) InitID() {
	w.id = GenerateID()
}

// GetID 获取 ID
func (w *WithID) GetID() ID {
	return w.id
}
