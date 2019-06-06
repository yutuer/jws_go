package main

import (
	"net/http"

	"fmt"
	"strconv"
	"sync"

	"strings"

	"github.com/gin-gonic/gin"
	"github.com/syndtr/goleveldb/leveldb"
)

func main() {
	r := gin.Default()

	db, err := leveldb.OpenFile("proto.db", nil)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	var rw sync.RWMutex
	r.GET("/user/*action", func(c *gin.Context) {
		name := c.Param("name")
		action := c.Param("action")
		message := name + " is " + action
		c.String(http.StatusOK, message)
	})

	jws := r.Group("/prototag")
	{
		jws.GET("/*tagname", func(c *gin.Context) {
			//alltagname := strings.TrimPrefix(c.Param("tagname"), "/")

			//println(c.Param("tagname"))
			key := []byte(c.Param("tagname"))
			rawpath := strings.TrimSuffix(c.Param("tagname"), "/")
			allkeys := strings.Split(rawpath, "/")
			lenAllKeys := len(allkeys)
			countertag := strings.Join(allkeys[:lenAllKeys-1], "/")
			counterKey := []byte(countertag + "/")

			ok, err := db.Has(counterKey, nil)
			if err != nil {
				panic(err)
			} else {
				if !ok {
					func() {
						rw.Lock()
						defer rw.Unlock()
						if err := db.Put(counterKey, []byte(fmt.Sprintf("%d", 1)), nil); err != nil {
							panic(err)
						}
					}()
				}
			}

			ok, err = db.Has(key, nil)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"status": false,
					"error":  err.Error(),
				})
				return
			}

			if ok {
				ret := func() bool {
					//rw.RLock()
					//defer rw.RUnlock()
					if raw, err := db.Get(key, nil); err != nil {
						panic(err)
					} else {
						if i, err := strconv.Atoi(string(raw)); err != nil {
							panic(err)
						} else {
							iscounter := strings.HasSuffix(string(key), "/")
							c.JSON(http.StatusOK, gin.H{
								"status": true,
								"tag":    i,
								"c":      iscounter,
							})
							return true
						}
					}
					return false
				}()
				if ret {
					return
				}
			} else {
				ret := func() bool {
					rw.Lock()
					defer rw.Unlock()
					if raw, err := db.Get(counterKey, nil); err != nil {
						panic(err)
					} else {
						if i, err := strconv.Atoi(string(raw)); err != nil {
							panic(err)
						} else {
							if err := db.Put(key, []byte(fmt.Sprintf("%d", i)), nil); err != nil {
								panic(err)
							} else {
								if err := db.Put(counterKey, []byte(fmt.Sprintf("%d", i+1)), nil); err != nil {
									panic(err)
								} else {
									iscounter := strings.HasSuffix(string(key), "/")
									c.JSON(http.StatusOK, gin.H{
										"status": true,
										"tag":    i,
										"c":      iscounter,
									})
									return true
								}
							}
						}
					}
					return false
				}()
				if ret {
					return
				}
			}

			c.JSON(http.StatusOK, gin.H{
				"status": false,
				"error":  "error",
			})
		})
	}
	r.Run(":9876")
}
