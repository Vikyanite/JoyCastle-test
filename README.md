## 设计简述

1. **百万级别**的实时排行榜用**Redis**就能完全撑住，所以主要采用Redis实现，然后双写MySQL兜底
   1. 若达到**千万级别**可以考虑按照**分数区间**进行拆分，如果出现区间不均可以在小区间继续拆分
2. 高可用与一致性可以通过Redis自带的**集群模式**去保证，当Redis挂掉以后会对服务进行限流，主要的功能就会用MySQL去实现；当Redis重启的时候会通过MySQL中的数据进行恢复

## 逻辑图

           ┌────────────┐
           │  用户打榜   │
           └────┬───────┘
                │
         ┌──────▼──────┐
         │ SetScore	   │
         └──────┬──────┘
                │ Redis存活？
     ┌──────────┴──────────┐
     │                     │
     ▼                     ▼
     Redis写入             MySQL降级写入 + 限流判断
         │                     │
         │               ➜ 记录到 Kafka/内存队列
         ▼                     │
        正常                  等待Redis恢复
                              ➜ 自动补偿Replay
