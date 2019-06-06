package timer

import (
	"fmt"
	"testing"
	"time"
)

func TestTargetTimer(t *testing.T) {
	fmt.Println("start wait 10 second ", time.Now().String())
	c := time.NewTimer(3 * time.Second)
	tt1 := <-c.C
	fmt.Println("end wait 10 second ", tt1.String())

	fmt.Println("start wait 10 second ", time.Now().String())
	c2 := time.AfterFunc(3*time.Second, handle)
	//go func() {
	//	<-time.After(5*time.Second)
	//	c2.Stop()
	//}()
	<-time.After(2 * time.Second)
	c2.Stop()
	fmt.Println("Stop timer 2", time.Now().String())

	fmt.Println("start wait -10 second ", time.Now().String())
	time.AfterFunc(-10*time.Second, handle)

	<-time.After(4 * time.Second)
}

func handle() {
	fmt.Println("handle", time.Now().String())
}
