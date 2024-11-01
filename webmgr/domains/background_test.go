package domains

import (
	"testing"
	"time"
)

func TestWorker_SubmitFnTask(t *testing.T) {
	worker := NewWorker()
	result := worker.SubmitFnTask(func(result chan<- any) {
		result <- "hello world"
	})

	select {
	case v := <-result:
		t.Log(v)
		val := v.(string)
		if val != "hello world" {
			t.Fatal("result mismatch")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}
}
