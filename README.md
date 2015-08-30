## About

This is a distributed mutex library for the Go programming language
(http://golang.org/).

Now only support Memcached.

## Installing

### Using *go get*

    $ go get github.com/bradfitz/gomemcache/memcache
    $ go get github.com/harrykobe/mutex

After this command *mutex* is ready to use. Its source will be in:

    $GOPATH/src/github.com/harrykobe/mutex

## Example

    import (
            "github.com/bradfitz/gomemcache/memcache"
            "github.com/harrykobe/mutex"
    )

    func main() {
        mc := memcache.New("127.0.0.1:11211")
	      mmutex := NewMemcacheMutex("test", mc)
	      mmutex.Lock()
	       ...
	      mmutex.Unlock()
    }

## Full docs, see:

Coming soon...
