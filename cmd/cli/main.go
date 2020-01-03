package main

func main() {
	//	fmt.Fprintf(os.Stderr, "start\n")
	//	var ch chan struct{} = nil
	//	ch2 := make(chan chan struct{})
	//	quit := make(chan struct{})
	//	go func() {
	//		time.Sleep(7 * time.Second)
	//		quit <- struct{}{}
	//	}()
	//	go func() {
	//		time.Sleep(5 * time.Second)
	//		ch2 <- make(chan struct{}, 5)
	//	}()
	//outer:
	//	for {
	//		select {
	//		case ch <- struct{}{}:
	//			fmt.Println("Sent")
	//		case ch3 := <-ch2:
	//			ch = ch3
	//			fmt.Println("Got channel")
	//		case <-quit:
	//			break outer
	//		default:
	//			fmt.Println("default")
	//		}
	//		time.Sleep(time.Second)
	//	}
	//
	//	fmt.Fprintf(os.Stderr, "end\n")
}
