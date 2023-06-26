package quotes

import (
	"fmt"
	"math/rand"
	"time"
)

var quoteList = []string{
	"Life without worry. You seek Hakuna Matata",
	"Everybody is somebody, even a nobody.",
	"Remember, the journey of a thousand miles begins with the first step.",
	"Ah. Change is good!",
	"Any story worth telling is worth telling twice!",
	"You're a baboon, and I'm not",
	"The Roar is a very powerful gift. It can be used for great good. but it can also lead to terrible evil.",
}

func PrintQuote() {
	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(quoteList))

	quote := quoteList[randomIndex]

	fmt.Printf("\n%s\n- Rafiki\n", quote)
}
