package solve

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type B009 struct{}

func (o *B009) Solve() {
	s := bufio.NewScanner(os.Stdin)
	s.Scan()
	nm := o.split(s.Text())
	if len(nm) != 2 {
		panic("invalid args")
	}
	n := nm[0]
	m := nm[1]
	if n == 0 || m == 0 {
		fmt.Println(0)
		return
	}

	allBenefit := 0
	for s.Scan() {
		benefits := o.split(s.Text())
		sumBen := o.sum(benefits)
		if sumBen > 0 {
			allBenefit += sumBen
		}
	}
	fmt.Println(allBenefit)
}

func (o *B009) atoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return i
}

func (o *B009) split(s string) []int {
	res := []int{}
	is := strings.Split(s, " ")
	for _, i := range is {
		res = append(res, o.atoi(i))
	}
	return res
}

func (o *B009) sum(is []int) int {
	res := 0
	for _, i := range is {
		res += i
	}
	return res
}
