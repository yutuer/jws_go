package redis

import (
	"reflect"
	"testing"

	"vcs.taiyouxi.net/platform/planx/redigo/redis"
)

func TestPipelineCmdBuffer(t *testing.T) {
	c, err := redis.DialDefaultServer()
	if err != nil {
		t.Fatalf("error connection to database, %v", err)
	}
	defer c.Close()

	cb := redis.NewCmdBuffer()

	for _, cmd := range testCommands {
		if err := cb.Send(cmd.args[0].(string), cmd.args[1:]...); err != nil {
			t.Fatalf("Send(%v) returned error %v", cmd.args, err)
		}
	}
	reply, err := c.DoCmdBuffer(cb, false)

	if err != nil {
		t.Errorf("DoCmdBuffer() returned error %v", err)
	}

	values, err2 := redis.Values(reply, err)
	if err2 != nil {
		t.Errorf("DoCmdBuffer() Values returned error %v", err2)
	}

	if len(values) != len(testCommands) {
		t.Fatalf("DoCmdBuffer len(vs) == %d, want %d", len(values), len(testCommands))
	}

	for i, cmd := range testCommands {
		actual := values[i]
		if err != nil {
			t.Fatalf("Receive(%v) returned error %v", cmd.args, err)
		}
		if !reflect.DeepEqual(actual, cmd.expected) {
			t.Errorf("Receive(%v) = %v, want %v", cmd.args, actual, cmd.expected)
		}
	}
}

func TestExecCmdBufferTransaction(t *testing.T) {
	c, err := redis.DialDefaultServer()
	if err != nil {
		t.Fatalf("error connection to database, %v", err)
	}
	defer c.Close()

	// Execute commands that fail before EXEC is called.
	c.Do("DEL", "k0")
	c.Do("ZADD", "k0", 0, 0)
	cb := redis.NewCmdBuffer()
	cb.Send("NOTACOMMAND", "k0", 0, 0)
	cb.Send("ZINCRBY", "k0", 0, 0)
	v, err := c.DoCmdBuffer(cb, true)
	if err == nil {
		t.Fatalf("EXEC returned values %v, expected error", v)
	}

	// Execute commands that fail after EXEC is called. The first command
	// returns an error.
	c.Do("DEL", "k1")
	c.Do("ZADD", "k1", 0, 0)
	cb = redis.NewCmdBuffer()
	cb.Send("HSET", "k1", 0, 0)
	cb.Send("ZINCRBY", "k1", 0, 0)
	v, err = c.DoCmdBuffer(cb, true)
	if err != nil {
		t.Fatalf("EXEC returned error %v", err)
	}

	vs, err := redis.Values(v, nil)
	if err != nil {
		t.Fatalf("Values(v) returned error %v", err)
	}

	if len(vs) != 2 {
		t.Fatalf("len(vs) == %d, want 2", len(vs))
	}

	if _, ok := vs[0].(error); !ok {
		t.Fatalf("first result is type %T, expected error", vs[0])
	}

	if _, ok := vs[1].([]byte); !ok {
		t.Fatalf("second result is type %T, expected []byte", vs[2])
	}

	// Execute commands that fail after EXEC is called. The second command
	// returns an error.

	c.Do("ZADD", "k2", 0, 0)
	cb = redis.NewCmdBuffer()
	cb.Send("ZINCRBY", "k2", 0, 0)
	cb.Send("HSET", "k2", 0, 0)
	v, err = c.DoCmdBuffer(cb, true)
	if err != nil {
		t.Fatalf("EXEC returned error %v", err)
	}

	vs, err = redis.Values(v, nil)
	if err != nil {
		t.Fatalf("Values(v) returned error %v", err)
	}

	if len(vs) != 2 {
		t.Fatalf("len(vs) == %d, want 2", len(vs))
	}

	if _, ok := vs[0].([]byte); !ok {
		t.Fatalf("first result is type %T, expected []byte", vs[0])
	}

	if _, ok := vs[1].(error); !ok {
		t.Fatalf("second result is type %T, expected error", vs[2])
	}
}
