# Failover 알고리즘 변경
* Key+1.. 예상과 달리 분산이 잘 되지 않는다.
* 별도의 랜덤 공간을 만들어서, 여기에서 Backup 노드를 선택했다. 
```golang
func (h Hash) GetBackUpNode(key uint64) int32 {
	r := rand.New(rand.NewSource(int64(key)))
	pivot := r.Int63n(20481)
	for {
		n := jump.Hash(key, h.nodeSize)
		n = n + h.offset
		if h.failNode[int(n)] == 0 {
			return n
		}
		key = uint64(pivot)
		pivot = r.Int63n(20481)
	}
}
```

# Fail Node Test
1024개의 노드에 대한 Failover 성능 테스트 수행 결과. Fail node를 0, 2, 4, 8, 16, 32, 64, 128, 256, 512 로 늘려가면서 테스트 했다. Fail node가 늘어날 수록 백업노드를 찾는데, 더 많은 시간이 걸릴 것이므로 초당 처리 갯수가 줄어들 것으로 예상 할 수 있다.  

Fail Node Num  | Request/Sec
---------------|-------------
0              | 42509.43
2              | 42161.48
4              | 42613.03
8              | 42301.11
16             | 42891.38
32             | 42303.38
64             | 41893.75
128            | 42186.87
256            | 41680.17
512            | 41901.22

![](/result/request_failover_1024_16c.png)
* 테스트는 HTTP 서버, 클라이언트 방식으로 진행했다.  
* HTTP Request, 네트워크 연산에 비해서 CPU 연산이 무시 할 만큼 작기 때문에, Fail Node가 증가해서 연산이 늘어남에도 불구하고 Request/Sec는 차이가 없었다.
* **표준편차를 넣어야 할 것 같다.**
