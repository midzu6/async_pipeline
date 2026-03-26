package async_pipeline

import (
	"fmt"
	"sort"
	"sync"
)

func RunPipeline(cmds ...cmd) {
	if len(cmds) == 0 {
		return
	}

	var in chan interface{}
	wg := &sync.WaitGroup{}

	for _, command := range cmds {
		out := make(chan interface{})

		wg.Add(1)
		go func(cmd cmd, in, out chan interface{}) {
			defer wg.Done()
			cmd(in, out)
			close(out)
		}(command, in, out)

		in = out
	}

	wg.Wait()
}

func SelectUsers(in, out chan interface{}) {
	// 	in - string
	// 	out - User
	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}

	uniqueIds := make(map[uint64]struct{})

	for email := range in {
		wg.Add(1)
		go func(email interface{}) {
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
				mu.Unlock()
				out <- user
			} else {
				mu.Unlock()
			}
		}(email)
	}

	wg.Wait()
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
			go func(tmp []User) {
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
			}(tmp)
			batch = batch[:0]
		}
	}
	if len(batch) > 0 {
		tmp := make([]User, len(batch))
		copy(tmp, batch)
		wg.Add(1)
		sem <- struct{}{}
		go func(tmp []User) {
			defer func() {
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
		}(tmp)
	}

	wg.Wait()
}

func CheckSpam(in, out chan interface{}) {
	// in - MsgID
	// out - MsgData

	wg := &sync.WaitGroup{}
	sem := make(chan struct{}, 5)

	for mId := range in {
		id, ok := mId.(MsgID)
		if !ok {
			continue
		}
		sem <- struct{}{}
		wg.Add(1)
		go func(id MsgID) {
			defer func() {
				wg.Done()
				<-sem
			}()
			isSpam, err := HasSpam(id)
			if err != nil {
				return
			}
			resp := MsgData{
				ID:      id,
				HasSpam: isSpam,
			}
			out <- resp
		}(id)
	}

	wg.Wait()

}

func CombineResults(in, out chan interface{}) {
	// in - MsgData
	// out - string

	arr := make([]MsgData, 0)

	for data := range in {
		dt, ok := data.(MsgData)
		if !ok {
			continue
		}
		arr = append(arr, dt)
	}

	sort.Slice(arr, func(i, j int) bool {
		if arr[i].HasSpam && !arr[j].HasSpam {
			return true
		}
		if !arr[i].HasSpam && arr[j].HasSpam {
			return false
		}
		return arr[i].ID < arr[j].ID
	})

	for _, val := range arr {
		out <- fmt.Sprintf("%v %v", val.HasSpam, val.ID)
	}
}
