# 协作式go程

### 为什么要协作式go程

考虑如下开发框架，一组网络接收goroutine接收网络包，解包，然后将逻辑包推送到消息队列，由一个单一的逻辑处理goroutine负责从队列中提取逻辑包并处理(这样主处理逻辑中基本上不用考虑多线程竞争的锁问题了)。

如果逻辑包的处理涉及到调用可能会阻塞的函数调用怎么办，如果在处理函数中直接调用这样的函数将导致逻辑处理goroutine被阻塞，无法继续处理队列中被排队的数据包，这将严重降低服务的处理能力。

一种方式启动一个新的go程去执行阻塞调用，并注册回调函数，当阻塞调用返回后将回调闭包重新push到消息对列中，由逻辑处理goroutine继续处理后续逻辑。但我本人不大喜欢在逻辑处理上使用回调的方式。我希望可以线性的编写逻辑代码。

为了实现这个目的，我需要一个类似lua的单线程协作式coroutine调度机制，单线程让使用者不用担心数据竞争,协作式可以让coroutine在执行异步调用前将执行权交出去，等异步结果返回后再将执行权切换回来，线性的执行后续代码。

但是，goroutine天生就是多线程调度执行的，有办法实现这个目标吗？答案是肯定的。

我们可以实现一个逻辑上的单线程，从全局上看，只有唯一一个goroutine可以执行逻辑处理代码。核心思想就是由调度器从任务队列中提取任务，挑选一个空闲的goroutine,将其唤醒并让自己阻塞，当goroutine需要阻塞时就唤醒调度器并将自己阻塞。这样全局上就只有唯一的goroutine在执行逻辑代码。



下面是一个使用示例：

~~~go
package main


import (
	"fmt"
	"time"
	"coop-go"
)


func main() {

	count := int32(0)
	
	var p *coop.CoopScheduler

	p = coop.NewCoopScheduler(func (e interface{}){
		count++
		if count >= 30000000 {
			p.Close()
			return
		}

      	//调用阻塞函数
		p.BlockCall(func () {
			time.Sleep(time.Millisecond * time.Duration(10))
		})
		//继续投递任务
		p.PostEvent(1)
	})

	for i := 0; i < 10000; i++ {
	  //投递任务	
      p.PostEvent(1)
	}


	p.Start()

	fmt.Printf("scheduler stop,total taskCount:%d\n",c2)


}
~~~



首先用一个任务处理函数作为参数创建调度器。然后向调度器投递任务触发处理循环，最后启动处理。

这里唯一需要关注的是BlockCall,它的参数是一个函数闭包，在里面可以调用会阻塞的函数，BlockCall在内部会将调度器唤醒，但不阻塞自身。也就是说在这个点上，调度器和当前goroutine是并行执行的，因此传递给BallCall的闭包中不允许执行任何线程不安全代码，应该只是一个简单的阻塞调用,例如一个同步获取redis数据的调用如下所示:

~~~go
ret

BlockCall(func() {
   /*
   *这里面不能有任何线程不安全的代码,只是一个简单的函数调用
   */  
	ret = redis.get()
})

if ret {
  //根据返回值执行处理逻辑
}
~~~



### 性能

在我的 i5 双核 2.5GHz mac mini上每秒钟可以执行100W次的调度，虽然跟C协程数千万的调度次数没法比，但是也基本够用了，毕竟在实现的使用中，每秒能处理10W的请求已经相当不错了。







