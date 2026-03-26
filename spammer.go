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
	sem := make(chan struct{}, 30)
	wg := &sync.WaitGroup{}
	batch := make([]User, 0, 2)

	for user := range in {
		u, ok := user.(User)
		if !ok {
			continue
		}
		batch = append(batch, u)
		if len(batch) == 2 {
			wg.Add(1)
			tmp := make([]User, len(batch))
			copy(tmp, batch)
			sem <- struct{}{}
			go func() {
				defer func() {
					wg.Done()
					<-sem
				}()
				messages, err := GetMessages(tmp[0], tmp[1])
				if err != nil {
					return
				}
				for _, msg := range messages {
					out <- msg
				}
			}()
			batch = batch[:0]
		}
	}
	if len(batch) > 0 {
		tmp := make([]User, len(batch))
		copy(tmp, batch)
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			go func() {
				wg.Done()
				<-sem
			}()
			messages, err := GetMessages(tmp[0])
			if err != nil {
				return
			}
			for _, msg := range messages {
				out <- msg
			}
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()
}

func CheckSpam(in, out chan interface{}) {
	// in - MsgID
	// out - MsgData
}

func CombineResults(in, out chan interface{}) {
	// in - MsgData
	// out - string
}
