# Failover 알고리즘 변경
* Key+1.. 예상과 달리 분산이 잘 되지 않는다. **왜 이런 문제가 발생하는가?**
* 별도의 랜덤 공간을 만들어서, 여기에서 Backup 노드를 선택했다. 
```golang
// 원래 계획
func (h Hash) GetNode(key uint64) int32 {
	for {
		n := jump.Hash(key, h.nodeSize)
		n = n + h.offset
		if _, ok := h.failNode[int(n)]; !ok {
			return n
		}
		key++
	}
	return 1
}

// 수정
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

Fail Node Num  | 4 thread      | 16 thread
---------------|---------------|------------
0              | 42509.43      | 53656.41
2              | 42161.48      | 53263.14
4              | 42613.03      | 53273.63
8              | 42301.11      | 53055.43
16             | 42891.38      | 53082.07
32             | 42303.38      | 53369.13
64             | 41893.75      | 53586.34
128            | 42186.87      | 53380.83
256            | 41680.17      | 52649.65
512            | 41901.22      | 53379.21
 1. 1024개의 노드가 있을 때, 최대 512 개의 노드가 실패하는 경우 
 1. FailNode가 늘어나더라도 Request/sec에 영향이 없다. 
 1. 512개의 노드가 실패할 경우, 평균 2번의 Hash 연산이 이루어진다. 각 연산은 마이크로세컨드 내에서 이루어지기 때문에, 성능에 미치는 영향이 거의 없다. 
 1. **각 FailNode에 대해서 CPU 연산을 측정했는데, 유의미한 증감을 찾을 수 없었습니다.**. 네트워크가 개입하는 실 운영환경에서는 Fail Node에 따른 성능감소는 거의 없다고 할 수 있겠습니다.

![](/result/request_failover_1024_16c.png)
* 테스트는 HTTP 서버, 클라이언트 방식으로 진행했다.  
* HTTP Request, 네트워크 연산에 비해서 CPU 연산이 무시 할 만큼 작기 때문에, Fail Node가 증가해서 연산이 늘어남에도 불구하고 Request/Sec는 차이가 없었다.
* **표준편차를 넣어야 할 것 같다.**

## Connection Table을 유지하는 모델
![](https://docs.google.com/drawings/d/1zn5uTmy2_MUP2UF5hSoq8krKPDkIkppTh0bUHgGSHzw/pub?w=780&h=572)

노드가 실패 할 경우, 실패한 노드로 향하는 어뎁터 연결은 RoundRobin 방식으로 다른 노드에 할당한다. 이때, Connection table에 **Key -> node**에 기록한다. 이 후 Fail 노드로 향하는 메시지는 Connection table을 조회해서 메시지를 전송한다. 아래의 문제를 예상 할 수 있다. 
 1. 메시지 경로 설정에 대한 성능은 '''네트워크'''에 좌우된다.  
 1. Connection Table을 유지 할 경우, 네트워크를 경유해야 한다.
 1. 중앙에 집중된 Connection Table은 Fail point다. 

테스트 방법
 1. Connection table은 Redis로 유지한다.  
 1. 미래 1,000,000개의 테이블을 등록한다. 
 1. 노드가 실패하면 1,000,000개의 테이블에서 key로 찾는다.
## 추가작업 
1. 일반방식 추가 : connection table 유지 
1. 나머지 3개는 : connection table 유지 할 필요 없음. : CPU Usage & 표준편차
   n/2 > random > k++ 
