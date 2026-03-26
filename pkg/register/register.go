package register

import "sync"

var (
	contentions = make(map[string][]any)
	lock        sync.Mutex
)

func Register(containerName string, object any) {
	lock.Lock()
	defer lock.Unlock()
	contentions[containerName] = append(contentions[containerName], object)
}

func Do(containerName string, f func(objects any) error) error {
	lock.Lock()
	defer lock.Unlock()
	if objects, ok := contentions[containerName]; ok {
		for _, object := range objects {
			if err := f(object); err != nil {
				return err
			}

		}
		delete(contentions, containerName)
	}
	return nil

}
func Clear(containerName string) {
	lock.Lock()
	defer lock.Unlock()
	delete(contentions, containerName)
}
func ClearAll() {
	lock.Lock()
	defer lock.Unlock()
	contentions = make(map[string][]any)
}
