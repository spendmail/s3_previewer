package cache

import (
	"errors"
	"math/rand"
	"strconv"
	"sync"
	"testing"

	internalconfig "github.com/spendmail/s3_previewer/internal/config"
	internallogger "github.com/spendmail/s3_previewer/internal/logger"
	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	t.Run("empty cache", func(t *testing.T) {
		config, err := internalconfig.NewConfig("../../configs/previewer.toml")
		if err != nil {
			t.Fatal(err)
		}

		logger, err := internallogger.New(config)
		if err != nil {
			t.Fatal(err)
		}

		c, err := New(config, logger)
		if err != nil {
			t.Fatal(err)
		}

		_, err = c.Get("aaa")
		require.Truef(t, errors.Is(err, ErrItemNotExists), "actual error %q", err)

		_, err = c.Get("bbb")
		require.Truef(t, errors.Is(err, ErrItemNotExists), "actual error %q", err)
	})

	t.Run("cache capacity", func(t *testing.T) {
		config, err := internalconfig.NewConfig("../../configs/previewer.toml")
		if err != nil {
			t.Fatal(err)
		}

		config.Cache.Capacity = 2

		logger, err := internallogger.New(config)
		if err != nil {
			t.Fatal(err)
		}

		c, err := New(config, logger)
		if err != nil {
			t.Fatal(err)
		}

		err = c.Set("aaa", []byte("aaa"))
		require.NoError(t, err)

		err = c.Set("bbb", []byte("bbb"))
		require.NoError(t, err)

		err = c.Set("ccc", []byte("ccc"))
		require.NoError(t, err)

		_, err = c.Get("aaa")
		require.Truef(t, errors.Is(err, ErrItemNotExists), "actual error %q", err)

		val, err := c.Get("bbb")
		require.NoError(t, err)
		require.Equal(t, []byte("bbb"), val)

		val, err = c.Get("ccc")
		require.NoError(t, err)
		require.Equal(t, []byte("ccc"), val)
	})
}

func TestCacheMultithreading(t *testing.T) {
	config, err := internalconfig.NewConfig("../../configs/previewer.toml")
	if err != nil {
		t.Fatal(err)
	}

	config.Cache.Capacity = 2

	logger, err := internallogger.New(config)
	if err != nil {
		t.Fatal(err)
	}

	c, err := New(config, logger)
	if err != nil {
		t.Fatal(err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000; i++ {
			key := strconv.Itoa(i)
			_ = c.Set(key, []byte(key))
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000; i++ {
			key := strconv.Itoa(rand.Intn(1_000))
			_, _ = c.Get(key)
		}
	}()

	wg.Wait()
}
