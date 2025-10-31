package main

import (
	"fmt"
	"net"
	"sync"

	"github.com/apernet/hysteria/core/v2/client"
)

type RefClient struct {
	client client.Client
	refCnt int
}
type HYHandle struct {
	conn net.Conn
	addr string
}

type HYPool struct {
	clients map[string]*RefClient
	mu      sync.Mutex

	hds   map[int64]*HYHandle
	hdCnt int64
}

func (pool *HYPool) TCP(addr string, configFunc func() (*client.Config, error), dstAddr string) (int64, error) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	refc, err := pool.GetClient(addr, configFunc)
	if err != nil {
		return 0, err
	}

	rConn, err := refc.client.TCP(dstAddr)
	fmt.Println("TCP ESTALiSH", addr, dstAddr, err)
	if err != nil {
		return 0, err
	}
	pool.hdCnt++
	pool.hds[pool.hdCnt] = &HYHandle{conn: rConn, addr: addr}
	return pool.hdCnt, nil
}

func (pool *HYPool) Close(hd int64) error {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	if handler, ok := pool.hds[hd]; ok {
		defer func(pool *HYPool, addr string) {
			err := pool.ReleaseClient(addr)
			if err != nil {
				fmt.Println("Release error", handler.addr, err)
			}
		}(pool, handler.addr)
		delete(pool.hds, hd)
		err := handler.conn.Close()
		if err != nil {
			fmt.Println("Close error", handler.addr, err)
			return err
		}
		return nil
	}
	fmt.Println("Should Not here", hd)
	return fmt.Errorf("hd not found! %d", hd)
}

func (pool *HYPool) GetClient(addr string, configFunc func() (*client.Config, error)) (*RefClient, error) {

	if refc, ok := pool.clients[addr]; ok {
		refc.refCnt++
		fmt.Println("OldClient GET", addr)
		return refc, nil // 复用已有底层连接
	}

	// 否则新建一个底层连接
	return pool.NewClient(addr, configFunc)
}

func (pool *HYPool) NewClient(addr string, configFunc func() (*client.Config, error)) (*RefClient, error) {
	hyc, err := client.NewReconnectableClient(configFunc,
		func(c client.Client, info *client.HandshakeInfo, count int) {
			connectLog(info, count)
		}, false)

	if err != nil {
		return nil, err
	}
	fmt.Println("NewClient OK", addr)

	refc := &RefClient{client: hyc, refCnt: 1}

	pool.clients[addr] = refc
	return refc, err
}

func (pool *HYPool) ReleaseClient(addr string) error {
	if refc, ok := pool.clients[addr]; ok {
		refc.refCnt--

		if refc.refCnt <= 0 {
			delete(pool.clients, addr)
			fmt.Println("DELClient OK", addr)
			err := refc.client.Close()
			fmt.Println("CloseClient result", addr, err)
			if err != nil {
				return err
			}
		}
	}
	fmt.Println("ReleaseClient OK", addr)
	return nil
}
