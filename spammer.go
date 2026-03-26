package async_pipeline

import "sync"

func RunPipeline(cmds ...cmd) {

}

func SelectUsers(in, out chan interface{}) {
	// 	in - string
	// 	out - User
	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}

	uniqueIds := make(map[uint64]struct{})

	for email := range in {
		wg.Add(1)
		go func() {
			defer func() {
				wg.Done()
			}()
			m, ok := email.(string)
			if !ok {
				return
			}

			user := GetUser(m)
			mu.Lock()
			if _, ok = uniqueIds[user.ID]; !ok {
				uniqueIds[user.ID] = struct{}{}
				out <- user
				mu.Unlock()
			}
		}()
	}
	go func() {
		wg.Wait()
		close(out)
	}()
}

func SelectMessages(in, out chan interface{}) {
	// 	in - User
	// 	out - MsgID
}

func CheckSpam(in, out chan interface{}) {
	// in - MsgID
	// out - MsgData
}

func CombineResults(in, out chan interface{}) {
	// in - MsgData
	// out - string
}
