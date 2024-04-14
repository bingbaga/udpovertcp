

package main

import (

	
)

func main() {

b:=make([]byte,7000)
c:=b

go func ()  {
	for i := 0; i < 1000; i++ {
		for i := 0; i < 7000; i++ {
			b[i]=1
		}
	}
	
}()
for i := 0; i < 1000; i++ {
	for i := 0; i < 7000; i++ {
		c[i]=1
	}
}
}