package mutex

import(
	"testing"
	"time"
	"github.com/bradfitz/gomemcache/memcache"
	"fmt"
)

func Test_mutex(t *testing.T) {
	c := make(chan time.Time)

	go mutex_fun(c)
	go mutex_fun(c)

	Here:
		select {
		case v := <- c:
			fmt.Println(v)
			goto Here
		case <- time.After(10 * time.Second):
			fmt.Println("timeout")
			break
		}

}

func mutex_fun(c chan time.Time){
	mc := memcache.New("127.0.0.1:11211")
	mmutex := NewMemcacheMutex("test", mc)
	mmutex.Tries = 1000
	err := mmutex.Lock()
	if err != nil {
		fmt.Println("Err")
		c <- time.Now()
		return
	}
	time.Sleep(3 * time.Second)
	mmutex.Unlock()
	fmt.Println("Success")
	c <- time.Now()
}