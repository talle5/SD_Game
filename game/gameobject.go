package game

import "slices"

type GameObject interface {
	Draw()
	Update()
}

var objects = make([]GameObject, 0, 100)

func RunAll() {
	for _, obj := range objects {
		obj.Update()
		obj.Draw()
	}
}

func DestroyAll() {
	objects = objects[:0]
}

func Destroy(obj GameObject) {
	idx := slices.Index(objects, obj)
	if idx != -1 {
		objects = slices.Delete(objects, idx, idx+1)
	}
}

func New[T any, PT interface {
	*T
	GameObject
}](obj T) PT {
	var ptr PT = &obj
	objects = append(objects, ptr)
	return ptr
}
