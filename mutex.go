package mutex

import (
	"github.com/bradfitz/gomemcache/memcache"
	"time"
	"errors"
	"sync"
)

const (
// DefaultExpiry is used when Mutex Duration is 0
	DefaultExpiry = 8 * time.Second
// DefaultTries is used when Mutex Duration is 0
	DefaultTries = 16
// DefaultDelay is used when Mutex Delay is 0
	DefaultDelay = 512 * time.Millisecond
// DefaultFactor is used when Mutex Factor is 0
	DefaultFactor = 0.01
)

var (
	// ErrFailed is returned when lock cannot be acquired
	ErrFailed = errors.New("failed to acquire lock")
	// default Memcached Client
	defaultMemcache *memcache.Client
)

// Locker interface with Lock returning an error when lock cannot be aquired
type Locker interface {
	Lock() error
	Unlock()
}

type RWMutex struct {
	Name string
	Expiry time.Duration

	Tries int           // Number of attempts to acquire lock before admitting failure, DefaultTries if 0
	Delay time.Duration // Delay between two attempts to acquire lock, DefaultDelay if 0

	Factor float64 // Drift factor, DefaultFactor if 0

	value string
	until time.Time
}

type MemcacheRWMutex struct{
	RWMutex
	client *memcache.Client
	clientm sync.Mutex
}

//
func SetDefaultMemcache(client *memcache.Client){
	defaultMemcache = client
}


//
func NewMemcacheMutex(name string, clients ...*memcache.Client)(mRWMutex *MemcacheRWMutex){
	mRWMutex = &MemcacheRWMutex{}
	if len(clients) > 0 {
		mRWMutex.client = clients[0]
	}else{
		mRWMutex.client = defaultMemcache
	}
	mRWMutex.Name = "mutex-" + name
	return
}

//
func (this *MemcacheRWMutex) Lock() error{
	this.clientm.Lock()
	defer this.clientm.Unlock()

	value := uuid()

	expiry := this.Expiry
	if expiry == 0 {
		expiry = DefaultExpiry
	}

	retries := this.Tries
	if retries == 0 {
		retries = DefaultTries
	}

	for i := 0; i < retries; i++ {
		start := time.Now()


		err := this.client.Add(&memcache.Item{
			Key: this.Name,
			Value: []byte(value),
			Expiration: int32(expiry / time.Second),
		})

		factor := this.Factor
		if factor == 0 {
			factor = DefaultFactor
		}
		until := time.Now().Add(
			expiry - time.Now().Sub(start) -
			time.Duration(int64(float64(expiry) * factor)) +
			2 * time.Millisecond)
		if time.Now().Before(until) && err == nil {
			this.value = value
			this.until = until
			return nil
		}

		delay := this.Delay
		if delay == 0 {
			delay = DefaultDelay
		}
		time.Sleep(delay)
	}

	return ErrFailed
}

//
func (this *MemcacheRWMutex) Unlock() {
	this.clientm.Lock()
	defer this.clientm.Unlock()

	value := this.value
	if value == "" {
		panic("redsync: unlock of unlocked mutex")
	}

	this.value = ""
	this.until = time.Unix(0, 0)
	cas := this.Name + "-CAS"
	err := this.client.Add(&memcache.Item{
		Key: cas,
		Value: []byte(value),
		Expiration: 6,
	})
	if err != nil {
		return
	}
	item, _ := this.client.Get(this.Name)
	if string(item.Value) == value {
		this.client.Delete(this.Name)
	}
	this.client.Delete(cas)
}