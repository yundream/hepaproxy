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

Fail Node Num  | 4 thread      | 16 thread  | 16 thread, Redis
---------------|---------------|------------|------------------
0              |  57120.73     | 70927.46   |  68879.34
2              |  56854.97     | 71277.05   |  67469.99
4              |  56654.72     | 70716.54   |  65758.45
8              |  56639.01     | 70444.11   |  64417.19
16             |  56130.34     | 70480.09   |  61179.93
32             |  56004.10     | 70711.82   |  57447.92
64             |  55758.04     | 69732.67   |  52779.92
128            |  54098.67     | 68854.66   |  45125.89
256            |  51985.98     | 66070.08   |  34267.80
512            |  47807.93     | 61551.06   |  24868.62
 1. 1024개의 노드가 있을 때, 최대 Fail Node를 512 개까지 늘렸다.
 1. FailNode가 늘어날 경우, Hash 연산이 추가되므로 이에 따라서 처리량이 감소한다.   
 1. 512개의 노드가 실패할 경우, 평균 2번(평균 2번이 맞는지 확인)의 Hash 연산이 이루어진다. Hash연산에 들어가는 CPU 연산은 네트워크 연산에 비해서 매우 작기 때문에, 요청이 크게 감소하지 않는다. 16 thread 기준으로 대략 14% 정도 감소한다. 
 1. Node fail이 발생했을 때, 실패한 노드로 향하는 요청에 대한 connection table을 유지해서 라우팅하는 방식의 경우, 큰폭의 성능 하락이 있다. connection table을 조회하는데 걸리는 네트워크 시간 때문이다. 

![](/result/request_failover_1024_16c.png)
* 테스트는 HTTP 서버, 클라이언트 방식으로 진행했다.  
* HTTP Request, 네트워크 연산에 비해서 CPU 연산이 무시 할 만큼 작기 때문에, Fail Node가 증가해서 연산이 늘어남에도 불구하고 Request/Sec는 차이가 없었다.
* **표준편차를 넣어야 할 것 같다.**

## Fail node test-2. node를 64개로 제한
제한한 이유는 다음과 같다.
 1. 현실적으로 1024개 정도의 노드를 하나의 클러스터로 구성하지는 않을 것이다.
 1. 좀 더 현실적인 환경을 위해서 Node를 64개로 제한했다.
 1. 노드의 갯수가 줄어든 상태에서 실패가 발생하는 경우, connection table를 참고하는 모델은 분모가 작기 때문에 성능하락이 좀 더 명확히 눈에 보일 것이고, Hepa 모델의 장점은 좀더 눈에 띌 것이다. 

Fail Node Num  | 16 thread  | 16 thread, Redis
---------------|------------|------------------
0              | 71280.74   |  69527.45
4              | 71059.19   |  69379.63
8              | 70631.93   |  53117.25
12             | 68611.25   |  45551.25
16             | 67119.63   |  38893.48
20             | 66178.63   |  34417.78
24             | 65151.12   |  31435.78
28             | 63937.12   |  28746.21
32             | 62988.31   |  26623.65

![](/result/request_failover_64.png)

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
